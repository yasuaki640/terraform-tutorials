#!/bin/bash

# build the Go application for lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
zip function.zip bootstrap
rm bootstrap

# deploy the lambda function using terraform
terraform apply
