package main

import (
	"encoding/json"
	"os"
)

type config struct {
	DefaultRepoType string `json:"default_repo_type"`
	HTTP            struct {
		Address        string `json:"address"`
		Port           int    `json:"port"`
		TLSPort        int    `json:"tls_port"`
		EnableInsecure bool   `json:"enable_insecure"`

		TLS struct {
			Cert string `json:"cert"`
			Key  string `json:"key"`
		} `json:"tls"`
	} `json:"http"`

	Paths map[string][]*path `json:"paths"`
	paths map[string]*path
}

type path struct {
	Import   string `json:"import"`
	Repo     string `json:"repo"`
	RepoType string `json:"repo_type"`
	Redirect string `json:"redirect"`
	FullPath string `json:"-"`
}

func parseConfigFile(filepath string) (*config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var c config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return nil, err
	}

	newPaths := make(map[string]*path)

	for prefix, p := range c.Paths {
		for _, apath := range p {
			fullpath := prefix + "/" + apath.Import
			newPaths[fullpath] = apath
			apath.FullPath = fullpath
			if apath.Redirect == "" {
				apath.Redirect = apath.Repo
			}
		}
	}

	c.paths = newPaths
	c.Paths = nil

	return &c, nil
}
