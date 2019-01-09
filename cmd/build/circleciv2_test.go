package build

import (
	"strings"
	"fmt"
	"testing"
)

func TestProvider_CircleCIv2(t *testing.T) {

	ctx := mockContext{
		App:     "my-app",
		Env:     "dev",
		Account: "123456789",
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
	if artifacts[0].FilePath != ".circleci/config.yml" {
		t.Fail()
	}
	t.Log(artifacts[0].FileContents)

	if artifacts[1].FilePath != ".circleci/config.env" {
		t.Fail()
	}
	t.Log(artifacts[1].FileContents)

	configEnv := fmt.Sprintf(`
export FARGATE_CLUSTER="%v-%v"
export FARGATE_SERVICE"=%v-%v"
export REPO="%v.dkr.ecr.us-east-1.amazonaws.com/%v"
export VERSION="0.1.0"
	`, ctx.App, ctx.Env, ctx.App, ctx.Env, ctx.Account, ctx.App)
	t.Log(configEnv)

	if strings.Contains(strings.TrimSpace(artifacts[1].FileContents), strings.TrimSpace(configEnv)) {
		t.Errorf("unexpected config.env")
	}
}
