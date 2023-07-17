// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_secretsmanager_secret_rotation")
func ResourceSecretRotation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretRotationCreate,
		ReadWithoutTimeout:   resourceSecretRotationRead,
		UpdateWithoutTimeout: resourceSecretRotationUpdate,
		DeleteWithoutTimeout: resourceSecretRotationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:     schema.TypeString,
				Required: true,
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
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								_, exists := d.GetOk("rotation_rules.0.schedule_expression")
								return exists
							},
							DiffSuppressOnRefresh: true,
							ValidateFunc:          validation.IntBetween(1, 1000),
						},
						"duration": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`[0-9h]+`), ""),
						},
						"schedule_expression": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"rotation_rules.0.automatically_after_days"},
							ExactlyOneOf:  []string{"rotation_rules.0.automatically_after_days", "rotation_rules.0.schedule_expression"},
							ValidateFunc:  validation.StringMatch(regexp.MustCompile(`[0-9A-Za-z\(\)#\?\*\-\/, ]+`), ""),
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
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)
	secretID := d.Get("secret_id").(string)

	if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
		input := &secretsmanager.RotateSecretInput{
			RotationLambdaARN: aws.String(v.(string)),
			RotationRules:     expandRotationRules(d.Get("rotation_rules").([]interface{})),
			SecretId:          aws.String(secretID),
		}

		// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
		outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
			return conn.RotateSecretWithContext(ctx, input)
		}, "AccessDeniedException")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Secrets Manager Secret Rotation (%s): %s", secretID, err)
		}

		d.SetId(aws.StringValue(outputRaw.(*secretsmanager.RotateSecretOutput).ARN))
	}

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return FindSecretByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Secrets Manager Secret Rotation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*secretsmanager.DescribeSecretOutput)

	d.Set("secret_id", d.Id())
	d.Set("rotation_enabled", output.RotationEnabled)
	if aws.BoolValue(output.RotationEnabled) {
		d.Set("rotation_lambda_arn", output.RotationLambdaARN)
		if err := d.Set("rotation_rules", flattenRotationRules(output.RotationRules)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting rotation_rules: %s", err)
		}
	} else {
		d.Set("rotation_lambda_arn", "")
		d.Set("rotation_rules", []interface{}{})
	}

	return diags
}

func resourceSecretRotationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)
	secretID := d.Get("secret_id").(string)

	if d.HasChanges("rotation_lambda_arn", "rotation_rules") {
		if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
			input := &secretsmanager.RotateSecretInput{
				RotationLambdaARN: aws.String(v.(string)),
				RotationRules:     expandRotationRules(d.Get("rotation_rules").([]interface{})),
				SecretId:          aws.String(secretID),
			}

			// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
				return conn.RotateSecretWithContext(ctx, input)
			}, "AccessDeniedException")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
			}
		} else {
			input := &secretsmanager.CancelRotateSecretInput{
				SecretId: aws.String(d.Id()),
			}

			_, err := conn.CancelRotateSecretWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "cancelling Secrets Manager Secret Rotation (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceSecretRotationRead(ctx, d, meta)...)
}

func resourceSecretRotationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	_, err := conn.CancelRotateSecretWithContext(ctx, &secretsmanager.CancelRotateSecretInput{
		SecretId: aws.String(d.Get("secret_id").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Secret Manager Secret Rotation (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRotationRules(l []interface{}) *secretsmanager.RotationRulesType {
	if len(l) == 0 {
		return nil
	}
	rules := &secretsmanager.RotationRulesType{}

	tfMap := l[0].(map[string]interface{})

	if v, ok := tfMap["automatically_after_days"].(int); ok && v != 0 {
		rules.AutomaticallyAfterDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		rules.Duration = aws.String(v)
	}

	if v, ok := tfMap["schedule_expression"].(string); ok && v != "" {
		rules.ScheduleExpression = aws.String(v)
	}

	return rules
}

func flattenRotationRules(rules *secretsmanager.RotationRulesType) []interface{} {
	if rules == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := rules.AutomaticallyAfterDays; v != nil && rules.ScheduleExpression == nil {
		// Only populate automatically_after_days if schedule_expression is not set, otherwise we won't be able to update the resource
		m["automatically_after_days"] = int(aws.Int64Value(v))
	}

	if v := rules.Duration; v != nil {
		m["duration"] = aws.StringValue(v)
	}

	if v := rules.ScheduleExpression; v != nil {
		m["schedule_expression"] = aws.StringValue(v)
	}

	return []interface{}{m}
}
