package build

import (
	"fmt"
	"strings"
	"testing"
)

func TestProvider_CircleCIv2(t *testing.T) {

	ctx := mockContext{
		App:     "my-app",
		Env:     "dev",
		Account: "123456789",
		Region:  "us-west-1",
	}

	provider, err := GetProvider("circleciv2")
	if err != nil {
		t.Fail()
	}
	artifacts, err := provider.ProvideArtifacts(ctx)
	if err != nil {
		t.Fail()
	}
	if artifacts == nil {
		t.Fail()
	}

	t.Log(artifacts[0].FileContents)
	if artifacts[0].FilePath != ".circleci/config.yml" {
		t.Fail()
	}

	configEnv := artifacts[1].FileContents
	t.Log(configEnv)
	if artifacts[1].FilePath != ".circleci/config.env" {
		t.Fail()
	}

	lines := strings.Split(configEnv, "\n")

	if strings.TrimSpace(lines[1]) != fmt.Sprintf(`export FARGATE_CLUSTER="%v-%v"`, ctx.App, ctx.Env) {
		t.Error("not expecting", lines[1])
	}
	if strings.TrimSpace(lines[2]) != fmt.Sprintf(`export FARGATE_SERVICE="%v-%v"`, ctx.App, ctx.Env) {
		t.Error("not expecting", lines[2])
	}
	if strings.TrimSpace(lines[3]) != fmt.Sprintf(`export REPO="%v.dkr.ecr.%s.amazonaws.com/%v"`, ctx.Account, ctx.Region, ctx.App) {
		t.Error("not expecting", lines[3])
	}
	if strings.TrimSpace(lines[4]) != `export VERSION="0.1.0"` {
		t.Error("not expecting", lines[4])
	}
}
