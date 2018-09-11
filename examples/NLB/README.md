### Network Application (NLB)

[template](https://github.com/turnerlabs/terraform-ecs-fargate-nlb)

`terraform.tfvars`
```HCL
# app/env to scaffold
app = "my-app"
environment = "dev"

internal = "true"
container_port = "8080"
replicas = "1"
region = "us-east-1"
aws_profile = "default"
saml_role = "admin"
vpc = "vpc-123"
private_subnets = "subnet-123,subnet-456"
public_subnets = "subnet-789,subnet-012"
tags = {
  application   = "my-app"
  environment   = "dev"
  team          = "my-team"
  customer      = "my-customer"
  contact-email = "me@example.com"
}
```

```shell
$ fargate-create -f terraform.tfvars -t git@github.com:turnerlabs/terraform-ecs-fargate-nlb
```