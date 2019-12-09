package build

import "fmt"

//GithubActions represents a Github Actions build provider
type GithubActions struct{}

//ProvideArtifacts is the Provider implementation
func (provider GithubActions) ProvideArtifacts(context Context) ([]*Artifact, error) {
	artifacts := []*Artifact{}
	artifacts = append(artifacts, createArtifact(fmt.Sprintf(".github/workflows/%s.yml", context.GetEnvironment()), getGithubActionsYAML(context)))

	fmt.Println()
	fmt.Println(`Be sure to add the following secrets to your Github repository:
  AWS_ACCESS_KEY_ID (terraform state show aws_iam_access_key.cicd_keys)
  AWS_SECRET_ACCESS_KEY (terraform state show aws_iam_access_key.cicd_keys)`)
	fmt.Println()

	return artifacts, nil
}

func getGithubActionsYAML(context Context) string {
	contextTemplate := getContextTemplate(context)

	textTemplate := `name: {{ .Env }}
on:
  push:
    branches:
      - develop
jobs:
  cicd:
    name: Deploy develop branch to {{ .Env }} environment
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master

      - name: Set docker image
        env:
          REPO: {{ .Account }}.dkr.ecr.{{ .Region }}.amazonaws.com/{{ .App }}
          VERSION: 0.1.0
        run: |
          BRANCH=$(echo $GITHUB_REF | cut -d "/" -f 3)
          SHA_SHORT=$(echo $GITHUB_SHA | head -c7)
          echo "export IMAGE=$REPO:$VERSION-$BRANCH.$SHA_SHORT" >> ./env
          cat ./env
      - name: Build image
        uses: turnerlabs/fargate-cicd-action@master
        with:
          args: . ./env; docker build -t $IMAGE .

      - name: Login to ECR
        uses: turnerlabs/fargate-cicd-action@master
        env:
          AWS_DEFAULT_REGION: {{ .Region }}
          AWS_ACCESS_KEY_ID: ${{"{{ secrets.AWS_ACCESS_KEY_ID }}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{ secrets.AWS_SECRET_ACCESS_KEY }}"}}
        with:
          args: login=$(aws ecr get-login --no-include-email) && eval "$login"

      - name: Push image to ECR
        uses: turnerlabs/fargate-cicd-action@master
        with:
          args: . ./env; docker push $IMAGE

      - name: Deploy image to fargate
        uses: turnerlabs/fargate-cicd-action@master
        env:
          AWS_DEFAULT_REGION: {{ .Region }}
          AWS_ACCESS_KEY_ID: ${{"{{ secrets.AWS_ACCESS_KEY_ID }}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{ secrets.AWS_SECRET_ACCESS_KEY }}"}}
          FARGATE_CLUSTER: {{ .App }}-{{ .Env }}
          FARGATE_SERVICE: {{ .App }}-{{ .Env }}
        with:
          args: . ./env; fargate service deploy -i $IMAGE`

	return applyTemplate(textTemplate, contextTemplate)
}
