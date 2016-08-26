Use the ECR plugin to build and push Docker images to an AWS Elastic Container Registry.

## Config
This plugin is built on top of the [docker plugin](https://github.com/drone-plugins/drone-docker)
so you can use parameters from [docker plugin docs][docker_plugin_docs]

The following parameters are used to configure this plugin:

* **access_key** - authenticates with this key
* **secret_key** - authenticates with this secret
* **region** - uses this region
* **create_repository** - create aws ecr repot if provided repo does not exist (default false)

The following secret values can be set to configure the plugin.

* **ECR_ACCESS_KEY** - corresponds to **access_key**
* **ECR_SECRET_KEY** - corresponds to **secret_key**
* **ECR_REGION** - corresponds to **region**
* **ECR_CREATE_REPOSITORY** - corresponds to **create_repository**

It is highly recommended to put the **ECR_ACCESS_KEY** or **ECR_SECRET_KEY** into
secrets so it is not exposed to users. This can be done using the drone-cli.

```bash
drone secret add --image=plugins/ecr \
    octocat/hello-world ECR_ACCESS_KEY pa55word

drone secret add --image=plugins/ecr \
    octocat/hello-world ECR_SECRET_KEY pa55word
```

Then sign the YAML file after all secrets are added.

```bash
drone sign octocat/hello-world
```

See [secrets](http://readme.drone.io/0.5/usage/secrets/) for additional
information on secrets

The following is a sample Docker with ECR configuration in your .drone.yml file:

```yaml
publish:
  ecr:
    access_key: MyAWSAccessKey
    secret_key: MyAWSSecretKey
    region: us-east-1
    repo: foo/bar
    tag: latest
    file: Dockerfile
```

For more examples look at the [docker plugin docs][docker_plugin_docs]

[docker_plugin_docs]: https://github.com/drone-plugins/drone-docker/blob/HEAD/DOCS.md
