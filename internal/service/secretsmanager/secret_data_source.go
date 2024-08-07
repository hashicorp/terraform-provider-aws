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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_secretsmanager_secret", name="Secret")
// @Tags(identifierAttribute="arn")
func dataSourceSecret() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecretRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrARN, names.AttrName},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_changed_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrARN, names.AttrName},
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	var secretID string
	if v, ok := d.GetOk(names.AttrARN); ok {
		secretID = v.(string)
	} else if v, ok := d.GetOk(names.AttrName); ok {
		secretID = v.(string)
	}

	secret, err := findSecretByID(ctx, conn, secretID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s): %s", secretID, err)
	}

	arn := aws.ToString(secret.ARN)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, aws.String(secret.CreatedDate.Format(time.RFC3339)))
	d.Set(names.AttrDescription, secret.Description)
	d.Set(names.AttrKMSKeyID, secret.KmsKeyId)
	d.Set("last_changed_date", aws.String(secret.LastChangedDate.Format(time.RFC3339)))
	d.Set(names.AttrName, secret.Name)

	policy, err := findSecretPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret (%s) policy: %s", d.Id(), err)
	} else if v := policy.ResourcePolicy; v != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(v))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}

	setTagsOut(ctx, secret.Tags)

	return diags
}
