GOBIN ?= ${GOPATH}/bin

cmd: fmt vet
	go build -ldflags="-w -s" -o bin/addon github.com/konveyor/tackle2-addon-jkube/cmd