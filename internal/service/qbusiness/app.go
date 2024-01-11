// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qbusiness_app", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {

	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceAppUpdate,
		DeleteWithoutTimeout: resourceAppDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
							ValidateFunc: validation.StringInSlice(qbusiness.AttachmentsControlMode_Values(), false),
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
				ForceNew:    true,
				MaxItems:    1,
				Description: "The identifier of the AWS KMS key that is used to encrypt your data. Amazon Q doesn't support asymmetric keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidKMSKeyID,
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

		CustomizeDiff: verify.SetTagsDiff,
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
		return sdkdiag.AppendErrorf(diags, "creating qbusiness application: %s", err)
	}

	d.SetId(aws.StringValue(output.ApplicationId))
	d.Set("arn", aws.StringValue(output.ApplicationArn))

	if _, err := waitApplicationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness application (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	app, err := FindAppByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] qbusiness application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness application: %s", err)
	}

	d.Set("application_id", aws.StringValue(app.ApplicationId))
	d.Set("arn", aws.StringValue(app.ApplicationArn))
	d.Set("description", aws.StringValue(app.Description))
	d.Set("display_name", aws.StringValue(app.DisplayName))
	d.Set("iam_service_role_arn", aws.StringValue(app.RoleArn))

	if v, ok := d.GetOk("attachments_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if err := d.Set("attachments_configuration", flattenAttachmentsConfiguration(app.AttachmentsConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting qbusiness application attachments_configuration: %s", err)
		}
	}

	if v, ok := d.GetOk("encryption_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if err := d.Set("encryption_configuration", flattenEncryptionConfiguration(app.EncryptionConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting qbusiness application encryption_configuration: %s", err)
		}
	}
	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	input := &qbusiness.UpdateApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_nanme").(string))
	}

	if d.HasChange("iam_service_role_arn") {
		input.RoleArn = aws.String(d.Get("iam_service_role_arn").(string))
	}

	if d.HasChange("attachments_configuration") {
		input.AttachmentsConfiguration = &qbusiness.AttachmentsConfiguration{
			AttachmentsControlMode: aws.String(d.Get("attachments_configuration").([]interface{})[0].(map[string]interface{})["attachments_control_mode"].(string)),
		}
	}

	_, err := conn.UpdateApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness application: %s", err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	_, err := conn.DeleteApplicationWithContext(ctx, &qbusiness.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness application (%s): %s", d.Id(), err)
	}

	if _, err := waitApplicationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness app (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindAppByID(ctx context.Context, conn *qbusiness.QBusiness, id string) (*qbusiness.GetApplicationOutput, error) {
	input := &qbusiness.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplicationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func flattenAttachmentsConfiguration(v *qbusiness.AppliedAttachmentsConfiguration) []interface{} {
	if v == nil {
		return nil
	}

	fmt.Printf("\nflattenAttachmentsConfiguration: %v\n", v)

	return []interface{}{
		map[string]interface{}{
			"attachments_control_mode": aws.StringValue(v.AttachmentsControlMode),
		},
	}
}

func flattenEncryptionConfiguration(v *qbusiness.EncryptionConfiguration) []interface{} {
	if v == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"kms_key_id": aws.StringValue(v.KmsKeyId),
		},
	}
}
