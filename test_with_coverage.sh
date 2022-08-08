#!/bin/sh

go test ./... -coverprofile=c.out && go tool cover -html=c.out