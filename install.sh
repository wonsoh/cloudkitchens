#!/bin/sh

mkdir -p $GOPATH/src/wonsoh.private/cloudkitchens
cp -r ./ $GOPATH/src/wonsoh.private/cloudkitchens
cd $GOPATH/src/wonsoh.private/cloudkitchens
go mod download
