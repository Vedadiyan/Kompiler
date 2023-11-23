package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed dockerfile.go.tmpl
	dockerfile         string
	dockerfileTemplate *template.Template
)

type Kompiler struct {
	Package string
	Deps    []string
	Entry   string
	Build   map[string]string
}

type Dockerfile struct {
	Type string
}

func init() {
	_dockerfileTemplate, err := template.New("dockerfile").Parse(dockerfile)
	if err != nil {
		panic(err)
	}
	dockerfileTemplate = _dockerfileTemplate
}

func (kompiler Kompiler) InstallDeps() error {
	protocHome := os.Getenv("PROTOC_HOME")
	for _, dep := range kompiler.Deps {
		depSegments := strings.Split(dep, "/")
		path := fmt.Sprintf("%s/include/%s", protocHome, depSegments[len(depSegments)-1])
		_, err := os.ReadDir(path)
		if err == nil {
			continue
		}
		fmt.Println("cloning", dep)
		cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s.git", dep), path)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			return err
		}
		fmt.Println("done")
	}
	return nil
}

func (kompiler Kompiler) Compile() error {
	protogenic := os.Getenv("PROTOGENIC")
	for serviceType, path := range kompiler.Build {
		fmt.Println("compiling", serviceType)
		{
			cmd := exec.Command(protogenic, "-f", kompiler.Entry, "-t", strings.ToLower(serviceType), "-m", kompiler.Package)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return err
			}
		}
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		{
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Env = os.Environ()
			cmd.Dir = kompiler.Package
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return err
			}
		}
		{
			cmd := exec.Command("go", "build", "-o", "app", fmt.Sprintf("%s/cmd", kompiler.Package))
			cmd.Env = os.Environ()
			cmd.Dir = kompiler.Package
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return err
			}
		}
		var buffer bytes.Buffer
		err = dockerfileTemplate.Execute(&buffer, Dockerfile{Type: serviceType})
		if err != nil {
			return err
		}
		err = os.WriteFile(fmt.Sprintf("%s/Dockerfile", path), buffer.Bytes(), os.ModePerm)
		if err != nil {
			return err
		}
		os.Rename(fmt.Sprintf("%s/app", kompiler.Package), fmt.Sprintf("%s/app", path))
		for i := 0; i <= 10; i++ {
			err = os.RemoveAll(strings.FieldsFunc(kompiler.Package, func(r rune) bool {
				return r == '/' || r == '\\'
			})[0])
			if err == nil {
				break
			}
			<-time.After(time.Second * time.Duration(i))
			if i == 10 {
				return err
			}
		}
		fmt.Println("done")
	}
	return nil
}

func main() {
	wd := os.Args[0]
	kompilerYaml, err := os.ReadFile(fmt.Sprintf("%s/.kompiler/kompile.yaml", filepath.Dir(wd)))
	if err != nil {
		panic(err)
	}
	kompiler := Kompiler{}
	err = yaml.Unmarshal(kompilerYaml, &kompiler)
	if err != nil {
		panic(err)
	}
	err = kompiler.InstallDeps()
	if err != nil {
		panic(err)
	}
	err = kompiler.Compile()
	if err != nil {
		panic(err)
	}
}
