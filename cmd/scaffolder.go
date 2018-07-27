package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
)

func scaffold(context *scaffoldContext) {

	//scaffold out infrastructure files
	_, envDir := scaffoldInfrastructure(context)

	//scaffold application files
	scaffoldApplication(context, envDir)
}

func scaffoldInfrastructure(context *scaffoldContext) (string, string) {

	//fetch terraform template
	repoDir := downloadTerraformTemplate()
	debug("downloaded to:", repoDir)

	baseDir, envDir, baseDirInstalled := installTerraformTemplate(repoDir, context.Env)
	debug("environment installed to:", envDir)

	//copy var file into base module
	if baseDirInstalled {
		debug(fmt.Sprintf("copying %s to %s", varFile, baseDir))
		err := copyFile(varFile, fmt.Sprintf("%s/terraform.tfvars", baseDir))
		check(err)
	}

	//copy var file into environment module
	debug(fmt.Sprintf("copying %s to %s", varFile, envDir))
	err := copyFile(varFile, filepath.Join(envDir, "terraform.tfvars"))
	check(err)

	//update tf backend in main.tf to match app/env
	mainTfFile := filepath.Join(envDir, "main.tf")
	fileBits, err := ioutil.ReadFile(mainTfFile)
	check(err)
	maintf := updateTerraformBackend(string(fileBits), context.Profile, context.App, context.Env)
	err = ioutil.WriteFile(mainTfFile, []byte(maintf), 0644)
	check(err)

	return baseDir, envDir
}

func scaffoldApplication(context *scaffoldContext, envDir string) {

	//write the application files to the env directory
	targetAppDir := envDir

	//write a docker-compose.yml file
	dockerComposeYml := getDockerComposeYml(context)
	dockerComposeYmlFile := filepath.Join(targetAppDir, "docker-compose.yml")
	debug("writing", dockerComposeYmlFile)
	err := ioutil.WriteFile(dockerComposeYmlFile, []byte(dockerComposeYml), 0644)
	check(err)

	//write hidden.env
	hiddenEnvFileName := filepath.Join(targetAppDir, "hidden.env")
	sampleContents := "#FOO=bar\n"
	err = ioutil.WriteFile(hiddenEnvFileName, []byte(sampleContents), 0644)
	check(err)

	//write a fargate.yml for the cli
	fargateYml := getFargateYaml(context)
	fargateYmlFile := filepath.Join(targetAppDir, "fargate.yml")
	debug("writing", fargateYmlFile)
	err = ioutil.WriteFile(fargateYmlFile, []byte(fargateYml), 0644)
	check(err)

	//write deploy.sh
	deployScript := getDeployScript(context)
	deployScriptFile := filepath.Join(targetAppDir, "deploy.sh")
	debug("writing", deployScriptFile)
	err = ioutil.WriteFile(deployScriptFile, []byte(deployScript), 0755)
	check(err)

	//ignored file
	hiddenenv := strings.Split(hiddenEnvFileName, "/")
	ignoredFiles := []string{hiddenenv[len(hiddenenv)-1], ".terraform"}
	appendToFile(".gitignore", ignoredFiles)
	appendToFile(".dockerignore", ignoredFiles)
}

func getFargateYaml(context *scaffoldContext) string {
	textTemplate := `cluster: {{.App}}-{{.Env}}
service: {{.App}}-{{.Env}}
`
	return applyTemplate(textTemplate, context)
}

func getDockerComposeYml(context *scaffoldContext) string {
	t := `version: "3.4"
services:
	{{.App}}:
		build: ../../../
		image: {{.AccountID}}.dkr.ecr.{{.Region}}.amazonaws.com/{{.App}}:0.1.0
		ports:    
		- 80:8080
		env_file:
		- hidden.env	
`
	return applyTemplate(t, context)
}

func getDeployScript(context *scaffoldContext) string {
	t := `#! /bin/bash
set -e

# build image
docker-compose build

# push image to ECR repo
export AWS_PROFILE={{.Profile}}
export AWS_DEFAULT_REGION={{.Region}}
login=$(aws ecr get-login --no-include-email) && eval "$login"
docker-compose push

# deploy image and env vars
fargate service deploy -f docker-compose.yml
`
	return applyTemplate(t, context)
}

//fetches and installs the tf template and returns the output directory
func downloadTerraformTemplate() string {

	client := getter.Client{
		Src:  templateURL,
		Dst:  "./" + tempDir,
		Mode: getter.ClientModeDir,
	}

	fmt.Println("downloading terraform template", templateURL)
	err := client.Get()
	check(err)
	debug("done")

	repoDir, err := getter.SubdirGlob("./"+tempDir, "*")
	check(err)

	return repoDir
}

//installs a template for the specified environment,
//indicating whether or not the base directory was installed
func installTerraformTemplate(repoDir string, environment string) (string, string, bool) {

	//create infrastructure directory (if not already there)
	targetInfraDir := targetDir
	fmt.Println("installing terraform template")
	if _, err := os.Stat(targetInfraDir); os.IsNotExist(err) {
		debug("creating directory:", targetInfraDir)
		err = os.MkdirAll(targetInfraDir, 0755)
		check(err)
	} else {
		debug(targetInfraDir + " already exists")
	}

	//copy over infrastructure/base (if not already there)
	baseDir := "base"
	sourceBaseDir := filepath.Join(repoDir, baseDir)
	destBaseDir := filepath.Join(targetInfraDir, baseDir)
	createdBaseDir := false
	if _, err := os.Stat(destBaseDir); os.IsNotExist(err) {
		debug(fmt.Sprintf("copying %s to %s", sourceBaseDir, destBaseDir))
		err = copyDir(sourceBaseDir, destBaseDir)
		check(err)
		createdBaseDir = true
	} else {
		fmt.Println(destBaseDir + " already exists, ignoring")
	}

	//if environment directory exists, prompt to override, if no, then exit
	sourceEnvDir := filepath.Join(repoDir, "env", "dev")
	destEnvDir := filepath.Join(targetInfraDir, "env", environment)

	yes := true
	if _, err := os.Stat(destEnvDir); err == nil {
		//exists
		fmt.Print(destEnvDir + " already exists. Overwrite? ")
		if yes = askForConfirmation(); yes {
			debug("deleting", destEnvDir)
			//delete environment directory (all files)
			err = os.RemoveAll(destEnvDir)
			check(err)
		}
	} else {
		//doesn't exist
		debug(destEnvDir + " doesn't exist")
	}

	if yes {
		//env directory either doesn't exist or user wants to overwrite
		//copy repo/env/${env} -> ./infrastructure/env/${env}
		debug(fmt.Sprintf("copying %s to %s", sourceEnvDir, destEnvDir))
		err := copyDir(sourceEnvDir, destEnvDir)
		check(err)
	}

	//finally, delete temp dir
	debug("deleting:", tempDir)
	err := os.RemoveAll(tempDir)
	check(err)

	return destBaseDir, destEnvDir, createdBaseDir
}
