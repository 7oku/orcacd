package main

import (
	"fmt"
	"os"

	"github.com/docker/compose/v2/pkg/api"
)

func initialize(c *OrcaConfig) ([]*OrcaStack, api.Service) {
	// clean old working dir
	os.RemoveAll(c.Workdir)

	// create new working dir
	err := os.Mkdir(c.Workdir, 0777)
	if err != nil {
		logOrcacd.Fatalf("Cannot create temporary working directory: %v", err)
	}

	// initialize a session to the system's docker services
	dsession, serr := createDockerSession()
	if serr != nil {
		logOrcacd.Fatalf("Could not open a session with docker daemon. Is docker.sock configured correctly? Error: %v", err)
	}

	// initialize api endpoint
	api := NewServer(c)
	api.AddRoutes()
	go api.router.Run(":6666")

	// build stacks
	var stacks []*OrcaStack
	var stack *OrcaStack
	for name, repo := range c.Repos {
		r := OrcaConfigRepo{
			User:       repo["user"],
			Secret:     repo["secret"],
			Url:        repo["url"],
			Searchpath: repo["searchpath"],
		}

		stack = NewOrcaStack(name, &r, c)

		// append
		stacks = append(stacks, stack)
	}

	return stacks, dsession
}

func printBanner() {
	fmt.Println(`
	::::::::  :::::::::   ::::::::      :::      ::::::::  :::::::::  
	:+:    :+: :+:    :+: :+:    :+:   :+: :+:   :+:    :+: :+:    :+: 
	+:+    +:+ +:+    +:+ +:+         +:+   +:+  +:+        +:+    +:+ 
	+#+    +:+ +#++:++#:  +#+        +#++:++#++: +#+        +#+    +:+ 
	+#+    +#+ +#+    +#+ +#+        +#+     +#+ +#+        +#+    +#+ 
	#+#    #+# #+#    #+# #+#    #+# #+#     #+# #+#    #+# #+#    #+# 
	 ########  ###    ###  ########  ###     ###  ########  #########  
	`)
}
