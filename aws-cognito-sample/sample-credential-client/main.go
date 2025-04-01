package main

import (
	"context"
	"github.com/yasuaki640/terraform-tutorials/aws-cognito-sample/sample-credential-client/actions"
	"github.com/yasuaki640/terraform-tutorials/aws-cognito-sample/sample-credential-client/appconf"
	"log"
	"time"

	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

func main() {
	appconf.Init()
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	client := cognitoidentityprovider.NewFromConfig(sdkConfig)

	cognitoActions := actions.NewCognitoActions(client)

	err = cognitoActions.AdminCreateUser(
		ctx,
		appconf.UserPoolID,
		"test"+fmt.Sprintf("%d", time.Now().Unix())+"@example.com",
		"test"+fmt.Sprintf("%d", time.Now().Unix())+"@example.com",
	)
	if err != nil {
		fmt.Println(err)
		return
	}
}
