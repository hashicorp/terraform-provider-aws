// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_secretsmanager_secret_version", name="Secret Version")
func dataSourceSecretVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretVersionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_binary": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secret_string": {
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
				Default:  secretVersionStageCurrent,
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
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID := d.Get("secret_id").(string)
	var version string
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("version_id"); ok {
		versionID := v.(string)
		input.VersionId = aws.String(versionID)
		version = versionID
	} else if v, ok := d.GetOk("version_stage"); ok {
		versionStage := v.(string)
		input.VersionStage = aws.String(versionStage)
		version = versionStage
	}

	id := secretVersionCreateResourceID(secretID, version)
	output, err := findSecretVersion(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrARN, output.ARN)
	d.Set(names.AttrCreatedDate, aws.String(output.CreatedDate.Format(time.RFC3339)))
	d.Set("secret_id", secretID)
	d.Set("secret_binary", string(output.SecretBinary))
	d.Set("secret_string", output.SecretString)
	d.Set("version_id", output.VersionId)
	d.Set("version_stages", output.VersionStages)

	return diags
}
