// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iam_security_token_service_preferences", name="Security Token Service Preferences")
func ResourceSecurityTokenServicePreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		ReadWithoutTimeout:   resourceSecurityTokenServicePreferencesRead,
		UpdateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"global_endpoint_token_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(iam.GlobalEndpointTokenVersion_Values(), false),
			},
		},
	}
}

func resourceSecurityTokenServicePreferencesUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	input := &iam.SetSecurityTokenServicePreferencesInput{
		GlobalEndpointTokenVersion: aws.String(d.Get("global_endpoint_token_version").(string)),
	}

	_, err := conn.SetSecurityTokenServicePreferencesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting IAM Security Token Service Preferences: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceSecurityTokenServicePreferencesRead(ctx, d, meta)...)
}

func resourceSecurityTokenServicePreferencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	output, err := conn.GetAccountSummaryWithContext(ctx, &iam.GetAccountSummaryInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Summary: %s", err)
	}

	d.Set("global_endpoint_token_version", fmt.Sprintf("v%dToken", aws.Int64Value(output.SummaryMap[iam.SummaryKeyTypeGlobalEndpointTokenVersion])))

	return diags
}
