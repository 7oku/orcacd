package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sync"
	"time"

	el "github.com/cdfmlr/ellipsis"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
)

type OrcaStack struct {
	servicename string
	repoconfig  *OrcaConfigRepo
	puller      *Puller

	id        string
	storepath string

	compose   *[]byte
	dockerprj *types.Project
	ctx       *context.Context
	status    string
}

func NewOrcaStack(name string, r *OrcaConfigRepo, c *OrcaConfig) *OrcaStack {
	initial := []byte("")
	ctx := context.TODO()

	// generate an unique id 12 chars long from repo url
	regex := regexp.MustCompile(`^.*://`)
	plainid := regex.ReplaceAllString(r.Url, "")
	h := sha256.New()
	h.Write([]byte(plainid))
	id := hex.EncodeToString(h.Sum(nil))[:12]

	// storepath
	storepath := c.Workdir + "/" + id
	o := &OrcaStack{
		servicename: name,
		repoconfig:  r,
		puller:      nil,
		compose:     &initial,
		ctx:         &ctx,
		id:          id,
		dockerprj:   &types.Project{},
		storepath:   storepath,
		status:      "🕒 INIT",
	}

	p := NewPuller(o, c)
	o.puller = p

	return o
}

// gets the stack id
func (orcastack *OrcaStack) GetId() string {
	return orcastack.id
}

// updates the docker compose project in our stack object
func (orcastack *OrcaStack) updateDockerProject(c *OrcaConfig) error {
	project, err := generateDockerProject(orcastack, c)
	if err != nil {
		return err
	}
	orcastack.dockerprj = project

	return nil
}

// compose up
func (orcastack *OrcaStack) ComposeUp(svc api.Service) error {
	err := svc.Up(*orcastack.ctx, orcastack.dockerprj, api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// runs a puller instance for the stack
func (s *OrcaStack) RunPuller(dsession api.Service, c *OrcaConfig, wg *sync.WaitGroup) {
	logPuller.Infof("Puller started for %v [ID: %v]", s.repoconfig.Url, s.id)

	for {
		status, err := s.Cycle(dsession, c)
		if err != nil {
			logPuller.Errorf("Exiting because of: %v", err)
			break
		}
		logOrcacd.Printf("===[ SERVICE: %v | REPO: %v | STATUS:%v ]===", s.servicename, el.Centering(s.repoconfig.Url, 50), status)
		time.Sleep(time.Duration(c.Interval) * time.Second)
	}

	defer wg.Done()
}

// main cycle pulling, updating and ensuring service up
func (s *OrcaStack) Cycle(dsession api.Service, c *OrcaConfig) (string, error) {
	// pull files
	logPuller.Debug("Cycle step 'pull'")
	_, err := s.puller.Pull()
	if err != nil {
		s.status = "❌ PULL ERROR"
		return s.status, nil
	}

	// update compose in object
	logCompose.Debug("Cycle step 'update_project'")
	err = s.updateDockerProject(c)
	if err != nil {
		s.status = "❌ COMPOSE ERROR (SYNTAX?)"
		logCompose.Errorf("Compose Error for %v: %v", s.servicename, err)
		return s.status, nil
	}

	// compose up
	logCompose.Debug("Cycle step 'compose up'")
	if c.Autosync == "on" {
		err = s.ComposeUp(dsession)
		if err != nil {
			logCompose.Errorf("Compose Error for %v: %v", s.servicename, err)
			s.status = "❌ START COMPOSE ERROR"
			return s.status, nil
		}
		s.status = "\x1b[32m✔\x1b[0m SYNCED"
		return s.status, nil
	} else {
		if err == nil {
			s.status = "⏸️  AUTOSYNC OFF"
			return s.status, nil
		}
	}

	return s.status, nil
}
