// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SecretRemovedMessage = "This data source has been replaced with the `aws_kms_secrets` data source. Upgrade information is available at: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-2-upgrade.html#data-source-aws_kms_secret"

// @SDKDataSource("aws_kms_secret")
func DataSourceSecret() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: func(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return diag.Errorf(SecretRemovedMessage)
		},

		Schema: map[string]*schema.Schema{
			"secret": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"payload": {
							Type:     schema.TypeString,
							Required: true,
						},
						"context": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"grant_tokens": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}
