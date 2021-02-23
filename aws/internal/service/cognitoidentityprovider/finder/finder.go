package finder

import (
	"reflect"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// CognitoUserPoolUICustomization returns the UI Customization corresponding to the UserPoolId and ClientId, if provided.
// Returns nil if no UI Customization is found.
func CognitoUserPoolUICustomization(conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId, clientId *string) (*cognitoidentityprovider.UICustomizationType, error) {
	input := &cognitoidentityprovider.GetUICustomizationInput{
		UserPoolId: userPoolId,
	}

	if clientId != nil {
		input.ClientId = clientId
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
	// to prevent non-empty plans of an empty ui_customization block
	if reflect.DeepEqual(output.UICustomization, &cognitoidentityprovider.UICustomizationType{}) {
		return nil, nil
	}

	return output.UICustomization, nil
}
