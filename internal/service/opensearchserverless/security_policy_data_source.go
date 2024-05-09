// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_opensearchserverless_security_policy")
func DataSourceSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 32),
					validation.StringMatch(regexache.MustCompile(`^[a-z][0-9a-z-]+$`), `must start with any lower case letter and can include any lower case letter, number, or "-"`),
				),
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SecurityPolicyType](),
			},
		},
	}
}

func dataSourceSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	securityPolicyName := d.Get(names.AttrName).(string)
	securityPolicyType := d.Get(names.AttrType).(string)
	securityPolicy, err := findSecurityPolicyByNameAndType(ctx, conn, securityPolicyName, securityPolicyType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Security Policy with name (%s) and type (%s): %s", securityPolicyName, securityPolicyType, err)
	}

	policyBytes, err := securityPolicy.Policy.MarshalSmithyDocument()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading JSON policy document for OpenSearch Security Policy with name %s and type %s: %s", securityPolicyName, securityPolicyType, err)
	}

	d.SetId(aws.ToString(securityPolicy.Name))
	d.Set(names.AttrDescription, securityPolicy.Description)
	d.Set(names.AttrName, securityPolicy.Name)
	d.Set(names.AttrPolicy, string(policyBytes))
	d.Set("policy_version", securityPolicy.PolicyVersion)
	d.Set(names.AttrType, securityPolicy.Type)

	createdDate := time.UnixMilli(aws.ToInt64(securityPolicy.CreatedDate))
	d.Set(names.AttrCreatedDate, createdDate.Format(time.RFC3339))

	lastModifiedDate := time.UnixMilli(aws.ToInt64(securityPolicy.LastModifiedDate))
	d.Set("last_modified_date", lastModifiedDate.Format(time.RFC3339))

	return diags
}
