package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceTransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayMulticastDomainRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"transit_gateway_attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"auto_accept_shared_associations": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"igmpv2_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"static_sources_support": {
				Type:     schema.TypeString,
				Computed: true,
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
			"transit_gateway_multicast_domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayMulticastDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{}

	if v, ok := d.GetOk("transit_gateway_multicast_domain_id"); ok {
		input.TransitGatewayMulticastDomainIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	transitGatewayMulticastDomain, err := FindTransitGatewayMulticastDomain(conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Transit Gateway Multicast Domain", err))
	}

	d.SetId(aws.StringValue(transitGatewayMulticastDomain.TransitGatewayMulticastDomainId))
	d.Set("arn", transitGatewayMulticastDomain.TransitGatewayMulticastDomainArn)
	d.Set("auto_accept_shared_associations", transitGatewayMulticastDomain.Options.AutoAcceptSharedAssociations)
	d.Set("igmpv2_support", transitGatewayMulticastDomain.Options.Igmpv2Support)
	d.Set("owner_id", transitGatewayMulticastDomain.OwnerId)
	d.Set("state", transitGatewayMulticastDomain.State)
	d.Set("static_sources_support", transitGatewayMulticastDomain.Options.StaticSourcesSupport)
	d.Set("transit_gateway_id", transitGatewayMulticastDomain.TransitGatewayId)
	d.Set("transit_gateway_multicast_domain_id", transitGatewayMulticastDomain.TransitGatewayMulticastDomainId)

	if err := d.Set("tags", KeyValueTags(transitGatewayMulticastDomain.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	associations, err := FindTransitGatewayMulticastDomainAssociations(conn, &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("error listing EC2 Transit Gateway Multicast Domain Associations (%s): %s", d.Id(), err)
	}

	if err := d.Set("associations", flattenTransitGatewayMulticastDomainAssociations(associations)); err != nil {
		return diag.Errorf("error setting associations: %s", err)
	}

	members, err := FindTransitGatewayMulticastGroups(conn, &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"is-group-member": "true",
			"is-group-source": "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("error listing EC2 Transit Gateway Multicast Group Members (%s): %s", d.Id(), err)
	}

	if err := d.Set("members", flattenTransitGatewayMulticastGroups(members)); err != nil {
		return diag.Errorf("error setting members: %s", err)
	}

	sources, err := FindTransitGatewayMulticastGroups(conn, &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"is-group-member": "false",
			"is-group-source": "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("error listing EC2 Transit Gateway Multicast Group Members (%s): %s", d.Id(), err)
	}

	if err := d.Set("sources", flattenTransitGatewayMulticastGroups(sources)); err != nil {
		return diag.Errorf("error setting sources: %s", err)
	}

	return nil
}

func flattenTransitGatewayMulticastDomainAssociation(apiObject *ec2.TransitGatewayMulticastDomainAssociation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Subnet.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}

	if v := apiObject.TransitGatewayAttachmentId; v != nil {
		tfMap["transit_gateway_attachment_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTransitGatewayMulticastDomainAssociations(apiObjects []*ec2.TransitGatewayMulticastDomainAssociation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTransitGatewayMulticastDomainAssociation(apiObject))
	}

	return tfList
}

func flattenTransitGatewayMulticastGroup(apiObject *ec2.TransitGatewayMulticastGroup) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GroupIpAddress; v != nil {
		tfMap["group_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTransitGatewayMulticastGroups(apiObjects []*ec2.TransitGatewayMulticastGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTransitGatewayMulticastGroup(apiObject))
	}

	return tfList
}
