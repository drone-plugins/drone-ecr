# Docker image for the ecr plugin
#
#     docker build --rm=true -t plugins/drone-ecr .

FROM rancher/docker:1.9.1

ADD drone-ecr /go/bin/
VOLUME /var/lib/docker
ENTRYPOINT ["/usr/bin/dockerlaunch", "/go/bin/drone-ecr"]
