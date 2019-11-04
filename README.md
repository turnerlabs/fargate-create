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

Specify your template's input parameters in [terraform.tfvars](https://www.terraform.io/docs/configuration/variables.html) (or terraform.json).  The [default web application template's](https://github.com/turnerlabs/terraform-ecs-fargate) input looks something like this. Note: [any Terraform template can be used](#extensibility)

```hcl
# app/env to scaffold
app = "my-app"
environment = "dev"

internal = "true"
container_port = "8080"
replicas = "1"
health_check = "/health"
region = "us-east-1"
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
downloading terraform template git@github.com:turnerlabs/terraform-ecs-fargate
installing terraform template

done
```

Now you have all the files you need to spin up something in Fargate. Note that the Terraform files can be edited or customized. You can also use your own Terraform template using the `--template` flag.

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
downloading terraform template git@github.com:turnerlabs/terraform-ecs-fargate
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

As changes are made to the various upstream templates over time, the `upgrade` command can be used to keep installed versions up to date.

```bash
fargate-create upgrade
```


### Stacks

The following stacks are popular configurations that can be used with `fargate-create`

- [Web Application (ALB - HTTP or HTTPS)](https://github.com/turnerlabs/terraform-ecs-fargate#fargate-create)
- [Web Application (ALB - HTTPS/DNS)](https://github.com/turnerlabs/terraform-ecs-fargate-dns-https#fargate-create)
- [Web API Gateway](https://github.com/turnerlabs/terraform-ecs-fargate-apigateway#fargate-create)
- [Scheduled Task](https://github.com/turnerlabs/terraform-ecs-fargate-scheduled-task#fargate-create)
- [Background Worker (service)](https://github.com/turnerlabs/terraform-ecs-fargate-background-worker#fargate-create)
- [Network Application (NLB)](https://github.com/turnerlabs/terraform-ecs-fargate-nlb#fargate-create)
- [Airflow](https://github.com/turnerlabs/terraform-ecs-fargate-airflow#fargate-create)
 

### Help

```
Scaffold out new AWS ECS/Fargate applications based on Terraform templates and Fargate CLI

Usage:
  fargate-create [flags]
  fargate-create [command]

Examples:

# Scaffold an environment using the latest default template
fargate-create

# Do not prompt for options
fargate-create -y

# Use a template stored in github
fargate-create -t git@github.com:turnerlabs/terraform-ecs-fargate?ref=v0.4.3

# Scaffold out files for various build systems
fargate-create build circleciv2

# keep your template up to date
fargate-create upgrade

# Use a template stored in s3
AWS_ACCESS_KEY=xyz AWS_SECRET_KEY=xyz AWS_REGION=us-east-1 \
  fargate-create -t s3::https://s3.amazonaws.com/my-bucket/my-template

# Use a template stored in your file system
fargate-create -t ~/my-template

# Use a specific input file
fargate-create -f app.tfvars

# Use a JSON input file
fargate-create -f app.json


Available Commands:
  build       Scaffold out artifacts for various build systems
  help        Help about any command
  upgrade     Keep a terraform template up to date

Flags:
  -f, --file string         file specifying Terraform input variables, in either HCL or JSON format (default "terraform.tfvars")
  -h, --help                help for fargate-create
  -d, --target-dir string   target directory where code is outputted (default "iac")
  -t, --template string     URL of a compatible Terraform template (default "git@github.com:turnerlabs/terraform-ecs-fargate")
  -v, --verbose             Verbose output
      --version             version for fargate-create
  -y, --yes                 don't ask questions and use defaults
```


### CI/CD

Using this technique, it's easy to codegen CI/CD pipelines for many popular build tools.  The `build` command supports this. For example:

```shell
$ fargate-create build circleciv2
```

### Extensibility

`fargate-create` can scaffold out any Terraform template (specified by `--template`) that meets the following requirements:

- `base` and `env/dev` directory structure 
- a `env/dev/main.tf` with an s3 remote state backend
- `app` and `environment` input variables

Your template can be downloaded from a variety of locations using a variety of protocols.  The following are supported:

- Local files (`~/my-template`)
- Git (`git@github.com:my-org/my-template`)
- Amazon S3 (`s3::https://s3.amazonaws.com/my-bucket/my-template`)
- HTTP (`http://server/my-template/`)

Optionally:

- add a `fargate-create.yml` ([example here](examples/fargate-create.yml)) to your template to drive custom configuration, prompting for defaults, etc. 

An [example](https://github.com/turnerlabs/terraform-ecs-fargate-scheduled-task/) of an extended template:
```shell
$ fargate-create -f my-scheduledtask.tfvars -t git@github.com:turnerlabs/terraform-ecs-fargate-scheduled-task
```
