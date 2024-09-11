// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_code_signing_config", name="Code Signing Config")
func resourceCodeSigningConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCodeSigningConfigCreate,
		ReadWithoutTimeout:   resourceCodeSigningConfigRead,
		UpdateWithoutTimeout: resourceCodeSigningConfigUpdate,
		DeleteWithoutTimeout: resourceCodeSigningConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_publishers": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"signing_profile_version_arns": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 20,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policies": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"untrusted_artifact_on_deployment": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CodeSigningPolicy](),
						},
					},
				},
			},
		},
	}
}

func resourceCodeSigningConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	input := &lambda.CreateCodeSigningConfigInput{
		AllowedPublishers: expandAllowedPublishers(d.Get("allowed_publishers").([]interface{})),
		Description:       aws.String(d.Get(names.AttrDescription).(string)),
	}

	if v, ok := d.GetOk("policies"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input.CodeSigningPolicies = &awstypes.CodeSigningPolicies{
			UntrustedArtifactOnDeployment: awstypes.CodeSigningPolicy(tfMap["untrusted_artifact_on_deployment"].(string)),
		}
	}

	output, err := conn.CreateCodeSigningConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Code Signing Config: %s", err)
	}

	d.SetId(aws.ToString(output.CodeSigningConfig.CodeSigningConfigArn))

	return append(diags, resourceCodeSigningConfigRead(ctx, d, meta)...)
}

func resourceCodeSigningConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	output, err := findCodeSigningConfigByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Code Signing Config %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Code Signing Config (%s): %s", d.Id(), err)
	}

	if err := d.Set("allowed_publishers", flattenAllowedPublishers(output.AllowedPublishers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting allowed_publishers: %s", err)
	}
	d.Set(names.AttrARN, output.CodeSigningConfigArn)
	d.Set("config_id", output.CodeSigningConfigId)
	d.Set(names.AttrDescription, output.Description)
	d.Set("last_modified", output.LastModified)
	if err := d.Set("policies", flattenCodeSigningPolicies(output.CodeSigningPolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policies: %s", err)
	}

	return diags
}

func resourceCodeSigningConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	input := &lambda.UpdateCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(d.Id()),
	}

	if d.HasChange("allowed_publishers") {
		input.AllowedPublishers = expandAllowedPublishers(d.Get("allowed_publishers").([]interface{}))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("policies") {
		if v, ok := d.GetOk("policies"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})
			input.CodeSigningPolicies = &awstypes.CodeSigningPolicies{
				UntrustedArtifactOnDeployment: awstypes.CodeSigningPolicy(tfMap["untrusted_artifact_on_deployment"].(string)),
			}
		}
	}

	_, err := conn.UpdateCodeSigningConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Code Signing Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCodeSigningConfigRead(ctx, d, meta)...)
}

func resourceCodeSigningConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	log.Printf("[INFO] Deleting Lambda Code Signing Config: %s", d.Id())
	_, err := conn.DeleteCodeSigningConfig(ctx, &lambda.DeleteCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Code Signing Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findCodeSigningConfigByARN(ctx context.Context, conn *lambda.Client, arn string) (*awstypes.CodeSigningConfig, error) {
	input := &lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(arn),
	}

	return findCodeSigningConfig(ctx, conn, input)
}

func findCodeSigningConfig(ctx context.Context, conn *lambda.Client, input *lambda.GetCodeSigningConfigInput) (*awstypes.CodeSigningConfig, error) {
	output, err := conn.GetCodeSigningConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CodeSigningConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CodeSigningConfig, nil
}

func expandAllowedPublishers(tfList []interface{}) *awstypes.AllowedPublishers {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &awstypes.AllowedPublishers{
		SigningProfileVersionArns: flex.ExpandStringValueSet(tfMap["signing_profile_version_arns"].(*schema.Set)),
	}
}

func flattenAllowedPublishers(apiObject *awstypes.AllowedPublishers) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"signing_profile_version_arns": apiObject.SigningProfileVersionArns,
	}}
}

func flattenCodeSigningPolicies(apiObject *awstypes.CodeSigningPolicies) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"untrusted_artifact_on_deployment": apiObject.UntrustedArtifactOnDeployment,
	}

	return []interface{}{tfMap}
}
