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

// AdminInitiateAuth initiates the administrator sign-in flow for a user.
func (actor CognitoActions) AdminInitiateAuth(ctx context.Context, userPoolId string, clientId string, userName string, password string) (*cognitoidentityprovider.AdminInitiateAuthOutput, error) {
	authOutput, err := actor.CognitoClient.AdminInitiateAuth(ctx, &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeAdminUserPasswordAuth,
		ClientId:       aws.String(clientId),
		UserPoolId:     aws.String(userPoolId),
		AuthParameters: map[string]string{"USERNAME": userName, "PASSWORD": password},
	})
	if err != nil {
		log.Printf("Couldn't initiate auth for user %v. Here's why: %v\n", userName, err)
		return nil, err
	}
	return authOutput, nil
}
