package build

//Local represents a local build provider
type Local struct{}

//ProvideArtifacts is the Provider implementation
func (provider Local) ProvideArtifacts(context Context) ([]*Artifact, error) {

	//output a build.sh that uses docker to build/push
	contextTemplate := getContextTemplate(context)
	artifacts := []*Artifact{}
	buildScript := createArtifact("build.sh", getLocalBuildScript(contextTemplate))
	buildScript.FileMode = 0700
	artifacts = append(artifacts, buildScript)
	return artifacts, nil
}

func getLocalBuildScript(context contextTemplate) string {
	textTemplate := `
#! /bin/bash
set -e

# build image
IMAGE="{{ .Account }}.dkr.ecr.us-east-1.amazonaws.com/{{ .App }}:0.1.0"
docker build -t ${IMAGE} .

# push image to ECR repo
login=$(aws ecr get-login --no-include-email) && eval "$login"
docker push ${IMAGE}
`
	return applyTemplate(textTemplate, context)
}
