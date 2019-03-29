package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/schema"
)

func cognitoUserPoolUICustomizationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"css": {
					Type:     schema.TypeString,
					Optional: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return new == stringHashSum(old)
					},
					StateFunc: func(v interface{}) string {
						switch v.(type) {
						case string:
							return stringHashSum(v.(string))
						default:
							return ""
						}
					},
				},
				"image_file": {
					Type:     schema.TypeString,
					Optional: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return new == remoteFileHashSum(old)
					},
					StateFunc: func(v interface{}) string {
						switch v.(type) {
						case string:
							return remoteFileHashSum(v.(string))
						default:
							return ""
						}
					},
				},
			},
		},
	}
}

func cognitoUserPoolUICustomizationSet(d *schema.ResourceData, conn *cognitoidentityprovider.CognitoIdentityProvider) error {
	var params *cognitoidentityprovider.SetUICustomizationInput

	// UI customization is applied to a single client or to all uncustomized pool
	// clients, depending on the provided ResourceData.
	// See https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_SetUICustomization.html

	if d.Get("user_pool_id") == nil {
		params = &cognitoidentityprovider.SetUICustomizationInput{
			UserPoolId: aws.String(d.Id()),
		}
	} else {
		params = &cognitoidentityprovider.SetUICustomizationInput{
			ClientId:   aws.String(d.Id()),
			UserPoolId: aws.String(d.Get("user_pool_id").(string)),
		}
	}

	if v, ok := d.GetOk("ui_customization"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			if v, ok := config["css"]; ok && v.(string) != "" {
				params.CSS = aws.String(v.(string))
			}

			if v, ok := config["image_file"]; ok && v.(string) != "" {
				body, err := remoteFileContent(v.(string))

				if err != nil {
					return fmt.Errorf("Error reading image file: %s", err.Error())
				}

				params.ImageFile = body
			}
		}
	} else {
		params.CSS = nil
		params.ImageFile = nil
	}

	_, err := conn.SetUICustomization(params)

	return err
}

func cognitoUserPoolUICustomizationGet(d *schema.ResourceData, conn *cognitoidentityprovider.CognitoIdentityProvider) (*cognitoidentityprovider.GetUICustomizationOutput, error) {
	var params *cognitoidentityprovider.GetUICustomizationInput

	// UI customization is read from a single client or from the user pool,
	// depending on the provided ResourceData.
	// See https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_GetUICustomization.html

	if d.Get("user_pool_id") == nil {
		params = &cognitoidentityprovider.GetUICustomizationInput{
			UserPoolId: aws.String(d.Id()),
		}
	} else {
		params = &cognitoidentityprovider.GetUICustomizationInput{
			ClientId:   aws.String(d.Id()),
			UserPoolId: aws.String(d.Get("user_pool_id").(string)),
		}
	}

	return conn.GetUICustomization(params)
}
