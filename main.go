package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Kompiler struct {
	Package string
	Deps    []string
	Entry   string
	Build   map[string]string
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
		cmd := exec.Command(protogenic, "-f", kompiler.Entry, "-t", strings.ToLower(serviceType), "-m", kompiler.Package)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return err
		}
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = kompiler.Package
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command("go", "build", "-o", "app", fmt.Sprintf("%s/cmd", kompiler.Package))
		cmd.Dir = kompiler.Package
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			return err
		}
		err = os.Rename("app", fmt.Sprintf("%s/app", path))
		if err != nil {
			return err
		}
		err = os.RemoveAll(strings.FieldsFunc(kompiler.Package, func(r rune) bool {
			return r == '/' || r == '\\'
		})[0])
		if err != nil {
			return err
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
