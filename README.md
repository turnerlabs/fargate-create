fargate-create
==============

A CLI tool for scaffolding out new [AWS ECS/Fargate](https://aws.amazon.com/fargate/) applications based on [terraform-ecs-fargate](https://github.com/turnerlabs/terraform-ecs-fargate) and [Fargate CLI](https://github.com/turnerlabs/fargate).


### Why?

The main design goal of this project is to create an easy and maintainable Fargate experience by separating `infrastructure` related concerns and `application` related concerns using tools that are optimized for each.

### Installation

You can install the CLI with a curl utility script (macos/linux) or by downloading the binary from the releases page. Once installed you'll get the `fargate-create` command.

```
curl -s get-fargate-create.turnerlabs.io | sh
````

### Usage

Assuming you have a project with a [Dockerfile]()...

Specify your template's input parameters in [terraform.tfvars](https://www.terraform.io/docs/configuration/variables.html) (or terraform.json).  The [default web application template's](https://github.com/turnerlabs/terraform-ecs-fargate) input looks something like this.

```hcl
# app/env to scaffold
app = "my-app"
environment = "dev"

internal = "true"
container_port = "8080"
replicas = "1"
health_check = "/health"
aws_profile = "default"
vpc = "vpc-123"
private_subnets = "subnet-123,subnet-456"
public_subnets = "subnet-789,subnet-012"
tags = {
  application = "my-app"
  environment = "dev"
}
```

```shell
$ fargate-create
scaffolding my-app dev
Looking up AWS Account ID using profile: default
downloading terraform template https://github.com/turnerlabs/terraform-ecs-fargate/archive/v0.4.3.zip
installing terraform template

done
```

Now you have all the files you need to spin up something in Fargate.

Infrastructure:  provision using Terraform
```shell
cd iac/base
terraform init && terraform apply
cd ../env/dev
terraform init && terraform apply
```

Application:  build/push using Docker and deploy using Fargate CLI
```shell
docker-compose build
login=$(aws ecr get-login --no-include-email) && eval "$login"
docker-compose push
fargate service deploy -f docker-compose.yml
```

To scaffold out additional environnments, simply change the `environment` input parameter in `terraform.tfvars` and re-run
```shell
$ fargate-create
scaffolding my-app prod
Looking up AWS Account ID using profile: default
downloading terraform template https://github.com/turnerlabs/terraform-ecs-fargate/archive/v0.2.0.zip
installing terraform template
iac/base already exists, ignoring

done
```

And then bring up the new environment (no need to apply base again since it's shared):
```shell
cd ../prod
terraform init && terraform apply
```

You'll end up with a directory structure that looks something like this:
```
.
|____iac
| |____base
| |____env
| | |____dev
| | |____prod
```


### CI/CD (coming soon)

Using this technique, it's easy to codegen CI/CD pipelines for many popular build tools.  The `build` command support this. For example:

```shell
$ fargate-create build circleciv2
```

### Extensibility

`fargate-create` can scaffold out any Terraform template (specified by `-t`) that meets the following requirements:

- `base` and `env/dev` directory structure 
- a `env/dev/main.tf` with an s3 remote state backend
- `app` and `environment` input variables

Optionally:

- add a `fargate-create.yml` ([example here](examples/fargate-create.yml)) to your template to drive custom configuration, prompting for defaults, etc. 

For example (coming soon):
```shell
$ fargate-create -f my-scheduledtask.tfvars -t https://github.com/example/terraform-scheduledtask/archive/v0.1.0.zip
```
