package main

import (
	"path"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

// generate docker project
func generateDockerProject(s *OrcaStack, c *OrcaConfig) (*types.Project, error) {
	configDetails := types.ConfigDetails{
		WorkingDir: c.Targetpath + "/" + s.servicename,
		ConfigFiles: []types.ConfigFile{
			{Filename: "docker-compose.yaml", Content: *s.compose},
		},
		// we can set shell ENV VARS here. theys can be
		// consumed in docker-compose.yml via ${TESTVAR}
		// Environment: map[string]string{
		// 	"TESTVAR": "testval",
		// },
	}

	projectName := path.Base(s.servicename)

	p, err := loader.LoadWithContext(*s.ctx, configDetails, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
	})
	if err != nil {
		return nil, err
	}

	addServiceLabels(p)
	return p, nil
}

// new docker session
func createDockerSession() (api.Service, error) {
	var srv api.Service
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return srv, err
	}

	dockerContext := "default"

	sessionOpts := &flags.ClientOptions{Context: dockerContext, LogLevel: "error"}
	err = dockerCli.Initialize(sessionOpts)
	if err != nil {
		return srv, err
	}

	srv = compose.NewComposeService(dockerCli)
	return srv, nil
}

// adds some labels to the service
func addServiceLabels(project *types.Project) {
	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  "/",
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		project.Services[i] = s
	}
}
