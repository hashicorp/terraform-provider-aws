package cognitoidp

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCognitoUserPoolUICustomization returns the UI Customization corresponding to the UserPoolId and ClientId.
// Returns nil if no UI Customization is found.
func FindCognitoUserPoolUICustomization(conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId, clientId string) (*cognitoidentityprovider.UICustomizationType, error) {
	input := &cognitoidentityprovider.GetUICustomizationInput{
		ClientId:   aws.String(clientId),
		UserPoolId: aws.String(userPoolId),
	}

	output, err := conn.GetUICustomization(input)

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
func FindCognitoUserInGroup(conn *cognitoidentityprovider.CognitoIdentityProvider, groupName, userPoolId, username string) (bool, error) {
	input := &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(username),
	}

	found := false

	err := conn.AdminListGroupsForUserPages(input, func(page *cognitoidentityprovider.AdminListGroupsForUserOutput, lastPage bool) bool {
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
		return false, fmt.Errorf("error reading groups for user: %w", err)
	}

	return found, nil
}

func FindCognitoUserPoolClient(conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId, clientId string) (*cognitoidentityprovider.UserPoolClientType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(clientId),
		UserPoolId: aws.String(userPoolId),
	}

	output, err := conn.DescribeUserPoolClient(input)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindRiskConfigurationById(conn *cognitoidentityprovider.CognitoIdentityProvider, id string) (*cognitoidentityprovider.RiskConfigurationType, error) {
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

	output, err := conn.DescribeRiskConfiguration(input)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
