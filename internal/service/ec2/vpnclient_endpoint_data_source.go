// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_client_vpn_endpoint", name="Client VPN Endpoint")
// @Tags
// @Testing(tagsTest=false)
func dataSourceClientVPNEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClientVPNEndpointRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_directory_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"root_certificate_chain_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"saml_provider_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"self_service_saml_provider_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"client_cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_connect_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"lambda_function_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"client_login_banner_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"banner_text": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"connection_log_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloudwatch_log_stream": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"self_service_portal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"self_service_portal_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_timeout_hours": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"split_tunnel": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"transport_protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceClientVPNEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeClientVpnEndpointsInput{}

	if v, ok := d.GetOk("client_vpn_endpoint_id"); ok {
		input.ClientVpnEndpointIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	ep, err := findClientVPNEndpoint(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Client VPN Endpoint", err))
	}

	d.SetId(aws.ToString(ep.ClientVpnEndpointId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("client-vpn-endpoint/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if err := d.Set("authentication_options", flattenClientVPNAuthentications(ep.AuthenticationOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_options: %s", err)
	}
	d.Set("client_cidr_block", ep.ClientCidrBlock)
	if ep.ClientConnectOptions != nil {
		if err := d.Set("client_connect_options", []interface{}{flattenClientConnectResponseOptions(ep.ClientConnectOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting client_connect_options: %s", err)
		}
	} else {
		d.Set("client_connect_options", nil)
	}
	if ep.ClientLoginBannerOptions != nil {
		if err := d.Set("client_login_banner_options", []interface{}{flattenClientLoginBannerResponseOptions(ep.ClientLoginBannerOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting client_login_banner_options: %s", err)
		}
	} else {
		d.Set("client_login_banner_options", nil)
	}
	d.Set("client_vpn_endpoint_id", ep.ClientVpnEndpointId)
	if ep.ConnectionLogOptions != nil {
		if err := d.Set("connection_log_options", []interface{}{flattenConnectionLogResponseOptions(ep.ConnectionLogOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting connection_log_options: %s", err)
		}
	} else {
		d.Set("connection_log_options", nil)
	}
	d.Set(names.AttrDescription, ep.Description)
	d.Set(names.AttrDNSName, ep.DnsName)
	d.Set("dns_servers", aws.StringSlice(ep.DnsServers))
	d.Set(names.AttrSecurityGroupIDs, aws.StringSlice(ep.SecurityGroupIds))
	if aws.ToString(ep.SelfServicePortalUrl) != "" {
		d.Set("self_service_portal", awstypes.SelfServicePortalEnabled)
	} else {
		d.Set("self_service_portal", awstypes.SelfServicePortalDisabled)
	}
	d.Set("self_service_portal_url", ep.SelfServicePortalUrl)
	d.Set("server_certificate_arn", ep.ServerCertificateArn)
	d.Set("session_timeout_hours", ep.SessionTimeoutHours)
	d.Set("split_tunnel", ep.SplitTunnel)
	d.Set("transport_protocol", ep.TransportProtocol)
	d.Set(names.AttrVPCID, ep.VpcId)
	d.Set("vpn_port", ep.VpnPort)

	setTagsOut(ctx, ep.Tags)

	return diags
}
