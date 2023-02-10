package secretsmanager

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceSecretRotation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretRotationRead,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 2048),
				Required:     true,
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rotation_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn()
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretID),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret: %s", input)
	output, err := conn.DescribeSecretWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret: %s", err)
	}

	if output.ARN == nil {
		return sdkdiag.AppendErrorf(diags, "Secrets Manager Secret %q not found", secretID)
	}

	d.SetId(aws.StringValue(output.ARN))
	d.Set("rotation_enabled", output.RotationEnabled)
	d.Set("rotation_lambda_arn", output.RotationLambdaARN)

	if err := d.Set("rotation_rules", flattenRotationRules(output.RotationRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rotation_rules: %s", err)
	}

	return diags
}
