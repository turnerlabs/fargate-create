package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tmpDir = "./tmptest"

func TestIgnoreSensitiveFiles_New(t *testing.T) {

	//arrange
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	file := filepath.Join(tmpDir, ".gitignore")

	//act
	ignored := []string{"foo"}
	ensureFileContains(file, ignored)

	//assert
	//fail if file doesn't contain 1 instance of foo
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		t.Error(err)
	}
	contents := string(dat)
	t.Log(contents)
	if strings.Count(contents, "foo") != 1 {
		t.Error("expecting 1 occurance of foo")
	}
}

func TestIgnoreSensitiveFiles_Existing(t *testing.T) {

	//arrange
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	file := filepath.Join(tmpDir, ".gitignore")

	//create .gitignore with existing entry
	d1 := []byte("foo\nbar")
	err := ioutil.WriteFile(file, d1, 0644)
	if err != nil {
		t.Error(err)
	}

	//test
	ignored := []string{"foo"}
	ensureFileContains(file, ignored)

	//assert
	//fail if file doesn't contain 1 instance of foo
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		t.Error(err)
	}
	contents := string(dat)
	t.Log(contents)
	if strings.Count(contents, "foo") != 1 {
		t.Error("expecting 1 occurance of foo")
	}
}

func TestIgnoreSensitiveFiles_Existing_Multiple(t *testing.T) {

	//arrange
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	file := filepath.Join(tmpDir, ".gitignore")

	//create .gitignore with existing entry
	d1 := []byte("foo\nbar")
	err := ioutil.WriteFile(file, d1, 0644)
	if err != nil {
		t.Error(err)
	}

	//test
	ignored := []string{"foo", "baz"}
	ensureFileContains(file, ignored)

	//assert
	//fail if file doesn't contain 1 instance of foo
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		t.Error(err)
	}
	contents := string(dat)
	t.Log(contents)
	if strings.Count(contents, "foo") != 1 {
		t.Error("expecting 1 occurance of foo")
	}
	if strings.Count(contents, "bar") != 1 {
		t.Error("expecting 1 occurance of bar")
	}
	if strings.Count(contents, "baz") != 1 {
		t.Error("expecting 1 occurance of baz")
	}
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	//create tmpDir
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Error("error creating", tmpDir)
	}

	return func(t *testing.T) {
		t.Log("teardown test case")

		//delete
		os.RemoveAll(tmpDir)
	}
}

func TestDockerComposeYml_Port(t *testing.T) {

	context := scaffoldContext{
		AccountID:     "abc",
		App:           "my-app",
		ContainerPort: "3000",
		Env:           "dev",
		Profile:       "default",
		Region:        "us-east-1",
		Format:        ".tfvars",
	}

	yml := getDockerComposeYml(&context)
	t.Log(yml)

	expected := "- 80:3000"
	if !strings.Contains(yml, expected) {
		t.Errorf("expected: %s; actual: %s", expected, yml)
	}
}

func TestDockerComposeYml_NoPort(t *testing.T) {

	context := scaffoldContext{
		AccountID:     "abc",
		App:           "my-app",
		ContainerPort: "",
		Env:           "dev",
		Profile:       "default",
		Region:        "us-east-1",
		Format:        ".tfvars",
	}

	yml := getDockerComposeYml(&context)
	t.Log(yml)

	notexpected := "- 80:"
	if strings.Contains(yml, notexpected) {
		t.Errorf("not expected: %s; actual: %s", notexpected, yml)
	}	
}
