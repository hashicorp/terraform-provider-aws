// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfiltersv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_secretsmanager_secrets", name="Secrets")
func dataSourceSecrets() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretsRead,
		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfiltersv2.Schema(),
			names.AttrNames: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecretsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	input := &secretsmanager.ListSecretsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfiltersv2.New(v.(*schema.Set)).SecretsmanagerFilters()
	}

	var results []types.SecretListEntry

	paginator := secretsmanager.NewListSecretsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Secrets Manager Secrets: %s", err)
		}

		if page != nil {
			results = append(results, page.SecretList...)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, tfslices.ApplyToAll(results, func(v types.SecretListEntry) string { return aws.ToString(v.ARN) }))
	d.Set(names.AttrNames, tfslices.ApplyToAll(results, func(v types.SecretListEntry) string { return aws.ToString(v.Name) }))

	return diags
}
