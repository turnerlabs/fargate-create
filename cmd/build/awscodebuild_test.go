package build

import (
	"fmt"
	"strings"
	"testing"
)

func TestProvider_AWSCodeBuild(t *testing.T) {

	ctx := mockContext{
		App:     "my-app",
		Env:     "dev",
		Account: "123456789",
		Region:  "us-west-1",
	}

	provider, err := GetProvider("awscodebuild")
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

	yaml := artifacts[0].FileContents
	t.Log(yaml)
	if artifacts[0].FilePath != "buildspec.yml" {
		t.Fail()
	}

	repo := fmt.Sprintf(`REPO=%v.dkr.ecr.%s.amazonaws.com/%v`, ctx.Account, ctx.Region, ctx.App)
	t.Log(repo)
	if !strings.Contains(yaml, repo) {
		t.Error("expecting", repo)
	}
	cluster := fmt.Sprintf("FARGATE_CLUSTER=%s-%s", ctx.App, ctx.Env)
	if !strings.Contains(yaml, cluster) {
		t.Error("expecting", cluster)
	}
}
