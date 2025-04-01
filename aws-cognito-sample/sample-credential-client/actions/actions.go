package actions

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"log"
)

type CognitoActions struct {
	CognitoClient *cognitoidentityprovider.Client
}

func NewCognitoActions(cognitoClient *cognitoidentityprovider.Client) *CognitoActions {
	return &CognitoActions{
		CognitoClient: cognitoClient,
	}
}

// AdminCreateUser uses administrator credentials to add a user to a user pool. This method leaves the user
// in a state that requires they enter a new password next time they sign in.
func (actor CognitoActions) AdminCreateUser(ctx context.Context, userPoolId string, userName string, userEmail string) error {
	_, err := actor.CognitoClient.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:    aws.String(userPoolId),
		Username:      aws.String(userName),
		MessageAction: types.MessageActionTypeSuppress,
		UserAttributes: []types.AttributeType{{
			Name:  aws.String("email"),
			Value: aws.String(userEmail)},
		},
	})
	if err != nil {
		var userExists *types.UsernameExistsException
		if errors.As(err, &userExists) {
			log.Printf("User %v already exists in the user pool.", userName)
			err = nil
		} else {
			log.Printf("Couldn't create user %v. Here's why: %v\n", userName, err)
		}
	}
	return err
}
