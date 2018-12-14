package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/fargate-create/cmd/build"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Scaffold out artifacts for various build systems",
	Args:  cobra.ExactArgs(1),
	Run:   doBuild,
	Example: `
fargate-create build local
fargate-create build circleciv2
`,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func doBuild(cmd *cobra.Command, args []string) {

	//load build provider
	providerString := args[0]
	provider, err := build.GetProvider(providerString)
	check(err)

	//get artifacts
	artifacts, err := provider.ProvideArtifacts(context)
	check(err)

	//write artifacts to file system
	if artifacts != nil {
		for _, artifact := range artifacts {
			//create directories if needed
			dirs := filepath.Dir(artifact.FilePath)
			err = os.MkdirAll(dirs, os.ModePerm)
			check(err)

			if _, err := os.Stat(artifact.FilePath); err == nil {
				//exists
				fmt.Print(artifact.FilePath + " already exists. Overwrite? ")
				if askForConfirmation() {
					err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
					fmt.Println("wrote " + artifact.FilePath)
					check(err)
				}
			} else {
				//doesn't exist
				err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
				fmt.Println("wrote " + artifact.FilePath)
				check(err)
			}
		}
	}

}
