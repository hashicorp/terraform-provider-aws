// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qbusiness_application", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {

	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceAppUpdate,
		DeleteWithoutTimeout: resourceAppDelete,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the Amazon Q application.",
			},
			"arn": {
				Type:        schema.TypeString,
				Description: "The Amazon Resource Name (ARN) of the Amazon Q application.",
				Computed:    true,
			},
			"attachments_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachments_control_mode": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Status information about whether file upload functionality is activated or deactivated for your end user.",
							ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
						},
					},
				},
			},
			"client_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A token that you provide to identify the request to create your Amazon Q application.",
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"encryption_configuration": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The identifier of the AWS KMS key that is used to encrypt your data. Amazon Q doesn't support asymmetric keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Amazon Resource Name (ARN) of an IAM role with permissions to access your Amazon CloudWatch logs and metrics.",
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	iam_service_role_arn := d.Get("iam_service_role_arn").(string)
	display_name := d.Get("display_name").(string)

	input := &qbusiness.CreateApplicationInput{
		RoleArn:     aws.String(iam_service_role_arn),
		DisplayName: aws.String(display_name),
	}

	if v, ok := d.GetOk("attachments_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AttachmentsConfiguration = &qbusiness.AttachmentsConfiguration{
			AttachmentsControlMode: aws.String(v.([]interface{})[0].(map[string]interface{})["attachments_control_mode"].(string)),
		}
	}

	if v, ok := d.GetOk("client_token"); ok {
		input.ClientToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionConfiguration = &qbusiness.EncryptionConfiguration{
			KmsKeyId: aws.String(v.([]interface{})[0].(map[string]interface{})["kms_key_id"].(string)),
		}
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "error creating qbusiness application: %s", err)
	}

	d.SetId(aws.StringValue(output.ApplicationArn))
	d.Set("application_id", aws.StringValue(output.ApplicationId))

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
