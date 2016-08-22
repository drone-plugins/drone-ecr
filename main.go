package main

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

// build number set at compile-time
var version string

func main() {
	app := cli.NewApp()
	app.Name = "ecr plugin"
	app.Usage = "ecr plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run disables docker push",
			EnvVar: "PLUGIN_DRY_RUN",
		},

		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
			Value:  "00000000",
		},

		// daemon parameters
		cli.StringFlag{
			Name:   "daemon.mirror",
			Usage:  "docker daemon registry mirror",
			EnvVar: "PLUGIN_MIRROR",
		},
		cli.StringFlag{
			Name:   "daemon.storage-driver",
			Usage:  "docker daemon storage driver",
			EnvVar: "PLUGIN_STORAGE_DRIVER",
		},
		cli.StringFlag{
			Name:   "daemon.storage-path",
			Usage:  "docker daemon storage path",
			Value:  "/var/lib/docker",
			EnvVar: "PLUGIN_STORAGE_PATH",
		},
		cli.StringFlag{
			Name:   "daemon.bip",
			Usage:  "docker daemon bride ip address",
			EnvVar: "PLUGIN_BIP",
		},
		cli.StringFlag{
			Name:   "daemon.mtu",
			Usage:  "docker daemon custom mtu setting",
			EnvVar: "PLUGIN_MTU",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns",
			Usage:  "docker daemon dns server",
			EnvVar: "PLUGIN_DNS",
		},
		cli.BoolFlag{
			Name:   "daemon.debug",
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DEBUG,DOCKER_LAUNCH_DEBUG",
		},
		cli.BoolFlag{
			Name:   "daemon.off",
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DAEMON_OFF",
		},

		// build parameters

		cli.StringFlag{
			Name:   "dockerfile",
			Usage:  "build dockerfile",
			Value:  "Dockerfile",
			EnvVar: "PLUGIN_DOCKERFILE",
		},
		cli.StringFlag{
			Name:   "context",
			Usage:  "build context",
			Value:  ".",
			EnvVar: "PLUGIN_CONTEXT",
		},
		cli.StringSliceFlag{
			Name:   "tags",
			Usage:  "build tags",
			Value:  &cli.StringSlice{"latest"},
			EnvVar: "PLUGIN_TAG,PLUGIN_TAGS",
		},
		cli.StringSliceFlag{
			Name:   "args",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS",
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "docker repository",
			EnvVar: "PLUGIN_REPO",
		},
		cli.BoolFlag{
			Name:   "ecr.create-repository",
			Usage:  "create aws ecr repository if does not exists",
			EnvVar: "ECR_CREATE_REPOSITORY,PLUGIN_CREATE_REPOSITORY",
		},

		// secret variables
		cli.StringFlag{
			Name:   "ecr.access-key",
			Usage:  "aws ecr access key",
			EnvVar: "ECR_ACCESS_KEY,PLUGIN_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "ecr.secret-key",
			Usage:  "aws ecr secret key",
			EnvVar: "ECR_SECRET_KEY,PLUGIN_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "ecr.region",
			Usage:  "aws ecr region",
			EnvVar: "ECR_REGION,PLUGIN_REGION",
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) {
	plugin := Plugin{
		Dryrun: c.Bool("dry-run"),
		Login: Login{
			Region:    c.String("ecr.region"),
			AccessKey: c.String("ecr.access-key"),
			SecretKey: c.String("ecr.secret-key"),
		},
		Build: Build{
			Name:             c.String("commit.sha"),
			Dockerfile:       c.String("dockerfile"),
			Context:          c.String("context"),
			Tags:             c.StringSlice("tags"),
			Args:             c.StringSlice("args"),
			Repo:             c.String("repo"),
			CreateRepository: c.Bool("ecr.create-repository"),
		},
		Daemon: Daemon{
			Mirror:        c.String("daemon.mirror"),
			StorageDriver: c.String("daemon.storage-driver"),
			StoragePath:   c.String("daemon.storage-path"),
			Disabled:      c.Bool("daemon.off"),
			Debug:         c.Bool("daemon.debug"),
			Bip:           c.String("daemon.bip"),
			DNS:           c.StringSlice("daemon.dns"),
			MTU:           c.String("daemon.mtu"),
		},
	}

	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
