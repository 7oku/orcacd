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
	Servicename string
	Repoconfig  *OrcaConfigRepo
	Puller      *Puller

	Id        string
	Storepath string

	Compose   *[]byte
	Dockerprj *types.Project
	Ctx       *context.Context
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
		Servicename: name,
		Repoconfig:  r,
		Puller:      nil,
		Compose:     &initial,
		Ctx:         &ctx,
		Id:          id,
		Dockerprj:   &types.Project{},
		Storepath:   storepath,
	}

	p := NewPuller(o, c)
	o.Puller = p

	return o
}

func (orcastack *OrcaStack) GetId() string {
	return orcastack.Id
}

func (orcastack *OrcaStack) updateDockerProject(c *OrcaConfig) error {
	project, err := generateDockerProject(orcastack, c)
	if err != nil {
		return err
	}
	orcastack.Dockerprj = project

	return nil
}

func (orcastack *OrcaStack) ComposeUp(svc api.Service) error {
	err := svc.Up(*orcastack.Ctx, orcastack.Dockerprj, api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *OrcaStack) RunPuller(dsession api.Service, c *OrcaConfig, wg *sync.WaitGroup) {
	logPuller.Infof("Puller started for %v [ID: %v]", s.Repoconfig.Url, s.Id)

	for {
		status, err := s.Cycle(dsession, c)
		if err != nil {
			logPuller.Errorf("Exiting because of: %v", err)
			break
		}
		logOrcacd.Printf("===[ SERVICE: %v | REPO: %v | STATUS:%v ]===", s.Servicename, el.Centering(s.Repoconfig.Url, 50), status)
		time.Sleep(time.Duration(c.Interval) * time.Second)
	}

	defer wg.Done()
}

func (s *OrcaStack) Cycle(dsession api.Service, c *OrcaConfig) (string, error) {
	var status string

	// pull files
	logPuller.Debug("Cycle step 'pull'")
	_, err := s.Puller.Pull()
	if err != nil {
		status = "❌ PULL ERROR"
		return status, nil
	}

	logCompose.Debug("Cycle step 'update_project'")
	// update compose in object
	err = s.updateDockerProject(c)
	if err != nil {
		status = "❌ COMPOSE ERROR (SYNTAX?)"
		logCompose.Errorf("Compose Error for %v: %v", s.Servicename, err)
		return status, nil
	}

	logCompose.Debug("Cycle step 'compose up'")
	// compose up
	if c.Autosync == "on" {
		err = s.ComposeUp(dsession)
		if err != nil {
			logCompose.Errorf("Compose Error for %v: %v", s.Servicename, err)
			status = "❌ START COMPOSE ERROR"
			return status, nil
		}
		status = "\x1b[32m✔\x1b[0m SYNCED"
		return status, nil
	} else {
		if err == nil {
			status = "⏸️  AUTOSYNC OFF"
			return status, nil
		}
	}

	return status, nil
}
