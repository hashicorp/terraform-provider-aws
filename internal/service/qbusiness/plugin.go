// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func authConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		Optional:     true,
		MaxItems:     1,
		ExactlyOneOf: []string{"basic_auth_configuration", "oauth2_client_credential_configuration"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"basic_auth": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
								ValidateFunc: verify.ValidARN,
							},
							"secret_arn": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.",
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
			},
		},
	}
}

// @SDKResource("aws_qbusiness_plugin", name="Plugin")
// @Tags(identifierAttribute="arn")
func ResourcePlugin() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourcePluginCreate,
		ReadWithoutTimeout:   resourcePluginRead,
		UpdateWithoutTimeout: resourcePluginUpdate,
		DeleteWithoutTimeout: resourcePluginDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identifier of the Amazon Q application associated with the plugin.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"basic_auth_configuration":               authConfigurationSchema(),
			"oauth2_client_credential_configuration": authConfigurationSchema(),
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Amazon Q plugin.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"server_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source URL used for plugin configuration.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 2048),
					validation.StringMatch(regexache.MustCompile(`^(https?|ftp|file)://([^\s]*)$`), "must be a valid URL"),
				),
			},
			"state": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "State of plugin. Valid value are `ENABLED` and `DISABLED`",
				ValidateFunc: validation.StringInSlice(qbusiness.PluginState_Values(), false),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of plugin. Valid value are `SERVICE_NOW`, `SALESFORCE`, `JIRA`, and `ZENDESK`",
				ValidateFunc: validation.StringInSlice(qbusiness.PluginType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePluginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourcePluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourcePluginUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourcePluginDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func flattenBasicAuthConfiguration(basicAuthConfiguration *qbusiness.BasicAuthConfiguration) []interface{} {
	if basicAuthConfiguration == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"role_arn":   aws.ToString(basicAuthConfiguration.RoleArn),
			"secret_arn": aws.ToString(basicAuthConfiguration.SecretArn),
		},
	}
}

func flattenOAuth2ClientCredentialConfiguration(oauth2ClientCredentialConfiguration *qbusiness.OAuth2ClientCredentialConfiguration) []interface{} {
	if oauth2ClientCredentialConfiguration == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"role_arn":   aws.ToString(oauth2ClientCredentialConfiguration.RoleArn),
			"secret_arn": aws.ToString(oauth2ClientCredentialConfiguration.SecretArn),
		},
	}
}

func expandBasicAuthConfiguration(basicAuthConfiguration []interface{}) *qbusiness.BasicAuthConfiguration {
	if len(basicAuthConfiguration) == 0 {
		return nil
	}

	basicAuth := basicAuthConfiguration[0].(map[string]interface{})

	return &qbusiness.BasicAuthConfiguration{
		RoleArn:   aws.String(basicAuth["role_arn"].(string)),
		SecretArn: aws.String(basicAuth["secret_arn"].(string)),
	}
}

func expandOAuth2ClientCredentialConfiguration(oauth2ClientCredentialConfiguration []interface{}) *qbusiness.OAuth2ClientCredentialConfiguration {
	if len(oauth2ClientCredentialConfiguration) == 0 {
		return nil
	}

	oAuth2 := oauth2ClientCredentialConfiguration[0].(map[string]interface{})

	return &qbusiness.OAuth2ClientCredentialConfiguration{
		RoleArn:   aws.String(oAuth2["role_arn"].(string)),
		SecretArn: aws.String(oAuth2["secret_arn"].(string)),
	}
}
