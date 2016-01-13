# drone-ecr
Drone plugin for publishing Docker images to the EC2 Container Registry


## Docker

Build the Docker container:

```sh
docker build --rm=true -t plugins/drone-ecr .
```

Build and Publish a Docker container

```sh
docker run -i --privileged -v $(pwd):/drone/src plugins/drone-ecr <<EOF
{
	"workspace": {
		"path": "/drone/src"
	},
	"build" : {
		"number": 1,
		"head_commit": {
			"sha": "9f2849d5",
			"branch": "master",
			"ref": "refs/heads/master"
		}
	},
	"vargs": {
		"access_key": "MyAWSAccessKey",
		"secret_key": "MyAWSSecretKey",
		"region": "us-east-1",
		"repo": "foo/bar",
		"storage_driver": "aufs"
	}
}
EOF
```
