// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_secretsmanager_secret_rotation", name="Secret Rotation")
func resourceSecretRotation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretRotationCreate,
		ReadWithoutTimeout:   resourceSecretRotationRead,
		UpdateWithoutTimeout: resourceSecretRotationUpdate,
		DeleteWithoutTimeout: resourceSecretRotationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSecretRotationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID := d.Get("secret_id").(string)
	input := &secretsmanager.RotateSecretInput{
		ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
		RotateImmediately:  aws.Bool(d.Get("rotate_immediately").(bool)),
		RotationRules:      expandRotationRules(d.Get("rotation_rules").([]interface{})),
		SecretId:           aws.String(secretID),
	}

	if v, ok := d.GetOk("rotation_lambda_arn"); ok {
		input.RotationLambdaARN = aws.String(v.(string))
	}

	// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.RotateSecret(ctx, input)
	}, "AccessDeniedException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Secrets Manager Secret Rotation (%s): %s", secretID, err)
	}

	d.SetId(aws.ToString(outputRaw.(*secretsmanager.RotateSecretOutput).ARN))

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	output, err := findSecretByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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
	} else {
		d.Set("rotation_lambda_arn", "")
		d.Set("rotation_rules", []interface{}{})
	}
	d.Set("secret_id", d.Id())

	return diags
}

func resourceSecretRotationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	if d.HasChanges("rotation_lambda_arn", "rotation_rules") {
		secretID := d.Get("secret_id").(string)
		input := &secretsmanager.RotateSecretInput{
			ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
			RotateImmediately:  aws.Bool(d.Get("rotate_immediately").(bool)),
			RotationRules:      expandRotationRules(d.Get("rotation_rules").([]interface{})),
			SecretId:           aws.String(secretID),
		}

		if v, ok := d.GetOk("rotation_lambda_arn"); ok {
			input.RotationLambdaARN = aws.String(v.(string))
		}

		// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
			return conn.RotateSecret(ctx, input)
		}, "AccessDeniedException")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func expandRotationRules(l []interface{}) *types.RotationRulesType {
	if len(l) == 0 {
		return nil
	}
	rules := &types.RotationRulesType{}

	tfMap := l[0].(map[string]interface{})

	if v, ok := tfMap["automatically_after_days"].(int); ok && v != 0 {
		rules.AutomaticallyAfterDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap[names.AttrDuration].(string); ok && v != "" {
		rules.Duration = aws.String(v)
	}

	if v, ok := tfMap[names.AttrScheduleExpression].(string); ok && v != "" {
		rules.ScheduleExpression = aws.String(v)
	}

	return rules
}

func flattenRotationRules(rules *types.RotationRulesType) []interface{} {
	if rules == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := rules.AutomaticallyAfterDays; v != nil && rules.ScheduleExpression == nil {
		// Only populate automatically_after_days if schedule_expression is not set, otherwise we won't be able to update the resource
		m["automatically_after_days"] = int(aws.ToInt64(v))
	}

	if v := rules.Duration; v != nil {
		m[names.AttrDuration] = aws.ToString(v)
	}

	if v := rules.ScheduleExpression; v != nil {
		m[names.AttrScheduleExpression] = aws.ToString(v)
	}

	return []interface{}{m}
}
