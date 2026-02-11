// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package secretsmanager

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_secretsmanager_secret_rotation", name="Secret Rotation")
// @ArnIdentity("secret_id")
// @Testing(preIdentityVersion="v6.8.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/secretsmanager;secretsmanager.DescribeSecretOutput")
// @Testing(importIgnore="rotate_immediately")
// @Testing(existsTakesT=false, destroyTakesT=false)
func resourceSecretRotation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretRotationCreate,
		ReadWithoutTimeout:   resourceSecretRotationRead,
		UpdateWithoutTimeout: resourceSecretRotationUpdate,
		DeleteWithoutTimeout: resourceSecretRotationDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    secretRotationResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: secretRotationStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"rotate_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"rotation_rules": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"rotation_rules.0.schedule_expression"},
							ExactlyOneOf:  []string{"rotation_rules.0.automatically_after_days", "rotation_rules.0.schedule_expression"},
							ValidateFunc:  validation.IntBetween(1, 1000),
						},
						names.AttrDuration: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9h]+`), ""),
						},
						names.AttrScheduleExpression: {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"rotation_rules.0.automatically_after_days"},
							ExactlyOneOf:  []string{"rotation_rules.0.automatically_after_days", "rotation_rules.0.schedule_expression"},
							ValidateFunc:  validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z\(\)#\?\*\-\/, ]+`), ""),
						},
					},
				},
			},
			"external_secret_rotation_metadata": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"external_secret_rotation_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSecretRotationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID := d.Get("secret_id").(string)
	input := &secretsmanager.RotateSecretInput{
		ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
		RotateImmediately:  aws.Bool(d.Get("rotate_immediately").(bool)),
		RotationRules:      expandRotationRules(d.Get("rotation_rules").([]any)),
		SecretId:           aws.String(secretID),
	}

	if v, ok := d.GetOk("rotation_lambda_arn"); ok {
		input.RotationLambdaARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("external_secret_rotation_metadata"); ok && len(v.([]any)) > 0 {
		input.ExternalSecretRotationMetadata = expandExternalSecretRotationMetadata(v.([]any))
	}

	if v, ok := d.GetOk("external_secret_rotation_role_arn"); ok {
		input.ExternalSecretRotationRoleArn = aws.String(v.(string))
	}

	// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
	// InvalidRequestException: Secrets Manager is unable to assume role (IAM propagation delay).
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func(ctx context.Context) (any, error) {
		return conn.RotateSecret(ctx, input)
	}, "AccessDeniedException", "InvalidRequestException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Secrets Manager Secret Rotation (%s): %s", secretID, err)
	}

	d.SetId(aws.ToString(outputRaw.(*secretsmanager.RotateSecretOutput).ARN))

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	output, err := findSecretByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Secrets Manager Secret Rotation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
	}

	rotationEnabled := aws.ToBool(output.RotationEnabled)
	d.Set("rotation_enabled", rotationEnabled)
	if rotationEnabled {
		d.Set("rotation_lambda_arn", output.RotationLambdaARN)
		if err := d.Set("rotation_rules", flattenRotationRules(output.RotationRules)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting rotation_rules: %s", err)
		}
		if err := d.Set("external_secret_rotation_metadata", flattenExternalSecretRotationMetadata(output.ExternalSecretRotationMetadata)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting external_secret_rotation_metadata: %s", err)
		}
		d.Set("external_secret_rotation_role_arn", output.ExternalSecretRotationRoleArn)
	} else {
		d.Set("rotation_lambda_arn", "")
		d.Set("rotation_rules", []any{})
		d.Set("external_secret_rotation_metadata", []any{})
		d.Set("external_secret_rotation_role_arn", "")
	}
	d.Set("secret_id", d.Id())

	return diags
}

func resourceSecretRotationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	if d.HasChanges("rotation_lambda_arn", "rotation_rules", "external_secret_rotation_metadata", "external_secret_rotation_role_arn") {
		secretID := d.Get("secret_id").(string)
		input := &secretsmanager.RotateSecretInput{
			ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
			RotateImmediately:  aws.Bool(d.Get("rotate_immediately").(bool)),
			RotationRules:      expandRotationRules(d.Get("rotation_rules").([]any)),
			SecretId:           aws.String(secretID),
		}

		if v, ok := d.GetOk("rotation_lambda_arn"); ok {
			input.RotationLambdaARN = aws.String(v.(string))
		}

		if v, ok := d.GetOk("external_secret_rotation_metadata"); ok && len(v.([]any)) > 0 {
			input.ExternalSecretRotationMetadata = expandExternalSecretRotationMetadata(v.([]any))
		}

		if v, ok := d.GetOk("external_secret_rotation_role_arn"); ok {
			input.ExternalSecretRotationRoleArn = aws.String(v.(string))
		}

		// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func(ctx context.Context) (any, error) {
			return conn.RotateSecret(ctx, input)
		}, "AccessDeniedException", "InvalidRequestException")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Secrets Manager Secret Rotation: %s", d.Id())
	_, err := conn.CancelRotateSecret(ctx, &secretsmanager.CancelRotateSecretInput{
		SecretId: aws.String(d.Get("secret_id").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Secret Manager Secret Rotation (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRotationRules(l []any) *types.RotationRulesType {
	if len(l) == 0 {
		return nil
	}
	rules := &types.RotationRulesType{}

	tfMap := l[0].(map[string]any)

	if v, ok := tfMap[names.AttrScheduleExpression].(string); ok && v != "" {
		rules.ScheduleExpression = aws.String(v)
	} else if v, ok := tfMap["automatically_after_days"].(int); ok && v != 0 {
		rules.AutomaticallyAfterDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap[names.AttrDuration].(string); ok && v != "" {
		rules.Duration = aws.String(v)
	}

	return rules
}

func flattenRotationRules(rules *types.RotationRulesType) []any {
	if rules == nil {
		return nil
	}

	m := map[string]any{}

	// If ScheduleExpression is set, AutomaticallyAfterDays will be the result of AWS calculating the number of days between rotations
	if v := rules.ScheduleExpression; v != nil && aws.ToString(v) != "" {
		m[names.AttrScheduleExpression] = aws.ToString(v)
	} else if v := rules.AutomaticallyAfterDays; v != nil && aws.ToInt64(v) != 0 {
		// Only populate automatically_after_days if schedule_expression is not set, otherwise we won't be able to update the resource
		m["automatically_after_days"] = int(aws.ToInt64(v))
	}

	if v := rules.Duration; v != nil {
		m[names.AttrDuration] = aws.ToString(v)
	}

	return []any{m}
}

func expandExternalSecretRotationMetadata(tfList []any) []types.ExternalSecretRotationMetadataItem {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.ExternalSecretRotationMetadataItem, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.ExternalSecretRotationMetadataItem{}

		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenExternalSecretRotationMetadata(apiObjects []types.ExternalSecretRotationMetadataItem) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.Key; v != nil {
			tfMap[names.AttrKey] = aws.ToString(v)
		}

		if v := apiObject.Value; v != nil {
			tfMap[names.AttrValue] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
