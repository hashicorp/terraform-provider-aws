// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_secretsmanager_secret", name="Secret")
func dataSourceSecret() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"arn", "name"},
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_changed_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"arn", "name"},
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var secretID string
	if v, ok := d.GetOk("arn"); ok {
		secretID = v.(string)
	} else if v, ok := d.GetOk("name"); ok {
		secretID = v.(string)
	}

	secret, err := findSecretByID(ctx, conn, secretID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s): %s", secretID, err)
	}

	arn := aws.ToString(secret.ARN)
	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("created_date", aws.String(secret.CreatedDate.Format(time.RFC3339)))
	d.Set("description", secret.Description)
	d.Set("kms_key_id", secret.KmsKeyId)
	d.Set("last_changed_date", aws.String(secret.LastChangedDate.Format(time.RFC3339)))
	d.Set("name", secret.Name)

	policy, err := findSecretPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s) policy: %s", d.Id(), err)
	} else if v := policy.ResourcePolicy; v != nil {
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.ToString(v))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("policy", policyToSet)
	} else {
		d.Set("policy", "")
	}

	if err := d.Set("tags", KeyValueTags(ctx, secret.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
