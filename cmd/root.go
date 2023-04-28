package cmd

import (
	ctx "context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

const (
	tempDir                 = "fargate-create-template"
	templateConfigFile      = "fargate-create.yml"
	targetInfrastructureDir = "iac"
	varFormatHCL            = ".tfvars"
	varFormatJSON           = ".json"
	defaultTemplate         = "git@github.com:turnerlabs/terraform-ecs-fargate"
)

var verbose bool
var varFile string
var targetDir string
var templateURL string
var yesUseDefaults bool
var context scaffoldContext

var rootCmd = &cobra.Command{
	Use:              "fargate-create",
	Short:            "Scaffold out new AWS ECS/Fargate applications based on Terraform templates and Fargate CLI",
	Run:              run,
	PersistentPreRun: persistentPreRun,
	Example: `
# Scaffold an environment using the latest default template
fargate-create

# Do not prompt for options
fargate-create -y

# Use a template stored in github
fargate-create -t git@github.com:turnerlabs/terraform-ecs-fargate?ref=v0.4.3

# Scaffold out files for various build systems
fargate-create build circleciv2

# keep your template up to date
fargate-create upgrade

# Use a template stored in s3
AWS_ACCESS_KEY=xyz AWS_SECRET_KEY=xyz AWS_REGION=us-east-1 \
  fargate-create -t s3::https://s3.amazonaws.com/my-bucket/my-template
	
# Use a template stored in your file system
fargate-create -t ~/my-template

# Use a specific input file
fargate-create -f app.tfvars

# Use a JSON input file
fargate-create -f app.json
`,
}

// Execute ...
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&varFile, "file", "f", "terraform.tfvars", "file specifying Terraform input variables, in either HCL or JSON format")
	rootCmd.PersistentFlags().StringVarP(&targetDir, "target-dir", "d", targetInfrastructureDir, "target directory where code is outputted")
	rootCmd.PersistentFlags().StringVarP(&templateURL, "template", "t", defaultTemplate, "URL of a compatible Terraform template")
	rootCmd.PersistentFlags().BoolVarP(&yesUseDefaults, "yes", "y", false, "don't ask questions and use defaults")
}

type scaffoldContext struct {
	App           string
	Env           string
	Profile       string
	AccountID     string
	Region        string
	Format        string
	ContainerPort string
}

func (context scaffoldContext) GetApp() string {
	return context.App
}

func (context scaffoldContext) GetEnvironment() string {
	return context.Env
}

func (context scaffoldContext) GetAccount() string {
	return context.AccountID
}

func (context scaffoldContext) GetRegion() string {
	return context.Region
}

// gets run before every command
func persistentPreRun(cmd *cobra.Command, args []string) {

	if !(cmd.Name() == "fargate-create" || cmd.Name() == "build") {
		return
	}

	//validate that input varFile exists
	if _, err := os.Stat(varFile); os.IsNotExist(err) {
		fmt.Printf("Can't find %s. Use the --file flag to specify a .tfvars or .json file \n", varFile)
		os.Exit(-1)
	}

	//parse app, env, profile from input file
	fileBits, err := ioutil.ReadFile(varFile)
	check(err)
	varFormat := strings.ToLower(filepath.Ext(varFile))
	app, env, profile, region, containerPort, err := parseInputVars(varFormat, string(fileBits))
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
	context = scaffoldContext{
		App:           app,
		Env:           env,
		Profile:       profile,
		Region:        region,
		AccountID:     accountID,
		Format:        varFormat,
		ContainerPort: containerPort,
	}
}

func run(cmd *cobra.Command, args []string) {

	//scaffold out project environment
	scaffold(&context)

	fmt.Println()
	fmt.Println("done")
}

func getAWSAccountID(profile string) (string, error) {
	os.Setenv("AWS_PROFILE", profile)
	//call sts get-caller-identity
	cfg, err := config.LoadDefaultConfig(
		ctx.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	client := sts.NewFromConfig(cfg)

	identity, err := client.GetCallerIdentity(
		ctx.TODO(),
		&sts.GetCallerIdentityInput{},
	)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Printf(
		"Account: %s\nUserID: %s\nARN: %s\n",
		aws.ToString(identity.Account),
		aws.ToString(identity.UserId),
		aws.ToString(identity.Arn),
	)

	return *identity.Account, nil
}
