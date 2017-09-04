#!/bin/sh

set -ex

# support PLUGIN_ and ECR_ variables
[ -n "$ECR_REGION" ] && export PLUGIN_REGION=${ECR_REGION}
[ -n "$ECR_ACCESS_KEY" ] && export PLUGIN_ACCESS_KEY=${ECR_ACCESS_KEY}
[ -n "$ECR_SECRET_KEY" ] && export PLUGIN_SECRET_KEY=${ECR_SECRET_KEY}
[ -n "$ECR_SESSION_TOKEN" ] && export PLUGIN_SESSION_TOKEN=${ECR_SESSION_TOKEN}
[ -n "$ECR_CREATE_REPOSITORY" ] && export PLUGIN_SECRET_KEY=${PLUGIN_CREATE_REPOSITORY}

# set the region
export AWS_DEFAULT_REGION=${PLUGIN_REGION:-'us-east-1'}

if [ -n "$PLUGIN_ACCESS_KEY" ] && [ -n "$PLUGIN_SECRET_KEY" ]; then
  export AWS_ACCESS_KEY_ID=${PLUGIN_ACCESS_KEY}
  export AWS_SECRET_ACCESS_KEY=${PLUGIN_SECRET_KEY}
fi

if [ -n "$PLUGIN_SESSION_TOKEN" ]; then
  export AWS_ACCESS_SESSION_TOKEN=${PLUGIN_SESSION_TOKEN}
fi

# Support external AWS_ variables and source it.
# Sample file of PLUGIN_TOKEN_FILE
# export AWS_ACCESS_KEY_ID=XXX
# export AWS_SECRET_ACCESS_KEY=XXX
# export AWS_SESSION_TOKEN=xxx
if [ "$PLUGIN_EXTERNAL" == "true" ] && [ -n "PLUGIN_TOKEN_FILE" ]; then
  . ${PLUGIN_TOKEN_FILE}
fi

# Authenticate your Docker client to ECR registry
aws ecr get-login --no-include-email --region ap-southeast-2|sh

# invoke the docker plugin
/bin/drone-docker "$@"
