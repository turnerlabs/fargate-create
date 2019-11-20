package build

import (
	"fmt"
	"strings"
	"testing"
)

func TestProvider_GithubActions(t *testing.T) {

	ctx := mockContext{
		App:     "my-app",
		Env:     "dev",
		Account: "123456789",
		Region:  "us-west-1",
	}

	provider, err := GetProvider("githubactions")
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
	if artifacts[0].FilePath != fmt.Sprintf(".github/workflows/%s.yml", ctx.Env) {
		t.Fail()
	}

	repo := fmt.Sprintf(`REPO: %v.dkr.ecr.%s.amazonaws.com/%v`, ctx.Account, ctx.Region, ctx.App)
	t.Log(repo)
	if !strings.Contains(yaml, repo) {
		t.Error("expecting", repo)
	}
	accessKey := "AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}"
	if !strings.Contains(yaml, accessKey) {
		t.Error("expecting", accessKey)
	}
	cluster := fmt.Sprintf("FARGATE_CLUSTER: %s-%s", ctx.App, ctx.Env)
	if !strings.Contains(yaml, cluster) {
		t.Error("expecting", cluster)
	}
}
