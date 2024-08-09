package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	if name.Name == "error" {
		return "", fmt.Errorf("test error")
	}
	if name.Name == "errors.New" {
		return "", errors.New("test error.New")
	}
	if name.Name == "panic" {
		panic("test panic")
	}
	if name.Name == "timeout" {
		time.Sleep(100 * time.Second)
	}
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	lambda.Start(HandleRequest)
}
