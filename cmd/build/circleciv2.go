package build

import "fmt"

//CircleCIv2 represents a circle ci v2 build provider
type CircleCIv2 struct{}

//ProvideArtifacts is the Provider implementation
func (provider CircleCIv2) ProvideArtifacts(context Context) ([]*Artifact, error) {

	contextTemplate := getContextTemplate(context)

	artifacts := []*Artifact{}
	artifacts = append(artifacts, createArtifact(".circleci/config.yml", getCircleCIv2YAML()))
	artifacts = append(artifacts, createArtifact(".circleci/config.env", getConfigEnv(contextTemplate)))

	fmt.Println()
	fmt.Println(`Be sure to supply the following environment variables in your Circle CI build:
  AWS_ACCESS_KEY_ID (terraform state show aws_iam_access_key.cicd_keys)
  AWS_SECRET_ACCESS_KEY (terraform state show aws_iam_access_key.cicd_keys)
  AWS_DEFAULT_REGION=us-east-1`)
	fmt.Println()

	return artifacts, nil
}

func getCircleCIv2YAML() string {
	return `
version: 2
jobs:
  build:
    docker:
      - image: quay.io/turner/fargate-cicd
    environment:
      VAR: .circleci/config.env
    steps:
      - checkout
      - setup_remote_docker:
          version: 18.06.0-ce
      - run:
          name: Set docker image
          command: |
            source ${VAR}
            # either manage version in config.env or some other way
            # for node.js apps you can use version from package.json
            # VERSION=$(jq -r .version < package.json)
            BUILD=${CIRCLE_BUILD_NUM}
            if [ "${CIRCLE_BRANCH}" != "master" ]; then
              BUILD=${CIRCLE_BRANCH}.${CIRCLE_BUILD_NUM}
            fi
            echo "export IMAGE=${REPO}:${VERSION}-${BUILD}" >> ${VAR}
            cat ${VAR}
      - run:        
          name: Login to registry
          command: login=$(aws ecr get-login --no-include-email) && eval "$login"
      - run:
          name: Build app image
          command: . ${VAR}; docker build -t ${IMAGE} .
      - run:
          name: Push app image to registry
          command: . ${VAR}; docker push ${IMAGE}
      - run:
          name: Deploy
          command: . ${VAR}; fargate service deploy -i ${IMAGE}`
}

func getConfigEnv(context contextTemplate) string {
	textTemplate := `export FARGATE_CLUSTER="{{ .App }}-{{ .Env }}"
export FARGATE_SERVICE="{{ .App }}-{{ .Env }}"
export REPO="{{ .Account }}.dkr.ecr.us-east-1.amazonaws.com/{{ .App }}"
export VERSION="0.1.0"
`
	return applyTemplate(textTemplate, context)
}
