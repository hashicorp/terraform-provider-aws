package ec2

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceTransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTransitGatewayMulticastDomainRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transit_gateway_attachment_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainAssociationsHash,
			},
			"auto_accept_shared_associations": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.AutoAcceptSharedAssociationsValueDisable,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AutoAcceptSharedAssociationsValueEnable,
					ec2.AutoAcceptSharedAssociationsValueDisable,
				}, false),
			},
			"filter": DataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"igmpv2_support": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "disable",
			},
			"members": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainGroupsHash,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sources": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainGroupsHash,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"static_source_support": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "disable",
			},
			"tags": tftags.TagsSchemaComputed(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayMulticastDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayMulticastDomainIds = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Multicast Domains: %s", input)
	output, err := conn.DescribeTransitGatewayMulticastDomains(input)

	if err != nil {
		log.Printf("[DEBUG] Reading EC2 Transit Gateway Multicast Domains: %s", input)
		return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domains: %ws", err)
	}

	if output == nil || len(output.TransitGatewayMulticastDomains) == 0 {
		return errors.New("error reading EC2 Transit Gateway Multicast Domains: no results found")
	}

	transitGatewayMulticastDomain := output.TransitGatewayMulticastDomains[0]

	if transitGatewayMulticastDomain == nil {
		return errors.New("error reading EC2 Transit Gateway Multicast Domain: empty result")
	}

	if transitGatewayMulticastDomain.Options == nil {
		return errors.New("error reading EC2 Transit Gateway Multicast Domain: missing options")
	}

	d.Set("transit_gateway_id", transitGatewayMulticastDomain.TransitGatewayId)
	d.Set("owner_id", transitGatewayMulticastDomain.OwnerId)
	d.Set("arn", transitGatewayMulticastDomain.TransitGatewayMulticastDomainArn)

	if err := d.Set("tags", KeyValueTags(transitGatewayMulticastDomain.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("igmpv2_support", transitGatewayMulticastDomain.Options.Igmpv2Support)
	d.Set("static_source_support", transitGatewayMulticastDomain.Options.StaticSourcesSupport)
	d.Set("auto_accept_shared_associations", transitGatewayMulticastDomain.Options.AutoAcceptSharedAssociations)

	d.SetId(aws.StringValue(transitGatewayMulticastDomain.TransitGatewayMulticastDomainId))

	return nil
}
