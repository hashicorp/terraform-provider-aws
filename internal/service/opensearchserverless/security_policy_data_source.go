// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_opensearchserverless_security_policy", name="Security Policy")
func dataSourceSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrCreatedDate: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the security policy was created.",
			},
			names.AttrDescription: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the security policy.",
			},
			"last_modified_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the security policy was last modified.",
			},
			names.AttrName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the policy.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 32),
					validation.StringMatch(regexache.MustCompile(`^[a-z][0-9a-z-]+$`), `must start with any lower case letter and can include any lower case letter, number, or "-"`),
				),
			},
			names.AttrPolicy: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The JSON policy document without any whitespaces.",
			},
			"policy_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version of the policy.",
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Type of security policy. One of `encryption` or `network`.",
				ValidateDiagFunc: enum.Validate[types.SecurityPolicyType](),
			},
		},
	}
}

func dataSourceSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	name := d.Get(names.AttrName).(string)
	securityPolicy, err := findSecurityPolicyByNameAndType(ctx, conn, name, d.Get(names.AttrType).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Serverless Security Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(securityPolicy.Name))
	d.Set(names.AttrCreatedDate, flex.Int64ToRFC3339StringValue(securityPolicy.CreatedDate))
	d.Set(names.AttrDescription, securityPolicy.Description)
	d.Set("last_modified_date", flex.Int64ToRFC3339StringValue(securityPolicy.LastModifiedDate))
	d.Set(names.AttrName, securityPolicy.Name)
	if securityPolicy.Policy != nil {
		v, err := tfsmithy.DocumentToJSONString(securityPolicy.Policy)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, v)
	} else {
		d.Set(names.AttrPolicy, nil)
	}
	d.Set("policy_version", securityPolicy.PolicyVersion)
	d.Set(names.AttrType, securityPolicy.Type)

	return diags
}
