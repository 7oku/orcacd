package main

import (
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type OrcaConfig struct {
	Targetpath string                       `koanf:"targetpath"`
	Workdir    string                       `koanf:"workdir"`
	Loglevel   string                       `koanf:"loglevel"`
	Autosync   string                       `koanf:"autosync"`
	Interval   int                          `koanf:"interval"`
	Basicauth  map[string]string            `koanf:"basicauth"`
	Repos      map[string]map[string]string `koanf:"repos"`
}

type OrcaConfigRepo struct {
	Servicename string
	Url         string `koanf:"url"`
	User        string `koanf:"user"`
	Secret      string `koanf:"secret"`
}

func NewConfig() *OrcaConfig {
	var k = koanf.New(".")
	var c OrcaConfig

	// load defaults
	k.Load(confmap.Provider(map[string]interface{}{
		"loglevel":   "info",
		"autosync":   "on",
		"interval":   300,
		"workdir":    "/tmp/ocd",
		"targetpath": "/tmp/ocd/opt/compose",
		"basicauth": map[string]string{
			"ocdadmin": "ocd1337god",
		},
	}, "."), nil)

	// load yml
	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		log.Printf("Could not load or fine a config file. Using defaults and ENV: %v", err)
	}

	// load env
	k.Load(env.Provider("OCD_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "OCD_")), "_", ".", -1)
	}), nil)

	// generate
	err := k.Unmarshal("", &c)
	if err != nil {
		log.Fatalf("Cannot unmarshal configuration. Valid YAML?")
	}

	return &c
}
