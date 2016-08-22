package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type (
	// Daemon defines Docker daemon parameters.
	Daemon struct {
		Mirror        string   // Docker registry mirror
		StorageDriver string   // Docker daemon storage driver
		StoragePath   string   // Docker daemon storage path
		Disabled      bool     // DOcker daemon is disabled (already running)
		Debug         bool     // Docker daemon started in debug mode
		Bip           string   // Docker daemon network bridge IP address
		DNS           []string // Docker daemon dns server
		MTU           string   // Docker daemon mtu setting
	}

	// Login defines Docker login parameters.
	Login struct {
		Region    string // AWS Region
		AccessKey string // AWS AccessKey
		SecretKey string // AWS SecretKey
	}

	ecrLogin struct {
		registry string // AWS ECR Docker registry address
		username string // AWS ECR UserName
		password string // AWS ECR Password
	}

	// Build defines Docker build parameters.
	Build struct {
		Name             string   // Docker build using default named tag
		Dockerfile       string   // Docker build Dockerfile
		Context          string   // Docker build context
		Tags             []string // Docker build tags
		Args             []string // Docker build args
		Repo             string   // Docker build repository
		CreateRepository bool     //Create repository if provided one does not exists
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login  Login  // Docker login configuration
		Build  Build  // Docker build configuration
		Daemon Daemon // Docker daemon configuration
		Dryrun bool   // Docker push is skipped
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {

	// start the Docker daemon server
	if !p.Daemon.Disabled {
		cmd := commandDaemon(p.Daemon)
		if p.Daemon.Debug {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
		}
		go func() {
			trace(cmd)
			cmd.Run()
		}()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < 15; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	session := ecrNewSession(p.Login)
	// login into aws and get temp username and password
	el, err := ecrDoLogin(session, p.Login)
	if err != nil {
		return fmt.Errorf("error authenticating: %s", err)
	}
	shortRepoName := p.Build.Repo
	// in case someone uses the shorthand repository name
	// with a custom registry, we should concatinate so that
	// we have the fully qualified image name.
	if strings.Count(p.Build.Repo, "/") <= 1 && len(el.registry) != 0 && !strings.HasPrefix(p.Build.Repo, el.registry) {
		shortRepoName = p.Build.Repo //take short name
		p.Build.Repo = fmt.Sprintf("%s/%s", el.registry, p.Build.Repo)
	}

	// login to the Docker registry
	cmd := commandLogin(el)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error authenticating: %s", err)
	}

	// creates repository on AWS ECR
	if p.Build.CreateRepository {
		if err := ecrCreateRepository(session, shortRepoName); err != nil {
			return err
		}
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())      // docker version
	cmds = append(cmds, commandInfo())         // docker info
	cmds = append(cmds, commandBuild(p.Build)) // docker build

	for _, tag := range p.Build.Tags {
		cmds = append(cmds, commandTag(p.Build, tag)) // docker tag

		if p.Dryrun == false {
			cmds = append(cmds, commandPush(p.Build, tag)) // docker push
		}
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

const dockerExe = "/usr/local/bin/docker"

// helper function to create the docker login command.
func commandLogin(login ecrLogin) *exec.Cmd {
	return exec.Command(
		dockerExe, "login",
		"-u", login.username,
		"-p", login.password,
		login.registry,
	)
}

// helper function to create the docker info command.
func commandVersion() *exec.Cmd {
	return exec.Command(dockerExe, "version")
}

// helper function to create the docker info command.
func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}

// helper function to create the docker build command.
func commandBuild(build Build) *exec.Cmd {
	cmd := exec.Command(
		dockerExe, "build",
		"--pull=true",
		"--rm=true",
		"-f", build.Dockerfile,
		"-t", build.Name,
	)
	for _, arg := range build.Args {
		cmd.Args = append(cmd.Args, "--build-arg", arg)
	}
	cmd.Args = append(cmd.Args, build.Context)
	return cmd
}

// helper function to create the docker tag command.
func commandTag(build Build, tag string) *exec.Cmd {
	var (
		source = build.Name
		target = fmt.Sprintf("%s:%s", build.Repo, tag)
	)
	return exec.Command(
		dockerExe, "tag", source, target,
	)
}

// helper function to create the docker push command.
func commandPush(build Build, tag string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", build.Repo, tag)
	return exec.Command(dockerExe, "push", target)
}

// helper function to create the docker daemon command.
func commandDaemon(daemon Daemon) *exec.Cmd {
	args := []string{"daemon", "-g", daemon.StoragePath}

	if daemon.StorageDriver != "" {
		args = append(args, "-s", daemon.StorageDriver)
	}
	if len(daemon.Mirror) != 0 {
		args = append(args, "--registry-mirror", daemon.Mirror)
	}
	if len(daemon.Bip) != 0 {
		args = append(args, "--bip", daemon.Bip)
	}
	for _, dns := range daemon.DNS {
		args = append(args, "--dns", dns)
	}
	if len(daemon.MTU) != 0 {
		args = append(args, "--mtu", daemon.MTU)
	}
	return exec.Command(dockerExe, args...)
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}

func ecrNewSession(login Login) *ecr.ECR {
	return ecr.New(session.New(&aws.Config{
		Region:      aws.String(login.Region),
		Credentials: credentials.NewStaticCredentials(login.AccessKey, login.SecretKey, ""),
	}))
}

func ecrDoLogin(svc *ecr.ECR, login Login) (ecrLogin, error) {
	resp, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return ecrLogin{}, fmt.Errorf("%s, unable to retrieve registry credentials from aws", err)
	}

	if len(resp.AuthorizationData) < 1 {
		return ecrLogin{}, fmt.Errorf("request did not return authorization data")
	}

	bytes, err := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return ecrLogin{}, fmt.Errorf("error decoding authorization token: %s", err)
	}
	token := string(bytes[:len(bytes)])

	authTokens := strings.Split(token, ":")
	if len(authTokens) != 2 {
		return ecrLogin{}, fmt.Errorf("authorization token does not contain data in <user>:<password> format: %s", token)
	}

	registryURL, err := url.Parse(*resp.AuthorizationData[0].ProxyEndpoint)
	if err != nil {
		return ecrLogin{}, fmt.Errorf("error parsing registry URL: %s", err)
	}
	return ecrLogin{
		username: authTokens[0],
		password: authTokens[1],
		registry: registryURL.Host,
	}, nil
}

func ecrCreateRepository(session *ecr.ECR, repository string) error {
	ri := &ecr.CreateRepositoryInput{
		RepositoryName: &repository,
	}
	_, err := session.CreateRepository(ri)
	if err != nil && !strings.HasPrefix(err.Error(), "RepositoryAlreadyExistsException") {
		return fmt.Errorf("error creating repository: %s", err.Error())
	}
	return nil
}
