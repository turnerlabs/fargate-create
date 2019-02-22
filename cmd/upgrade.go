package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeYes bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Keep a terraform template up to date",
	Run:   doUpgrade,
	Example: `
fargate-create upgrade
fargate-create upgrade -t git@github.com:turnerlabs/terraform-ecs-fargate-scheduled-task
fargate-create upgrade -d infrastructure
`,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func doUpgrade(cmd *cobra.Command, args []string) {

	//has the user done a previous install?
	notFound := false
	if _, err := os.Stat(filepath.Join(targetDir, baseDir)); err != nil {
		notFound = true
	}
	if _, err := os.Stat(filepath.Join(targetDir, envDir)); err != nil {
		notFound = true
	}
	if notFound {
		check(errors.New("no existing template found"))
	}

	//fetch the template from the source
	templateDir := downloadTerraformTemplate()
	debug("downloaded to:", templateDir)

	//process /base first, then iterate over /env
	srcDir := filepath.Join(templateDir, baseDir)
	destDir := filepath.Join(targetDir, baseDir)
	adds, updates := upgradeDirectory(srcDir, destDir)

	//process each installed environment
	srcDir = filepath.Join(templateDir, envDir, devDir)
	objects, err := ioutil.ReadDir(filepath.Join(targetDir, envDir))
	check(err)
	for _, o := range objects {
		if o.IsDir() {
			destDir = filepath.Join(targetDir, envDir, o.Name())
			debug(destDir)

			//look for default terraform.tfvars file for this environment
			tfVarsFile := filepath.Join(destDir, "terraform.tfvars")
			debug(tfVarsFile)
			if _, err = os.Stat(tfVarsFile); os.IsNotExist(err) {
				check(errors.New(tfVarsFile + " not found"))
			}

			fileBits, err := ioutil.ReadFile(tfVarsFile)
			check(err)
			varFormat := strings.ToLower(filepath.Ext(tfVarsFile))
			app, env, profile, _, _, err := parseInputVars(varFormat, string(fileBits))
			check(err)

			//apply env transformation in src before upgrading
			transformMainTFToContext(srcDir, profile, app, env)

			//upgrade env directory
			a, u := upgradeDirectory(srcDir, destDir)
			adds = append(adds, a...)
			updates = append(updates, u...)
		}
	}

	//delete download dir
	os.RemoveAll(templateDir)

	fmt.Println()
	fmt.Println("---------------------------------------")
	fmt.Printf("upgrade complete: %v add(s), %v update(s)\n", len(adds), len(updates))
	if len(adds) > 0 || len(updates) > 0 {
		fmt.Println()
		fmt.Println("updated files:")
		fmt.Println()
		for _, s := range updates {
			fmt.Printf("\t%s\n", s)
		}
		fmt.Println()
		fmt.Println("added files:")
		fmt.Println()
		for _, s := range adds {
			fmt.Printf("\t%s\n", s)
		}
		fmt.Println()
		fmt.Println(`run the following commands to apply these changes:
terraform init -upgrade=true
terraform apply`)
	}
}

func upgradeDirectory(srcDir string, destDir string) ([]string, []string) {

	//prompt for updates to existing local files
	//add new required files
	//prompt to add new optional files (using fargate-create.yml)

	fmt.Println()
	fmt.Println("---------------------------------------")
	fmt.Println("upgrading", destDir)
	fmt.Println("---------------------------------------")

	updates := []string{}
	adds := []string{}

	//iterate src files
	files, err := ioutil.ReadDir(srcDir)
	check(err)
	for _, f := range files {
		file := f.Name()
		debug(file)

		//only process .tf or .md files
		if !(strings.HasSuffix(file, ".tf") || strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".tpl")) {
			continue
		}

		//does matching file in template exist?
		source := filepath.Join(srcDir, file)
		dest := filepath.Join(destDir, file)
		debug(fmt.Sprintf("source: %s | dest: %s", source, dest))

		//can we access the source file?
		if _, err := os.Stat(source); err == nil {

			//does corresponding dest file exist?
			if _, err = os.Stat(dest); os.IsNotExist(err) {
				debug("new source file")

				//is this file required or optional?
				//load template config file
				templateConfig := loadTemplateConfig(srcDir)
				if templateConfig != nil {
					prompt := getFilePrompt(templateConfig, file)
					if prompt != nil {
						//prompt to install new optional file
						fmt.Println()
						q := fmt.Sprintf("%s (%s) ", prompt.Question, prompt.Default)
						response := promptAndGetResponse(q, prompt.Default)
						if containsString(okayResponses, response) {
							err = copyFile(source, dest)
							check(err)
							adds = append(adds, dest)
						}
					} else {
						//write new required file
						fmt.Println("writing", dest)
						copyFile(source, dest)
						adds = append(adds, dest)
					}
				} else {
					debug("no template config file found")
				}
			} else {
				//does dest file need updating?
				debug("diffing")
				if !deepCompare(source, dest) {
					fmt.Println()
					response := promptAndGetResponse(dest+" is out of date. Replace? (yes) ", "yes")
					if containsString(okayResponses, response) {
						err = copyFile(source, dest)
						check(err)
						updates = append(updates, dest)
					}
				} else {
					debug("files match")
				}
			}
		}
	}
	return adds, updates
}

func getFilePrompt(config *templateConfig, file string) *prompt {
	for _, p := range config.Prompts {
		for _, f := range p.FilesToDeleteIfNo {
			if f == file {
				return p
			}
		}
	}
	return nil
}

func deepCompare(file1, file2 string) bool {
	chunkSize := 4000

	f1, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
	}

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}
