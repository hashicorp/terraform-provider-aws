// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCognitoUserPoolClientByName(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) (*cognitoidentityprovider.UserPoolClientType, error) {
	clientDescs, err := listCognitoUserPoolClientDescriptions(ctx, conn, userPoolId, nameFilter)
	if err != nil {
		return nil, err
	}

	client, err := tfresource.AssertSinglePtrResult(clientDescs)
	if err != nil {
		return nil, err
	}

	return FindUserPoolClientByTwoPartKey(ctx, conn, userPoolId, aws.StringValue(client.ClientId))
}

type cognitoUserPoolClientDescriptionNameFilter func(string) (bool, error)

func listCognitoUserPoolClientDescriptions(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) ([]*cognitoidentityprovider.UserPoolClientDescription, error) {
	var errs []error
	var descs []*cognitoidentityprovider.UserPoolClientDescription

	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolId),
	}

	err := conn.ListUserPoolClientsPagesWithContext(ctx, input, func(page *cognitoidentityprovider.ListUserPoolClientsOutput, lastPage bool) bool {
		for _, client := range page.UserPoolClients {
			if ok, err := nameFilter(aws.StringValue(client.ClientName)); err != nil {
				errs = append(errs, err)
			} else if ok {
				descs = append(descs, client)
			}
		}
		return !lastPage
	})

	if err != nil {
		errs = append(errs, err)
		return descs, errors.Join(errs...)
	}

	return descs, nil
}
