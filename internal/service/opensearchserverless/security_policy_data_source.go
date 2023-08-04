// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_opensearchserverless_security_policy")
func DataSourceSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityPolicyRead,

		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 32),
					validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9-]+$`), `must start with any lower case letter and can include any lower case letter, number, or "-"`),
				),
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
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

	securityPolicyName := d.Get("name").(string)
	securityPolicyType := d.Get("type").(string)
	securityPolicy, err := findSecurityPolicyByNameAndType(ctx, conn, securityPolicyName, securityPolicyType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Security Policy with name (%s) and type (%s): %s", securityPolicyName, securityPolicyType, err)
	}

	policyBytes, err := securityPolicy.Policy.MarshalSmithyDocument()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading JSON policy document for OpenSearch Security Policy with name %s and type %s: %s", securityPolicyName, securityPolicyType, err)
	}

	d.SetId(aws.ToString(securityPolicy.Name))
	d.Set("description", securityPolicy.Description)
	d.Set("name", securityPolicy.Name)
	d.Set("policy", string(policyBytes))
	d.Set("policy_version", securityPolicy.PolicyVersion)
	d.Set("type", securityPolicy.Type)

	createdDate := time.UnixMilli(aws.ToInt64(securityPolicy.CreatedDate))
	d.Set("created_date", createdDate.Format(time.RFC3339))

	lastModifiedDate := time.UnixMilli(aws.ToInt64(securityPolicy.LastModifiedDate))
	d.Set("last_modified_date", lastModifiedDate.Format(time.RFC3339))

	return diags
}
