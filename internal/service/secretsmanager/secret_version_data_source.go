package secretsmanager

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceSecretVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretVersionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secret_binary": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_stage": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AWSCURRENT",
			},
			"version_stages": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn()
	secretID := d.Get("secret_id").(string)
	var version string

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("version_id"); ok {
		versionID := v.(string)
		input.VersionId = aws.String(versionID)
		version = versionID
	} else {
		versionStage := d.Get("version_stage").(string)
		input.VersionStage = aws.String(versionStage)
		version = versionStage
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret Version: %s", input)
	output, err := conn.GetSecretValueWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
			return sdkdiag.AppendErrorf(diags, "Secrets Manager Secret %q Version %q not found", secretID, version)
		}
		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			return sdkdiag.AppendErrorf(diags, "Secrets Manager Secret %q Version %q not found", secretID, version)
		}
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", secretID, version))
	d.Set("secret_id", secretID)
	d.Set("secret_string", output.SecretString)
	d.Set("version_id", output.VersionId)
	d.Set("secret_binary", string(output.SecretBinary))
	d.Set("arn", output.ARN)

	if err := d.Set("version_stages", flex.FlattenStringList(output.VersionStages)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting version_stages: %s", err)
	}

	return diags
}
