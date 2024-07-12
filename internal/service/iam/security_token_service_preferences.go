// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iam_security_token_service_preferences", name="Security Token Service Preferences")
func resourceSecurityTokenServicePreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		ReadWithoutTimeout:   resourceSecurityTokenServicePreferencesRead,
		UpdateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"global_endpoint_token_version": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.GlobalEndpointTokenVersion](),
			},
		},
	}
}

func resourceSecurityTokenServicePreferencesUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.SetSecurityTokenServicePreferencesInput{
		GlobalEndpointTokenVersion: awstypes.GlobalEndpointTokenVersion(d.Get("global_endpoint_token_version").(string)),
	}

	_, err := conn.SetSecurityTokenServicePreferences(ctx, input)

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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	output, err := conn.GetAccountSummary(ctx, &iam.GetAccountSummaryInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Summary: %s", err)
	}

	d.Set("global_endpoint_token_version", fmt.Sprintf("v%dToken", output.SummaryMap[string(awstypes.SummaryKeyTypeGlobalEndpointTokenVersion)]))

	return diags
}
