# Docker image for the ecr plugin
#
#     docker build --rm=true -t plugins/drone-ecr .

FROM rancher/docker:v1.10.2

ADD drone-ecr /go/bin/
VOLUME /var/lib/docker
ENTRYPOINT ["/usr/bin/dockerlaunch", "/go/bin/drone-ecr"]
