package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

const (
	tempDir            = "fargate-create-template"
	templateConfigFile = "fargate-create.yml"
	varFormatHCL       = ".tfvars"
	varFormatJSON      = ".json"
)

var verbose bool
var varFile string
var targetDir string
var templateURL string
var yesUseDefaults bool

var rootCmd = &cobra.Command{
	Use:   "fargate-create",
	Short: "Scaffold out new AWS ECS/Fargate applications based on Terraform templates and Fargate CLI",
	Run:   run,
}

// Execute ...
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&varFile, "file", "f", "terraform.tfvars", "file specifying Terraform input variables, in either HCL or JSON format")
	rootCmd.PersistentFlags().StringVarP(&targetDir, "target-dir", "d", "iac", "target directory where code is outputted")
	rootCmd.PersistentFlags().StringVarP(&templateURL, "template", "t", "https://github.com/turnerlabs/terraform-ecs-fargate/archive/v0.2.0.zip", "URL of a compatible Terraform template")
	rootCmd.PersistentFlags().BoolVarP(&yesUseDefaults, "yes", "y", false, "don't ask questions and use defaults")
}

type scaffoldContext struct {
	App       string
	Env       string
	Profile   string
	AccountID string
	Region    string
	Format    string
}

func run(cmd *cobra.Command, args []string) {

	//validate that input varFile exists
	if _, err := os.Stat(varFile); os.IsNotExist(err) {
		fmt.Printf("Can't find %s. Use the --file flag to specify a .tfvars or .json file \n", varFile)
		os.Exit(-1)
	}

	//parse app, env, profile from input file
	fileBits, err := ioutil.ReadFile(varFile)
	check(err)
	varFormat := strings.ToLower(filepath.Ext(varFile))
	app, env, profile, region, err := parseInputVars(varFormat, string(fileBits))
	check(err)
	fmt.Printf("scaffolding %s %s\n", app, env)

	//lookup aws account id using profile
	debug("looking up AWS Account ID")
	fmt.Println("Looking up AWS Account ID using profile: " + profile)
	accountID, err := getAWSAccountID(profile)
	if err != nil {
		fmt.Println()
		fmt.Printf("The following error occurred while looking up AWS Account ID using profile: \"%s\". Please make sure the profile %s exists in ~/.aws/credentials and has valid keys:", profile, profile)
		fmt.Println()
		fmt.Println()
		fmt.Println(err)
		os.Exit(-1)
	}

	//set context for scaffolder
	context := scaffoldContext{
		App:       app,
		Env:       env,
		Profile:   profile,
		Region:    region,
		AccountID: accountID,
		Format:    varFormat,
	}

	//scaffold out project environment
	scaffold(&context)

	fmt.Println()
	fmt.Println("done")
}

func getAWSAccountID(profile string) (string, error) {
	//call sts get-caller-identity
	os.Setenv("AWS_PROFILE", profile)
	svc := sts.New(session.New())
	input := &sts.GetCallerIdentityInput{}
	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		return "", err
	}
	return *result.Account, nil
}
