// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_transfer_server")
func DataSourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_type": {
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
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		ReadWithoutTimeout: dataSourceServerRead,
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	serverID := d.Get("server_id").(string)

	output, err := FindServerByID(ctx, conn, serverID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Server (%s): %s", serverID, err)
	}

	d.SetId(aws.StringValue(output.ServerId))
	d.Set("arn", output.Arn)
	d.Set("certificate", output.Certificate)
	d.Set("domain", output.Domain)
	d.Set("endpoint", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", serverID)))
	d.Set("endpoint_type", output.EndpointType)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("protocols", aws.StringValueSlice(output.Protocols))
	d.Set("security_policy_name", output.SecurityPolicyName)
	d.Set("structured_log_destinations", aws.StringValueSlice(output.StructuredLogDestinations))
	if output.IdentityProviderDetails != nil {
		d.Set("url", output.IdentityProviderDetails.Url)
	} else {
		d.Set("url", "")
	}

	return diags
}
