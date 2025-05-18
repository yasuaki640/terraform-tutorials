package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Message string `json:"message"`
}

// HandleRequest takes a context.Context and returns a Response containing
// a formatted string representation of the context, and a nil error
func HandleRequest(ctx context.Context) (Response, error) {
	return Response{
		Message: fmt.Sprintf("%#v", ctx),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
