#!/bin/bash
# Run tests in golang docker container

pushd internal/browsers
go run init.go --alsologtostderr --download_browsers --download_latest
popd
go test -test.v --running_under_docker
