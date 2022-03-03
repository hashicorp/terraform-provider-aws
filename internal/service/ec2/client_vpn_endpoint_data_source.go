package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceClientVPNEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClientVPNEndpointRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
						"type": {
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
						"enabled": {
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
						"enabled": {
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
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": DataSourceFiltersSchema(),
			"security_group_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"self_service_portal": {
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
			"tags": tftags.TagsSchemaComputed(),
			"transport_protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
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

func dataSourceClientVPNEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeClientVpnEndpointsInput{}

	if v, ok := d.GetOk("client_vpn_endpoint_id"); ok {
		input.ClientVpnEndpointIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	ep, err := FindClientVPNEndpoint(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Client VPN Endpoint", err)
	}

	d.SetId(aws.StringValue(ep.ClientVpnEndpointId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("client-vpn-endpoint/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if err := d.Set("authentication_options", flattenClientVpnAuthentications(ep.AuthenticationOptions)); err != nil {
		return fmt.Errorf("error setting authentication_options: %w", err)
	}
	d.Set("client_cidr_block", ep.ClientCidrBlock)
	if ep.ClientConnectOptions != nil {
		if err := d.Set("client_connect_options", []interface{}{flattenClientConnectResponseOptions(ep.ClientConnectOptions)}); err != nil {
			return fmt.Errorf("error setting client_connect_options: %w", err)
		}
	} else {
		d.Set("client_connect_options", nil)
	}
	if ep.ClientLoginBannerOptions != nil {
		if err := d.Set("client_login_banner_options", []interface{}{flattenClientLoginBannerResponseOptions(ep.ClientLoginBannerOptions)}); err != nil {
			return fmt.Errorf("error setting client_login_banner_options: %w", err)
		}
	} else {
		d.Set("client_login_banner_options", nil)
	}
	d.Set("client_vpn_endpoint_id", ep.ClientVpnEndpointId)
	if ep.ConnectionLogOptions != nil {
		if err := d.Set("connection_log_options", []interface{}{flattenConnectionLogResponseOptions(ep.ConnectionLogOptions)}); err != nil {
			return fmt.Errorf("error setting connection_log_options: %w", err)
		}
	} else {
		d.Set("connection_log_options", nil)
	}
	d.Set("description", ep.Description)
	d.Set("dns_name", ep.DnsName)
	d.Set("dns_servers", aws.StringValueSlice(ep.DnsServers))
	d.Set("security_group_ids", aws.StringValueSlice(ep.SecurityGroupIds))
	if aws.StringValue(ep.SelfServicePortalUrl) != "" {
		d.Set("self_service_portal", ec2.SelfServicePortalEnabled)
	} else {
		d.Set("self_service_portal", ec2.SelfServicePortalDisabled)
	}
	d.Set("server_certificate_arn", ep.ServerCertificateArn)
	d.Set("session_timeout_hours", ep.SessionTimeoutHours)
	d.Set("split_tunnel", ep.SplitTunnel)
	d.Set("transport_protocol", ep.TransportProtocol)
	d.Set("vpc_id", ep.VpcId)
	d.Set("vpn_port", ep.VpnPort)

	if err := d.Set("tags", KeyValueTags(ep.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
