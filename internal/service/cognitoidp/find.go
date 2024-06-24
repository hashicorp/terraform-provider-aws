// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCognitoUserInGroup checks whether the specified user is present in the specified group. Returns boolean value accordingly.
func FindCognitoUserInGroup(ctx context.Context, conn *cognitoidentityprovider.Client, groupName, userPoolId, username string) (bool, error) {
	input := &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(username),
	}

	found := false

	pages := cognitoidentityprovider.NewAdminListGroupsForUserPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return false, fmt.Errorf("reading groups for user: %w", err)
		}

		for _, group := range page.Groups {
			if aws.ToString(group.GroupName) == groupName {
				found = true
				break
			}
		}
	}

	return found, nil
}

// FindCognitoUserPoolClientByID returns a Cognito User Pool Client using the ClientId
func FindCognitoUserPoolClientByID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolId, clientId string) (*awstypes.UserPoolClientType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(clientId),
		UserPoolId: aws.String(userPoolId),
	}

	output, err := conn.DescribeUserPoolClient(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindCognitoUserPoolClientByName(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) (*awstypes.UserPoolClientType, error) {
	clientDescs, err := listCognitoUserPoolClientDescriptions(ctx, conn, userPoolId, nameFilter)
	if err != nil {
		return nil, err
	}

	client, err := tfresource.AssertSingleValueResult(clientDescs)
	if err != nil {
		return nil, err
	}

	return FindCognitoUserPoolClientByID(ctx, conn, userPoolId, aws.ToString(client.ClientId))
}

type cognitoUserPoolClientDescriptionNameFilter func(string) (bool, error)

func listCognitoUserPoolClientDescriptions(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolId string, nameFilter cognitoUserPoolClientDescriptionNameFilter) ([]awstypes.UserPoolClientDescription, error) {
	var errs []error
	var descs []awstypes.UserPoolClientDescription

	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolId),
	}

	pages := cognitoidentityprovider.NewListUserPoolClientsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = append(errs, err)
			return descs, errors.Join(errs...)
		}

		for _, client := range page.UserPoolClients {
			if ok, err := nameFilter(aws.ToString(client.ClientName)); err != nil {
				errs = append(errs, err)
			} else if ok {
				descs = append(descs, client)
			}
		}

	}

	return descs, nil
}

func FindRiskConfigurationById(ctx context.Context, conn *cognitoidentityprovider.Client, id string) (*awstypes.RiskConfigurationType, error) {
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

	output, err := conn.DescribeRiskConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
