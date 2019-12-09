package build

//AWSCodeBuild represents a Github Actions build provider
type AWSCodeBuild struct{}

//ProvideArtifacts is the Provider implementation
func (provider AWSCodeBuild) ProvideArtifacts(context Context) ([]*Artifact, error) {
	artifacts := []*Artifact{}
	artifacts = append(artifacts, createArtifact("buildspec.yml", getAWSBuildspecYAML(context)))
	return artifacts, nil
}

func getAWSBuildspecYAML(context Context) string {
	contextTemplate := getContextTemplate(context)

	textTemplate := `version: 0.2
  phases:
    install:
      runtime-versions:
        docker: 18
      commands:
        - nohup /usr/bin/dockerd --host=unix:///var/run/docker.sock --host=tcp://127.0.0.1:2375 --storage-driver=overlay2&
    pre_build:
      commands:
        - export FARGATE_CLUSTER={{ .App }}-{{ .Env }}
        - export FARGATE_SERVICE={{ .App }}-{{ .Env }}
        - export REPO={{ .Account }}.dkr.ecr.{{ .Region }}.amazonaws.com/{{ .App }}
  
        # build image:tag      
        - export VERSION=0.1.0
        # - export VERSION=$(jq -r .version < package.json)
        - export BUILD=$(echo ${CODEBUILD_BUILD_ID} | cut -d ":" -f 2)
        - export BRANCH=$(echo ${CODEBUILD_WEBHOOK_HEAD_REF} | cut -d "/" -f 3)
        - export IMAGE=${REPO}:${VERSION}-${BRANCH}.${BUILD}
  
        # login to ECR registry
        - login=$(aws ecr get-login --no-include-email) && eval "$login"
    build:
      commands:
        - docker build -t ${IMAGE} .
        - docker push ${IMAGE}
    post_build:
      commands:
        - fargate service deploy -i ${IMAGE}`

	return applyTemplate(textTemplate, contextTemplate)
}
