# drone-ecr

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-ecr/status.svg)](http://beta.drone.io/drone-plugins/drone-ecr)
[![Coverage Status](https://aircover.co/badges/drone-plugins/drone-ecr/coverage.svg)](https://aircover.co/drone-plugins/drone-ecr)
[![](https://badge.imagelayers.io/plugins/drone-ecr:latest.svg)](https://imagelayers.io/?images=plugins/drone-ecr:latest 'Get your own badge on imagelayers.io')

Drone plugin to build and publish Docker images to AWS EC2 Container Registry. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t plugins/ecr .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-ecr' not found or does not exist..
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e ECR_ACCESS_KEY=N1DOBESIHFPDZBI2YBGA \
  -e ECR_SECRET_KEY=HdUp4yYnTjeDaYfH2NICMdHg0V5qHdpce1vxAySv \
  -e ECR_REGION=us-east-1 \
  -e ECR_CREATE_REPOSITORY=true \
  -e DRONE_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  --privileged \
  plugins/ecr --dry-run
```
