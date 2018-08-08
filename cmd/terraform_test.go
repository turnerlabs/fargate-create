package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestUpdateTerraformBackend(t *testing.T) {

	tf := `
terraform {
	backend "s3" {
		region  = "us-east-1"
		profile = ""
		bucket  = ""
		key     = "dev.terraform.tfstate"
	}
}

provider "aws" {
	region  = "${var.region}"
	profile = "${var.aws_profile}"
}
`

	profile := "my-profile"
	app := "my-app"
	env := "qa"
	result := updateTerraformBackend(tf, profile, app, env)
	t.Log(result)

	expected := fmt.Sprintf(`profile = "%s"`, profile)
	if !strings.Contains(result, expected) {
		t.Errorf("expected: %s; actual: %s", expected, result)
	}
	expected = fmt.Sprintf(`bucket  = "tf-state-%s"`, app)
	if !strings.Contains(result, expected) {
		t.Errorf("expected: %s; actual: %s", expected, result)
	}
}

func TestParseInputVars(t *testing.T) {

	tf := `
	region = "us-east-2"

	aws_profile = "default"
	
	saml_role = "devops"
	
	app = "my-app"
	
	environment = "qa"
	
	tags = {
		application   = ""
		environment   = "dev"
		team          = ""
		customer      = ""
		contact-email = ""
	}
	
	internal = "true"
`
	app, env, profile, region, err := parseInputVars(varFormatHCL, tf)

	t.Log(app)
	t.Log(env)
	t.Log(profile)
	t.Log(region)
	if err != nil {
		t.Error(err)
	}

	expected := "my-app"
	if app != expected {
		t.Errorf("expected: %s; actual: %s", expected, app)
	}
	expected = "qa"
	if env != expected {
		t.Errorf("expected: %s; actual: %s", expected, env)
	}
	expected = "default"
	if profile != expected {
		t.Errorf("expected: %s; actual: %s", expected, profile)
	}
	expected = "us-east-2"
	if region != expected {
		t.Errorf("expected: %s; actual: %s", expected, profile)
	}
}

func TestParseInputVars_JSON(t *testing.T) {

	tf := `{
  "region": "us-east-2",
	"aws_profile": "default",	
	"saml_role": "devops",	
	"app": "my-app",	
	"environment": "qa",	
	"tags": {
		"application": "",
		"environment": "dev",
		"team": "",
		"customer": "",
		"contact-email": ""
	},	
	"internal": true
}
`
	app, env, profile, region, err := parseInputVars(varFormatJSON, tf)

	t.Log(app)
	t.Log(env)
	t.Log(profile)
	t.Log(region)
	if err != nil {
		t.Error(err)
	}

	expected := "my-app"
	if app != expected {
		t.Errorf("expected: %s; actual: %s", expected, app)
	}
	expected = "qa"
	if env != expected {
		t.Errorf("expected: %s; actual: %s", expected, env)
	}
	expected = "default"
	if profile != expected {
		t.Errorf("expected: %s; actual: %s", expected, profile)
	}
	expected = "us-east-2"
	if region != expected {
		t.Errorf("expected: %s; actual: %s", expected, profile)
	}
}
