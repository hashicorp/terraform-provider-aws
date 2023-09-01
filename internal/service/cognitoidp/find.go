// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCognitoUserPoolUICustomization returns the UI Customization corresponding to the UserPoolId and ClientId.
// Returns nil if no UI Customization is found.
func FindCognitoUserPoolUICustomization(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId, clientId string) (*cognitoidentityprovider.UICustomizationType, error) {
	input := &cognitoidentityprovider.GetUICustomizationInput{
		ClientId:   aws.String(clientId),
		UserPoolId: aws.String(userPoolId),
	}

	output, err := conn.GetUICustomizationWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.UICustomization == nil {
		return nil, nil
	}

	// The GetUICustomization API operation will return an empty struct
	// if nothing is present rather than nil or an error, so we equate that with nil
	if reflect.DeepEqual(output.UICustomization, &cognitoidentityprovider.UICustomizationType{}) {
		return nil, nil
	}

	return output.UICustomization, nil
}

// FindCognitoUserInGroup checks whether the specified user is present in the specified group. Returns boolean value accordingly.
func FindCognitoUserInGroup(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, groupName, userPoolId, username string) (bool, error) {
	input := &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(username),
	}

	found := false

	err := conn.AdminListGroupsForUserPagesWithContext(ctx, input, func(page *cognitoidentityprovider.AdminListGroupsForUserOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.Groups {
			if group == nil {
				continue
			}

			if aws.StringValue(group.GroupName) == groupName {
				found = true
				break
			}
		}

		if found {
			return false
		}

		return !lastPage
	})

	if err != nil {
		return false, fmt.Errorf("reading groups for user: %w", err)
	}

	return found, nil
}

// FindCognitoUserPoolClientByID returns a Cognito User Pool Client using the ClientId
func FindCognitoUserPoolClientByID(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId, clientId string) (*cognitoidentityprovider.UserPoolClientType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(clientId),
		UserPoolId: aws.String(userPoolId),
	}

	output, err := conn.DescribeUserPoolClientWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.UserPoolClient == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.UserPoolClient, nil
}

func FindCognitoUserPoolClientByName(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) (*cognitoidentityprovider.UserPoolClientType, error) {
	clientDescs, err := listCognitoUserPoolClientDescriptions(ctx, conn, userPoolId, nameFilter)
	if err != nil {
		return nil, err
	}

	client, err := tfresource.AssertSinglePtrResult(clientDescs)
	if err != nil {
		return nil, err
	}

	return FindCognitoUserPoolClientByID(ctx, conn, userPoolId, aws.StringValue(client.ClientId))
}

type cognitoUserPoolClientDescriptionNameFilter func(string) (bool, error)

func listCognitoUserPoolClientDescriptions(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) ([]*cognitoidentityprovider.UserPoolClientDescription, error) {
	var errs *multierror.Error
	var descs []*cognitoidentityprovider.UserPoolClientDescription

	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolId),
	}
	err := conn.ListUserPoolClientsPagesWithContext(ctx, input, func(page *cognitoidentityprovider.ListUserPoolClientsOutput, lastPage bool) bool {
		for _, client := range page.UserPoolClients {
			if ok, err := nameFilter(aws.StringValue(client.ClientName)); err != nil {
				errs = multierror.Append(errs, err)
			} else if ok {
				descs = append(descs, client)
			}
		}
		return !lastPage
	})
	if err != nil {
		if e := errs.ErrorOrNil(); e == nil {
			return descs, err
		} else {
			return descs, multierror.Append(errs, err)
		}
	}
	return descs, nil
}

func FindRiskConfigurationById(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, id string) (*cognitoidentityprovider.RiskConfigurationType, error) {
	userPoolId, clientId, err := RiskConfigurationParseID(id)
	if err != nil {
		return nil, err
	}

	input := &cognitoidentityprovider.DescribeRiskConfigurationInput{
		UserPoolId: aws.String(userPoolId),
	}

	if clientId != "" {
		input.ClientId = aws.String(clientId)
	}

	output, err := conn.DescribeRiskConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RiskConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RiskConfiguration, nil
}
