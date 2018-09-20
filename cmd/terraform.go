package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func parseInputVars(format string, input string) (string, string, string, string, string, error) {
	if format == varFormatHCL {
		return parseInputVarsHCL(input)
	}
	if format == varFormatJSON {
		return parseInputVarsJSON(input)
	}
	return "", "", "", "", "", errors.New(`unknown var format: "` + format + `"`)
}

func parseInputVarsJSON(input string) (string, string, string, string, string, error) {
	var data map[string]interface{}

	err := json.Unmarshal([]byte(input), &data)
	check(err)

	app := data["app"].(string)
	environment := data["environment"].(string)
	profile := data["aws_profile"].(string)
	region := data["region"].(string)
	containerPort := data["container_port"].(string)

	//did we find it?
	if app == "" {
		return "", "", "", "", "", errors.New(`missing variable: "app"`)
	}
	if environment == "" {
		return "", "", "", "", "", errors.New(`missing variable: "environment"`)
	}
	if profile == "" {
		return "", "", "", "", "", errors.New(`missing variable: "profile"`)
	}
	if region == "" {
		return "", "", "", "", "", errors.New(`missing variable: "region"`)
	}

	return app, environment, profile, region, containerPort, nil
}

func parseInputVarsHCL(tf string) (string, string, string, string, string, error) {
	app := ""
	environment := ""
	profile := ""
	region := ""
	containerPort := ""

	//look for variables
	lines := strings.Split(tf, "\n")
	inTags := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		//ignore whitespace and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "tags = {" {
			inTags = true
			continue
		}
		if inTags {
			if trimmed == "}" {
				inTags = false
			}
			continue
		}
		//key = "value"
		parts := strings.Split(trimmed, "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			//remove quotes (app = "foo")
			value = strings.Replace(value, `"`, "", -1)

			//remove trailing comments (app = "foo" #comment)
			comments := strings.Split(value, "#")
			value = strings.TrimSpace(comments[0])

			if key == "app" {
				app = value
			}
			if key == "environment" {
				environment = value
			}
			if key == "aws_profile" {
				profile = value
			}
			if key == "region" {
				region = value
			}
			if key == "container_port" {
				containerPort = value
			}
		}
	}

	//did we find it?
	if app == "" {
		return "", "", "", "", "", errors.New(`missing variable: "app"`)
	}
	if environment == "" {
		return "", "", "", "", "", errors.New(`missing variable: "environment"`)
	}
	if profile == "" {
		return "", "", "", "", "", errors.New(`missing variable: "profile"`)
	}
	if region == "" {
		return "", "", "", "", "", errors.New(`missing variable: "region"`)
	}

	return app, environment, profile, region, containerPort, nil
}

func updateTerraformBackend(tf string, profile string, app string, env string) string {
	//update terraform.backend (which doesn't support dynamic variables)
	// profile = ""
	// bucket  = ""
	// key     = "dev.terraform.tfstate"
	tmp := strings.Split(tf, "\n")
	newTf := ""
	for _, line := range tmp {
		updatedLine := line
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `profile = ""`) {
			updatedLine = fmt.Sprintf(`    profile = "%s"`, profile)
		}
		if strings.HasPrefix(trimmed, "bucket") {
			updatedLine = fmt.Sprintf(`    bucket  = "tf-state-%s"`, app)
		}
		if strings.HasPrefix(trimmed, "key") {
			updatedLine = fmt.Sprintf(`    key     = "%s.terraform.tfstate"`, env)
		}
		newTf += updatedLine + "\n"
	}
	return newTf
}
