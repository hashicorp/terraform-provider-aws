package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceCodeSigningConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCodeSigningConfigRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"allowed_publishers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"signing_profile_version_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"untrusted_artifact_on_deployment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCodeSigningConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn()

	arn := d.Get("arn").(string)

	configOutput, err := conn.GetCodeSigningConfigWithContext(ctx, &lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(arn),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Lambda code signing config (%s): %s", arn, err)
	}

	if configOutput == nil {
		return sdkdiag.AppendErrorf(diags, "getting Lambda code signing config (%s): empty response", arn)
	}

	codeSigningConfig := configOutput.CodeSigningConfig
	if codeSigningConfig == nil {
		return sdkdiag.AppendErrorf(diags, "getting Lambda code signing config (%s): empty CodeSigningConfig", arn)
	}

	if err := d.Set("config_id", codeSigningConfig.CodeSigningConfigId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda code signing config id: %s", err)
	}

	if err := d.Set("description", codeSigningConfig.Description); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda code signing config description: %s", err)
	}

	if err := d.Set("last_modified", codeSigningConfig.LastModified); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda code signing config last modified: %s", err)
	}

	if err := d.Set("allowed_publishers", flattenCodeSigningConfigAllowedPublishers(codeSigningConfig.AllowedPublishers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda code signing config allowed publishers: %s", err)
	}

	if err := d.Set("policies", []interface{}{
		map[string]interface{}{
			"untrusted_artifact_on_deployment": codeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment,
		},
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda code signing config code signing policies: %s", err)
	}

	d.SetId(aws.StringValue(codeSigningConfig.CodeSigningConfigArn))

	return diags
}
