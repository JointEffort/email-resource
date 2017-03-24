#!/bin/bash

mkdir ../bin
export GOOS=linux
export GOARCH=amd64

cd ../actions/check
go build -o ../../bin/check

cd ../in
go build -o ../../bin/in

cd ../out
go build -o ../../bin/out

cd ../../
docker build --tag jointeffort/email-resource:latest .
docker push jointeffort/email-resource:latest
