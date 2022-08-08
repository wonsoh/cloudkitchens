#!/bin/sh

go install github.com/golang/mock/mockgen@v1.6.0
$GOPATH/bin/mockgen -package=mocks -destination=mocks/mock_rand.go "math/rand" Source