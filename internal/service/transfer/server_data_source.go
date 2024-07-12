// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_transfer_server", name="Server")
func dataSourceServer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpointType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invocation_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocols": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"security_policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"structured_log_destinations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID := d.Get("server_id").(string)
	output, err := findServerByID(ctx, conn, serverID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Server (%s): %s", serverID, err)
	}

	d.SetId(aws.ToString(output.ServerId))
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCertificate, output.Certificate)
	d.Set(names.AttrDomain, output.Domain)
	d.Set(names.AttrEndpoint, meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s.server.transfer", serverID)))
	d.Set(names.AttrEndpointType, output.EndpointType)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("protocols", output.Protocols)
	d.Set("security_policy_name", output.SecurityPolicyName)
	d.Set("structured_log_destinations", output.StructuredLogDestinations)
	if output.IdentityProviderDetails != nil {
		d.Set(names.AttrURL, output.IdentityProviderDetails.Url)
	} else {
		d.Set(names.AttrURL, "")
	}

	return diags
}
