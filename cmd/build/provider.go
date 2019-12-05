package build

import (
	"errors"
	"os"
	"strings"
)

//Artifact represents a build artifact
type Artifact struct {
	FilePath     string
	FileContents string
	FileMode     os.FileMode
}

//Context represents a build context
type Context interface {
	GetApp() string
	GetEnvironment() string
	GetAccount() string
	GetRegion() string
}

type contextTemplate struct {
	App     string
	Env     string
	Account string
	Region  string
}

func getContextTemplate(context Context) contextTemplate {
	return contextTemplate{
		App:     context.GetApp(),
		Env:     context.GetEnvironment(),
		Account: context.GetAccount(),
		Region:  context.GetRegion(),
	}
}

func createArtifact(filePath string, fileContents string) *Artifact {
	return &Artifact{
		FilePath:     filePath,
		FileContents: fileContents,
		FileMode:     0644,
	}
}

//Provider represents a build provider
type Provider interface {
	ProvideArtifacts(context Context) ([]*Artifact, error)
}

//GetProvider returns a build provider based on its name
func GetProvider(provider string) (Provider, error) {
	providerString := strings.ToLower(provider)

	if providerString == "local" {
		return Local{}, nil
	}

	if providerString == "circleciv2" {
		return CircleCIv2{}, nil
	}

	if providerString == "githubactions" {
		return GithubActions{}, nil
	}

	return nil, errors.New("build provider not supported: " + provider)
}
