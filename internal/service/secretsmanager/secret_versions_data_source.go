// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_secretsmanager_secret_versions", name="Secret Versions")
func dataSourceSecretVersions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretVersionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_deprecated": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"max_results": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_accessed_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"version_stages": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("include_deprecated"); ok {
		includeDeprecated := v.(bool)
		input.IncludeDeprecated = aws.Bool(includeDeprecated)
	}
	if v, ok := d.GetOk("max_results"); ok {
		maxResults := v.(int32)
		input.MaxResults = aws.Int32(maxResults)
	}

	id := secretID
	output, err := findSecretVersions(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Versions (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("arn", output.ARN)
	var versions []interface{}
	for _, version := range output.Versions {
		versions = append(versions, map[string]interface{}{
			"created_date":       version.CreatedDate,
			"last_accessed_date": version.LastAccessedDate,
			"version_id":         version.VersionId,
			"version_stages":     version.VersionStages,
		})
	}
	d.Set("versions", output.Versions)

	return diags
}
