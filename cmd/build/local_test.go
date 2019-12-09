package build

import (
	"fmt"
	"strings"
	"testing"
)

type mockContext struct {
	App     string
	Env     string
	Account string
	Region  string
}

func (c mockContext) GetApp() string {
	return c.App
}

func (c mockContext) GetEnvironment() string {
	return c.Env
}

func (c mockContext) GetAccount() string {
	return c.Account
}

func (c mockContext) GetRegion() string {
	return c.Region
}

func TestProvider_Local(t *testing.T) {

	ctx := mockContext{
		App:     "my-app",
		Env:     "dev",
		Account: "123456789",
		Region:  "us-east-1",
	}

	provider, err := GetProvider("local")
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
	if artifacts[0].FilePath != "build.sh" {
		t.Fail()
	}
	t.Log(artifacts[0].FileContents)

	image := fmt.Sprintf(`IMAGE="%v.dkr.ecr.us-east-1.amazonaws.com/%v:0.1.0"`, ctx.Account, ctx.App)
	if !strings.Contains(artifacts[0].FileContents, image) {
		t.Errorf("expecting " + image)
	}
}
