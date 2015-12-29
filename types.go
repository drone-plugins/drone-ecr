package main

import (
	"github.com/drone/drone-go/drone"
)

type Save struct {
	// Absolute or relative path
	File string `json:"destination"`
	// Only save specified tags (optional)
	Tags drone.StringSlice `json:"tag"`
}

type ECR struct {
	AccessKey string            `json:"access_key"`
	SecretKey string            `json:"secret_key"`
	Region    string            `json:"region"`
	Storage   string            `json:"storage_driver"`
	Mirror    string            `json:"mirror"`
	Repo      string            `json:"repo"`
	ForceTag  bool              `json:"force_tag"`
	Tag       drone.StringSlice `json:"tag"`
	File      string            `json:"file"`
	Context   string            `json:"context"`
	Bip       string            `json:"bip"`
	Dns       []string          `json:"dns"`
	Load      string            `json:"load"`
	Save      Save              `json:"save"`
	BuildArgs []string          `json:"build_args"`
}
