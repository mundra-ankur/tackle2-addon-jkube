package main

import (
	"fmt"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"os"
	"path"
	"strings"
)

var (
	// hub integration.
	addon     = hub.Addon
	SourceDir = ""
	Dir       = ""
)

func init() {
	Dir, _ = os.Getwd()
	SourceDir = path.Join(Dir, "source")
}

type SoftError = hub.SoftError

// Data Addon data passed in the secret.
type Data struct {
	// Output directory within application bucket.
	Output string `json:"output" binding:"required"`
}

// main
func main() {
	addon.Run(func() (err error) {
		// Get the addon data associated with the task.
		d := &Data{}
		err = addon.DataWith(d)
		if err != nil {
			err = &SoftError{Reason: err.Error()}
			return
		}

		// Jkube
		jkube := Jkube{}
		jkube.Data = d

		// Fetch application.
		addon.Activity("Fetching application.")
		application, err := addon.Task.Application()
		if err == nil {
			jkube.application = application
		} else {
			return
		}

		// SSH
		agent := ssh.Agent{}
		err = agent.Start()
		if err != nil {
			return
		}

		// Fetch repository.
		addon.Total(2)
		if application.Repository == nil {
			err = &SoftError{Reason: "Application repository not defined."}
			return
		}
		SourceDir = path.Join(Dir, strings.Split(path.Base(application.Repository.URL), ".")[0])
		var r repository.Repository
		r, err = repository.New(SourceDir, application)
		if err != nil {
			return
		}
		err = r.Fetch()
		if err == nil {
			addon.Increment()
			jkube.repository = r
		} else {
			fmt.Printf("Error: %s\n", err)
			return
		}

		// Run jkube.
		err = jkube.Run()
		if err == nil {
			addon.Increment()
		} else {
			return
		}
		return
	})
}
