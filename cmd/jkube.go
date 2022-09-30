package main

import (
	"encoding/xml"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"os"
	"path"
	"strings"
	"tackle2-addon-jkube/pom"
)

type Jkube struct {
	application *api.Application
	*Data
	repository repository.Repository
}

func (r *Jkube) Run() (err error) {
	output := r.output()
	cmd := command.Command{Path: "/usr/bin/rm"}
	cmd.Options.Add("-rf", output)
	err = cmd.Run()
	if err != nil {
		return
	}
	err = os.MkdirAll(output, 0777)
	if err != nil {
		err = liberr.Wrap(err, "path", output)
		return
	}
	addon.Activity("[Jkube] created: %s.", output)

	// Fetch the repository.
	err = r.repository.Fetch()
	if err != nil {
		return err
	}

	// Add the jkube plugin to the pom.xml
	groupId, artifactId, err := r.addJkubePlugin()
	if err != nil {
		return err
	}

	// Build the maven project
	err = r.buildMvnProject()
	if err != nil {
		return err
	}

	// Copy the resources to the output directory
	err = r.copyResources(groupId, artifactId)
	if err != nil {
		return err
	}
	return
}

// output returns output directory.
func (r *Jkube) output() string {
	return path.Join(r.application.Bucket, r.Output)
}

func (r *Jkube) addJkubePlugin() (groupId string, artifactId string, err error) {
	pomXml := path.Join(SourceDir, "pom.xml") // Path to the pom.xml file
	parsedPom, err := pom.Parse(pomXml)       // Parse the pom.xml file
	if err != nil {
		fmt.Printf("Error parsing pom.xml %s", err)
		return
	}

	jkubePlugin := pom.Plugin{
		GroupID:    "org.eclipse.jkube",
		ArtifactID: "kubernetes-maven-plugin",
		Version:    "1.9.1",
	}

	*parsedPom.Build.Plugins = append(*parsedPom.Build.Plugins, jkubePlugin)

	// Marshal the pom back to xml
	output, err := xml.MarshalIndent(parsedPom, "  ", "    ")
	if err != nil {
		fmt.Printf("Error marshalling pom.xml %s", err)
		return
	}

	// Write the xml to a file
	err = os.Chdir(SourceDir)
	err = os.WriteFile("pom.xml", output, 0644)
	if err != nil {
		fmt.Printf("Error writing pom.xml %s", err)
		return
	}

	return parsedPom.GroupID, parsedPom.ArtifactID, nil
}

func (r *Jkube) buildMvnProject() (err error) {
	// Run mvn k8s:build
	cmd := command.Command{
		Path:    "./mvnw",
		Options: []string{"k8s:build"},
		Dir:     SourceDir,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running mvn k8s:build %s", err)
		return
	}

	// Run mvn k8s:resource
	cmd = command.Command{
		Path:    "./mvnw",
		Options: []string{"k8s:resource"},
		Dir:     SourceDir,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running mvn k8s:resource %s", err)
		return
	}

	return
}

func (r *Jkube) copyResources(groupId string, artifactId string) (err error) {
	// Copy the k8s resources to the output directory
	cmd := command.Command{
		Path: "/usr/bin/cp",
		Options: []string{"-r",
			path.Join(SourceDir, "target", "classes", "META-INF", "jkube"),
			path.Join(r.output(), "manifest")},
		Dir: SourceDir,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error copying resources %s", err)
		return
	}

	// Copy the Dockerfile to the output directory
	group := strings.ToLower(strings.Split(groupId, ".")[0])
	artifactId = strings.ToLower(artifactId)
	cmd = command.Command{
		Path: "/usr/bin/cp",
		Options: []string{
			path.Join(SourceDir, "target", "docker", group, artifactId, "latest", "build", "Dockerfile"),
			path.Join(r.output(), "Dockerfile")},
		Dir: SourceDir,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error copying Dockerfile %s", err)
		return
	}
	return
}
