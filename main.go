package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

func main() {
	workspace := drone.Workspace{}
	build := drone.Build{}
	vargs := ECR{}

	plugin.Param("workspace", &workspace)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	//Perform ECR credential lookup and parse out username, password, registry
	if vargs.AccessKey == "" {
		fmt.Println("Please provide an access key id")
		os.Exit(1)
	}

	if vargs.SecretKey == "" {
		fmt.Println("Please provide a secret access key")
		os.Exit(1)
	}

	if vargs.Region == "" {
		fmt.Println("Please provide a region")
		os.Exit(1)
	}
	svc := ecr.New(session.New(&aws.Config{
		Region:      aws.String(vargs.Region),
		Credentials: credentials.NewStaticCredentials(vargs.AccessKey, vargs.SecretKey, ""),
	}))

	resp, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		fmt.Println("Unable to retrieve Registry credentials from AWS")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(resp.AuthorizationData) < 1 {
		fmt.Println("Request did not return authorization data")
		os.Exit(1)
	}

	bytes, err := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		fmt.Printf("Error decoding authorization token: %s", err)
		os.Exit(1)
	}
	token := string(bytes[:len(bytes)])

	authTokens := strings.Split(token, ":")
	if len(authTokens) != 2 {
		fmt.Printf("Authorization token does not contain data in <user>:<password> format: %s", token)
		os.Exit(1)
	}

	registryURL, err := url.Parse(*resp.AuthorizationData[0].ProxyEndpoint)
	if err != nil {
		fmt.Printf("Error parsing registry URL: %s", err)
		os.Exit(1)
	}

	username := authTokens[0]
	password := authTokens[1]
	registry := registryURL.Host
	// in case someone uses the shorthand repository name
	// with a custom registry, we should concatinate so that
	// we have the fully qualified image name.
	if strings.Count(vargs.Repo, "/") <= 1 && len(registry) != 0 && !strings.HasPrefix(vargs.Repo, registry) {
		vargs.Repo = fmt.Sprintf("%s/%s", registry, vargs.Repo)
	}

	// Set the Dockerfile name
	if len(vargs.File) == 0 {
		vargs.File = "Dockerfile"
	}
	// Set the Context value
	if len(vargs.Context) == 0 {
		vargs.Context = "."
	}
	// Set the Tag value
	if vargs.Tag.Len() == 0 {
		vargs.Tag.UnmarshalJSON([]byte("[\"latest\"]"))
	}
	// Get absolute path for 'save' file
	if len(vargs.Save.File) != 0 {
		if !filepath.IsAbs(vargs.Save.File) {
			vargs.Save.File = filepath.Join(workspace.Path, vargs.Save.File)
		}
	}
	// Get absolute path for 'load' file
	if len(vargs.Load) != 0 {
		if !filepath.IsAbs(vargs.Load) {
			vargs.Load = filepath.Join(workspace.Path, vargs.Load)
		}
	}

	go func() {
		args := []string{"-d"}

		if len(vargs.Storage) != 0 {
			args = append(args, "-s", vargs.Storage)
		}

		if len(vargs.Mirror) != 0 {
			args = append(args, "--registry-mirror", vargs.Mirror)
		}
		if len(vargs.Bip) != 0 {
			args = append(args, "--bip", vargs.Bip)
		}

		for _, value := range vargs.Dns {
			args = append(args, "--dns", value)
		}

		cmd := exec.Command("/usr/bin/docker", args...)
		if os.Getenv("DOCKER_LAUNCH_DEBUG") == "true" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
		}
		trace(cmd)
		cmd.Run()
	}()

	// ping Docker until available
	for i := 0; i < 3; i++ {
		cmd := exec.Command("/usr/bin/docker", "info")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	// Login to Docker

	cmd := exec.Command("/usr/bin/docker", "login", "-u", username, "-p", password, "-e", "none", registry)
	cmd.Dir = workspace.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Login failed.")
		os.Exit(1)
	}
	// Docker environment info
	cmd = exec.Command("/usr/bin/docker", "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()
	cmd = exec.Command("/usr/bin/docker", "info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()

	// Restore from tarred image repository
	if len(vargs.Load) != 0 {
		if _, err := os.Stat(vargs.Load); err != nil {
			fmt.Printf("Archive %s does not exist. Building from scratch.\n", vargs.Load)
		} else {
			cmd := exec.Command("/usr/bin/docker", "load", "-i", vargs.Load)
			cmd.Dir = workspace.Path
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			trace(cmd)
			err := cmd.Run()
			if err != nil {
				os.Exit(1)
			}
		}
	}

	// Build the container
	name := fmt.Sprintf("%s:%s", vargs.Repo, vargs.Tag.Slice()[0])
	cmd = exec.Command("/usr/bin/docker", "build", "--pull=true", "--rm=true", "-f", vargs.File, "-t", name)
	for _, value := range vargs.BuildArgs {
		cmd.Args = append(cmd.Args, "--build-arg", value)
	}
	cmd.Args = append(cmd.Args, vargs.Context)
	cmd.Dir = workspace.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}

	// Creates image tags
	for _, tag := range vargs.Tag.Slice()[1:] {
		name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
		cmd = exec.Command("/usr/bin/docker", "tag")
		if vargs.ForceTag {
			cmd.Args = append(cmd.Args, "--force=true")
		}
		cmd.Args = append(cmd.Args, name, name_)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err = cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

	// Push the image and tags to the registry
	for _, tag := range vargs.Tag.Slice() {
		name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
		cmd = exec.Command("/usr/bin/docker", "push", name_)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err = cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

	// Save to tarred image repository
	if len(vargs.Save.File) != 0 {
		// if the destination directory does not exist, create it
		dir := filepath.Dir(vargs.Save.File)
		os.MkdirAll(dir, 0755)

		cmd = exec.Command("/usr/bin/docker", "save", "-o", vargs.Save.File)

		// Limit saving to the given tags
		if vargs.Save.Tags.Len() != 0 {
			for _, tag := range vargs.Save.Tags.Slice() {
				name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
				cmd.Args = append(cmd.Args, name_)
			}
		} else {
			cmd.Args = append(cmd.Args, vargs.Repo)
		}

		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err := cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
