package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/udhos/equalfile"
)

type Puller struct {
	stack  *OrcaStack
	config *OrcaConfig
}

func NewPuller(orcastack *OrcaStack, config *OrcaConfig) *Puller {
	return &Puller{
		stack:  orcastack,
		config: config,
	}
}

// pull
func (p *Puller) Pull() (bool, error) {
	logPuller.Debugf("Pulling %v from %v", p.stack.Servicename, p.stack.Repoconfig.Url)
	// get file into temporary working dir
	err := p.Get()
	if err != nil {
		logPuller.Errorf("Could not download file from %v: %v", p.stack.Repoconfig.Url, err)
		return false, err
	}
	logPuller.Debugf("Downloaded from %v", p.stack.Repoconfig.Url)

	// prepare targetdir
	err = os.MkdirAll(p.config.Targetpath+"/"+p.stack.Servicename, os.ModePerm)
	if err != nil {
		logPuller.Errorf("Could not access target path at %v. Error: %v", p.config.Targetpath, err)
		return false, err
	}

	// work on files
	files, err := os.ReadDir(p.stack.Storepath)
	if err != nil {
		logPuller.Errorf("Cannot read workdir %v: %v", p.stack.Storepath, err)
		return false, err
	}
	for _, file := range files {
		sourcefile := p.stack.Storepath + "/" + file.Name()
		targetfile := p.config.Targetpath + "/" + p.stack.Servicename + "/" + file.Name()
		logPuller.Debugf("Working on Source(%v) and Target(%v)", sourcefile, targetfile)

		// update stack object with new compose file
		if file.Name() == "docker-compose.yaml" || file.Name() == "docker-compose.yml" {
			err = p.PopulateCompose(sourcefile)
			if err != nil {
				logPuller.Errorf("Could not populate compose from sourcefile %v into the orcastack obj: %v ...", sourcefile, err)
				return false, err
			}
		}

		// if file does not exist in target, copy it over
		logPuller.Debugf("Checking if file %v is equal to target file %v ...", sourcefile, targetfile)
		if _, err := os.Stat(targetfile); err != nil {
			// file does not exist, copy over!
			err := p.Copy(sourcefile, targetfile)
			if err != nil {
				logPuller.Errorf("Could not copy source file %v to target file %v: %v", sourcefile, targetfile, err)
				return false, err
			} else {
				logPuller.Infof("[ Targetfile did not exist, synced ✨! ")
				return true, err
			}
		} else {
			// ... compare it to check if we need update
			cmp := equalfile.New(nil, equalfile.Options{})
			equal, err := cmp.CompareFile(sourcefile, targetfile)
			if err != nil {
				logPuller.Errorf("Error comparing files: %v", err)
			} else {
				if !equal {
					// source file differs from targetfile, copy over!
					err = p.Copy(sourcefile, targetfile)
					if err != nil {
						logPuller.Errorf("Could not copy source file %v to target file %v: %v", sourcefile, targetfile, err)
						return false, err
					} else {
						logPuller.Infof("[ Files differ, synced ✨! ]")
						return true, err
					}
				} else {
					// file is equal, nothing to do
					logPuller.Infof("[ %v has no changes 👌! ]", p.stack.Servicename)
					return false, err
				}
			}
		}
	}
	return false, nil
}

// write contets into object
func (p *Puller) PopulateCompose(sourcefile string) error {
	compose, err := os.ReadFile(sourcefile)
	if err != nil {
		return err
	}
	p.stack.Compose = &compose

	return nil
}

// write contents to targetfile
func (p *Puller) Copy(sourcefile string, targetfile string) error {
	destination, err := os.Create(targetfile)
	if err != nil {
		return err
	}
	defer destination.Close()

	err = os.WriteFile(destination.Name(), *p.stack.Compose, 0777)
	if err != nil {
		return err
	}

	return nil
}

// get file from remote
func (p *Puller) Get() error {
	dstFilename := path.Base(p.stack.Repoconfig.Url)

	// create storepath
	err := os.MkdirAll(p.stack.Storepath, os.ModePerm)
	if err != nil {
		logPuller.Debug(err)
		return err
	}

	// create the request
	req, err := http.NewRequest("GET", p.stack.Repoconfig.Url, nil)
	if err != nil {
		logPuller.Debug(err)
		return err
	}

	// add headers for GitLab and GitHub
	switch true {
	case strings.Contains(p.stack.Repoconfig.Url, "/api/v4/projects/"):
		logPuller.Infof("%v seems to be a Gitlab URL! Applying PRIVATE-TOKEN header ... ", p.stack.Servicename)
		req.Header.Set("PRIVATE-TOKEN", p.stack.Repoconfig.Secret)

		// GitLab uses strange format for downloading files via api:
		// i.e. https://somedomain.com/api/v4/projects/<repoid>/repository/files/<folder1>%2F<folderN>%2Fdocker-compose.yaml/raw?ref=main
		// 										  	      ^^^(1)                ^^^(2)
		// ... therefore we need to extract file name from the part between the
		// last occurance of %2F (1) and the /raw part (2)
		re := regexp.MustCompile(`.*%2F([^%2F]*)\/raw`)
		dstFilename = re.FindStringSubmatch(p.stack.Repoconfig.Url)[1]

	case strings.Contains(p.stack.Repoconfig.Url, "raw.githubusercontent.com"):
		if p.stack.Repoconfig.Secret != "" {
			logPuller.Infof("%v seems to be a GitHub URL! Applying TOKEN header ... ", p.stack.Servicename)
			logPuller.Debugf("Token found: %v", p.stack.Repoconfig.Secret)
			req.Header.Set("Authorization", "token "+p.stack.Repoconfig.Secret)
		} else {
			logPuller.Infof("%v seems to be a GitHub URL, but TOKEN was not given. Downloading in plain mode ... ", p.stack.Servicename)
		}

	default:
		// Just a normal URL, so add authentication if provided
		if p.stack.Repoconfig.User != "" && p.stack.Repoconfig.Secret != "" {
			logPuller.Info("Applying AUTHORIZATION header for %v ...", p.stack.Servicename)
			req.SetBasicAuth(p.stack.Repoconfig.User, p.stack.Repoconfig.Secret)
		}
	}

	// get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logPuller.Debug(err)
		return err
	}
	defer resp.Body.Close()

	// check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// create the file
	out, err := os.Create(p.stack.Storepath + "/" + dstFilename)
	if err != nil {
		logPuller.Debug(err)
		return err
	}
	defer out.Close()

	// write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logPuller.Debug(err)
		return err
	}

	return nil
}
