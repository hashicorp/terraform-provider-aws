package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayCreate,
		Read:   resourceTransitGatewayRead,
		Update: resourceTransitGatewayUpdate,
		Delete: resourceTransitGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("default_route_table_association", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource
				return old.(string) == ec2.DefaultRouteTableAssociationValueDisable && new.(string) == ec2.DefaultRouteTableAssociationValueEnable
			}),
			customdiff.ForceNewIfChange("default_route_table_propagation", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource
				return old.(string) == ec2.DefaultRouteTablePropagationValueDisable && new.(string) == ec2.DefaultRouteTablePropagationValueEnable
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  64512,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_accept_shared_attachments": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.AutoAcceptSharedAttachmentsValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.AutoAcceptSharedAttachmentsValue_Values(), false),
			},
			"default_route_table_association": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.DefaultRouteTableAssociationValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.DefaultRouteTableAssociationValue_Values(), false),
			},
			"default_route_table_propagation": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.DefaultRouteTablePropagationValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.DefaultRouteTablePropagationValue_Values(), false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dns_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.DnsSupportValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.DnsSupportValue_Values(), false),
			},
			"multicast_support": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.MulticastSupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.MulticastSupportValue_Values(), false),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"propagation_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpn_ecmp_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.VpnEcmpSupportValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.VpnEcmpSupportValue_Values(), false),
			},
		},
	}
}

func resourceTransitGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTransitGatewayInput{
		Options: &ec2.TransitGatewayRequestOptions{
			AutoAcceptSharedAttachments:  aws.String(d.Get("auto_accept_shared_attachments").(string)),
			DefaultRouteTableAssociation: aws.String(d.Get("default_route_table_association").(string)),
			DefaultRouteTablePropagation: aws.String(d.Get("default_route_table_propagation").(string)),
			DnsSupport:                   aws.String(d.Get("dns_support").(string)),
			MulticastSupport:             aws.String(d.Get("multicast_support").(string)),
			VpnEcmpSupport:               aws.String(d.Get("vpn_ecmp_support").(string)),
		},
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGateway),
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		input.Options.AmazonSideAsn = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway: %s", input)
	output, err := conn.CreateTransitGateway(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGateway.TransitGatewayId))

	if err := waitForTransitGatewayCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway (%s) availability: %s", d.Id(), err)
	}

	return resourceTransitGatewayRead(d, meta)
}

func resourceTransitGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGateway, err := DescribeTransitGateway(conn, d.Id())

	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway: %s", err)
	}

	if transitGateway == nil {
		log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGateway.State) == ec2.TransitGatewayStateDeleting || aws.StringValue(transitGateway.State) == ec2.TransitGatewayStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGateway.State))
		d.SetId("")
		return nil
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway (%s): missing options", d.Id())
	}

	d.Set("amazon_side_asn", transitGateway.Options.AmazonSideAsn)
	d.Set("arn", transitGateway.TransitGatewayArn)
	d.Set("association_default_route_table_id", transitGateway.Options.AssociationDefaultRouteTableId)
	d.Set("auto_accept_shared_attachments", transitGateway.Options.AutoAcceptSharedAttachments)
	d.Set("default_route_table_association", transitGateway.Options.DefaultRouteTableAssociation)
	d.Set("default_route_table_propagation", transitGateway.Options.DefaultRouteTablePropagation)
	d.Set("description", transitGateway.Description)
	d.Set("dns_support", transitGateway.Options.DnsSupport)
	d.Set("owner_id", transitGateway.OwnerId)
	d.Set("multicast_support", transitGateway.Options.MulticastSupport)
	d.Set("propagation_default_route_table_id", transitGateway.Options.PropagationDefaultRouteTableId)

	tags := KeyValueTags(transitGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("vpn_ecmp_support", transitGateway.Options.VpnEcmpSupport)

	return nil
}

func resourceTransitGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	modifyTransitGatewayInput := &ec2.ModifyTransitGatewayInput{}
	transitGatewayModified := false

	if d.HasChange("description") {
		transitGatewayModified = true
		modifyTransitGatewayInput.Description = aws.String(d.Get("description").(string))
	}

	options := &ec2.ModifyTransitGatewayOptions{}

	if d.HasChange("auto_accept_shared_attachments") {
		transitGatewayModified = true
		options.AutoAcceptSharedAttachments = aws.String(d.Get("auto_accept_shared_attachments").(string))
	}

	if d.HasChange("default_route_table_association") {
		transitGatewayModified = true
		options.DefaultRouteTableAssociation = aws.String(d.Get("default_route_table_association").(string))
	}

	if d.HasChange("default_route_table_propagation") {
		transitGatewayModified = true
		options.DefaultRouteTablePropagation = aws.String(d.Get("default_route_table_propagation").(string))
	}

	if d.HasChange("dns_support") {
		transitGatewayModified = true
		options.DnsSupport = aws.String(d.Get("dns_support").(string))
	}

	if d.HasChange("vpn_ecmp_support") {
		transitGatewayModified = true
		options.VpnEcmpSupport = aws.String(d.Get("vpn_ecmp_support").(string))
	}
	if transitGatewayModified {
		modifyTransitGatewayInput.TransitGatewayId = aws.String(d.Id())
		modifyTransitGatewayInput.Options = options
		if _, err := conn.ModifyTransitGateway(modifyTransitGatewayInput); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway (%s) options: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteTransitGatewayInput{
		TransitGatewayId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway (%s): %s", d.Id(), input)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteTransitGateway(input)

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Transit Gateway Attachments") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted DirectConnect Gateway Attachments") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted VPN Attachments") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Transit Gateway Cross Region Peering Attachments") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteTransitGateway(input)
	}

	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway: %s", err)
	}

	if err := WaitForTransitGatewayDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
func DecodeTransitGatewayRouteID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_DESTINATION", id)
	}

	return parts[0], parts[1], nil
}

func DecodeTransitGatewayRouteTableAssociationID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_tgw-attach-ID", id)
	}

	return parts[0], parts[1], nil
}

func DecodeTransitGatewayRouteTablePropagationID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_tgw-attach-ID", id)
	}

	return parts[0], parts[1], nil
}

func DescribeTransitGateway(conn *ec2.EC2, transitGatewayID string) (*ec2.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: []*string{aws.String(transitGatewayID)},
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway (%s): %s", transitGatewayID, input)
	for {
		output, err := conn.DescribeTransitGateways(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.TransitGateways) == 0 {
			return nil, nil
		}

		for _, transitGateway := range output.TransitGateways {
			if transitGateway == nil {
				continue
			}

			if aws.StringValue(transitGateway.TransitGatewayId) == transitGatewayID {
				return transitGateway, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeTransitGatewayPeeringAttachment(conn *ec2.EC2, transitGatewayAttachmentID string) (*ec2.TransitGatewayPeeringAttachment, error) {
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
		TransitGatewayAttachmentIds: []*string{aws.String(transitGatewayAttachmentID)},
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Peering Attachment (%s): %s", transitGatewayAttachmentID, input)
	for {
		output, err := conn.DescribeTransitGatewayPeeringAttachments(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.TransitGatewayPeeringAttachments) == 0 {
			return nil, nil
		}

		for _, transitGatewayPeeringAttachment := range output.TransitGatewayPeeringAttachments {
			if transitGatewayPeeringAttachment == nil {
				continue
			}

			if aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				return transitGatewayPeeringAttachment, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeTransitGatewayRoute(conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		// As of the time of writing, the EC2 API reference documentation (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SearchTransitGatewayRoutes.html)
		// incorrectly states which filter Names are allowed. The below are example errors:
		// InvalidParameterValue: Value (transit-gateway-route-destination-cidr-block) for parameter Filters is invalid.
		// InvalidParameterValue: Value (transit-gateway-route-type) for parameter Filters is invalid.
		// InvalidParameterValue: Value (destination-cidr-block) for parameter Filters is invalid.
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("type"),
				Values: []*string{aws.String("static")},
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Searching EC2 Transit Gateway Route Table (%s): %s", transitGatewayRouteTableID, input)
	output, err := conn.SearchTransitGatewayRoutes(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, nil
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationCidrBlock), destination) {
			cidrString := verify.CanonicalCIDRBlock(aws.StringValue(route.DestinationCidrBlock))
			route.DestinationCidrBlock = aws.String(cidrString)
			return route, nil
		}
	}

	return nil, nil
}

func DescribeTransitGatewayRouteTable(conn *ec2.EC2, transitGatewayRouteTableID string) (*ec2.TransitGatewayRouteTable, error) {
	input := &ec2.DescribeTransitGatewayRouteTablesInput{
		TransitGatewayRouteTableIds: []*string{aws.String(transitGatewayRouteTableID)},
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Route Table (%s): %s", transitGatewayRouteTableID, input)
	for {
		output, err := conn.DescribeTransitGatewayRouteTables(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.TransitGatewayRouteTables) == 0 {
			return nil, nil
		}

		for _, transitGatewayRouteTable := range output.TransitGatewayRouteTables {
			if transitGatewayRouteTable == nil {
				continue
			}

			if aws.StringValue(transitGatewayRouteTable.TransitGatewayRouteTableId) == transitGatewayRouteTableID {
				return transitGatewayRouteTable, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeTransitGatewayRouteTableAssociation(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTableAssociation, error) {
	if transitGatewayRouteTableID == "" {
		return nil, nil
	}

	input := &ec2.GetTransitGatewayRouteTableAssociationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: []*string{aws.String(transitGatewayAttachmentID)},
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := conn.GetTransitGatewayRouteTableAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Associations) == 0 {
		return nil, nil
	}

	return output.Associations[0], nil
}

func DescribeTransitGatewayVPCAttachment(conn *ec2.EC2, transitGatewayAttachmentID string) (*ec2.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: []*string{aws.String(transitGatewayAttachmentID)},
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, input)
	for {
		output, err := conn.DescribeTransitGatewayVpcAttachments(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.TransitGatewayVpcAttachments) == 0 {
			return nil, nil
		}

		for _, transitGatewayVpcAttachment := range output.TransitGatewayVpcAttachments {
			if transitGatewayVpcAttachment == nil {
				continue
			}

			if aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				return transitGatewayVpcAttachment, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func DescribeTransitGatewayMulticastDomain(conn *ec2.EC2, domainID string) (*ec2.TransitGatewayMulticastDomain, error) {
	if conn == nil || domainID == "" {
		return nil, nil
	}

	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		// Note: one or more filters required
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-multicast-domain-id"),
				Values: []*string{aws.String(domainID)},
			},
		},
		TransitGatewayMulticastDomainIds: []*string{aws.String(domainID)},
	}

	output, err := conn.DescribeTransitGatewayMulticastDomains(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.TransitGatewayMulticastDomains) == 0 {
		return nil, nil
	}

	return output.TransitGatewayMulticastDomains[0], nil
}

func GetTransitGatewayMulticastDomainAssociations(conn *ec2.EC2, domainID string) ([]*ec2.TransitGatewayMulticastDomainAssociation, error) {
	if conn == nil || domainID == "" {
		return nil, nil
	}

	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		TransitGatewayMulticastDomainId: aws.String(domainID),
	}

	var associations []*ec2.TransitGatewayMulticastDomainAssociation
	log.Printf("[DEBUG] Reading EC2 Transit Gateway Multicast Domain (%s) Associations: %s", domainID, input)
	for {
		output, err := conn.GetTransitGatewayMulticastDomainAssociations(input)
		if err != nil {
			return nil, err
		}

		if output == nil {
			return nil, nil
		}

		associations = append(associations, output.MulticastDomainAssociations...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return associations, nil
}

func SearchTransitGatewayMulticastDomainGroups(conn *ec2.EC2, domainID string, filters []*ec2.Filter) ([]*ec2.TransitGatewayMulticastGroup, error) {
	if conn == nil || domainID == "" {
		return nil, nil
	}

	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters:                         filters,
		TransitGatewayMulticastDomainId: aws.String(domainID),
	}

	var groups []*ec2.TransitGatewayMulticastGroup
	log.Printf("[DEBUG] Reading EC2 Transit Gateway Multicast Domain (%s) groups: %s", domainID, input)
	for {
		output, err := conn.SearchTransitGatewayMulticastGroups(input)
		if err != nil {
			return nil, err
		}

		if output == nil {
			return nil, nil
		}

		groups = append(groups, output.MulticastGroups...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return groups, nil
}

func SearchTransitGatewayMulticastDomainGroupsByType(conn *ec2.EC2, domainID string, member bool) ([]*ec2.TransitGatewayMulticastGroup, error) {
	return SearchTransitGatewayMulticastDomainGroups(conn, domainID, SearchTransitGatewayMulticastDomainGroupsTypeFilter(member))
}

func SearchTransitGatewayMulticastDomainGroupsTypeFilter(member bool) []*ec2.Filter {
	var filters []*ec2.Filter
	if member {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("is-group-member"),
			Values: []*string{aws.String("true")},
		})
	} else {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("is-group-source"),
			Values: []*string{aws.String("true")},
		})
	}
	return filters
}

func SearchTransitGatewayMulticastDomainGroupIpFilters(member bool, groupIP string) []*ec2.Filter {
	return append(SearchTransitGatewayMulticastDomainGroupsTypeFilter(member), &ec2.Filter{
		Name:   aws.String("group-ip-address"),
		Values: []*string{aws.String(groupIP)},
	})
}

func transitGatewayPeeringAttachmentRefreshFunc(conn *ec2.EC2, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayPeeringAttachment, err := DescribeTransitGatewayPeeringAttachment(conn, transitGatewayAttachmentID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			return nil, ec2.TransitGatewayAttachmentStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway Peering Attachment (%s): %s", transitGatewayAttachmentID, err)
		}

		if transitGatewayPeeringAttachment == nil {
			return nil, ec2.TransitGatewayAttachmentStateDeleted, nil
		}

		if aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateFailed && transitGatewayPeeringAttachment.Status != nil {
			return transitGatewayPeeringAttachment, aws.StringValue(transitGatewayPeeringAttachment.State), fmt.Errorf("%s: %s", aws.StringValue(transitGatewayPeeringAttachment.Status.Code), aws.StringValue(transitGatewayPeeringAttachment.Status.Message))
		}

		return transitGatewayPeeringAttachment, aws.StringValue(transitGatewayPeeringAttachment.State), nil
	}
}

func transitGatewayRefreshFunc(conn *ec2.EC2, transitGatewayID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGateway, err := DescribeTransitGateway(conn, transitGatewayID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
			return nil, ec2.TransitGatewayStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if transitGateway == nil {
			return nil, ec2.TransitGatewayStateDeleted, nil
		}

		return transitGateway, aws.StringValue(transitGateway.State), nil
	}
}

func transitGatewayRouteTableAssociationRefreshFunc(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayAssociation, err := DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfawserr.ErrMessageContains(err, "InvalidRouteTableID.NotFound", "") {
			return nil, ec2.TransitGatewayRouteTableStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway Route Table (%s) Association for (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}

		if transitGatewayAssociation == nil {
			return nil, ec2.TransitGatewayRouteTableStateDeleted, nil
		}

		return transitGatewayAssociation, aws.StringValue(transitGatewayAssociation.State), nil
	}
}

func transitGatewayRouteTableAssociationUpdate(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string, associate bool) error {
	transitGatewayAssociation, err := DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment Route Table (%s) association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if associate && transitGatewayAssociation == nil {
		input := &ec2.AssociateTransitGatewayRouteTableInput{
			TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}

		if _, err := conn.AssociateTransitGatewayRouteTable(input); err != nil {
			return fmt.Errorf("error associating EC2 Transit Gateway Route Table (%s) association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}

		if err := waitForTransitGatewayRouteTableAssociationCreation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}
	} else if !associate && transitGatewayAssociation != nil {
		input := &ec2.DisassociateTransitGatewayRouteTableInput{
			TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}

		if _, err := conn.DisassociateTransitGatewayRouteTable(input); err != nil {
			return fmt.Errorf("error disassociating EC2 Transit Gateway Route Table (%s) disassociation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}

		if err := waitForTransitGatewayRouteTableAssociationDeletion(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) disassociation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}
	}

	return nil
}

func transitGatewayRouteTablePropagationUpdate(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string, enablePropagation bool) error {
	transitGatewayRouteTablePropagation, err := FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", transitGatewayAttachmentID, transitGatewayRouteTableID, err)
	}

	if enablePropagation && transitGatewayRouteTablePropagation == nil {
		input := &ec2.EnableTransitGatewayRouteTablePropagationInput{
			TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}

		if _, err := conn.EnableTransitGatewayRouteTablePropagation(input); err != nil {
			return fmt.Errorf("error enabling EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", transitGatewayAttachmentID, transitGatewayRouteTableID, err)
		}
	} else if !enablePropagation && transitGatewayRouteTablePropagation != nil {
		input := &ec2.DisableTransitGatewayRouteTablePropagationInput{
			TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}

		if _, err := conn.DisableTransitGatewayRouteTablePropagation(input); err != nil {
			return fmt.Errorf("error disabling EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", transitGatewayAttachmentID, transitGatewayRouteTableID, err)
		}
	}

	return nil
}

func transitGatewayRouteTableRefreshFunc(conn *ec2.EC2, transitGatewayRouteTableID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayRouteTable, err := DescribeTransitGatewayRouteTable(conn, transitGatewayRouteTableID)

		if tfawserr.ErrMessageContains(err, "InvalidRouteTableID.NotFound", "") {
			return nil, ec2.TransitGatewayRouteTableStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway Route Table (%s): %s", transitGatewayRouteTableID, err)
		}

		if transitGatewayRouteTable == nil {
			return nil, ec2.TransitGatewayRouteTableStateDeleted, nil
		}

		return transitGatewayRouteTable, aws.StringValue(transitGatewayRouteTable.State), nil
	}
}

func transitGatewayVPCAttachmentRefreshFunc(conn *ec2.EC2, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayVpcAttachment, err := DescribeTransitGatewayVPCAttachment(conn, transitGatewayAttachmentID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			return nil, ec2.TransitGatewayAttachmentStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
		}

		if transitGatewayVpcAttachment == nil {
			return nil, ec2.TransitGatewayAttachmentStateDeleted, nil
		}

		return transitGatewayVpcAttachment, aws.StringValue(transitGatewayVpcAttachment.State), nil
	}
}

func TransitGatewayMulticastDomainRefreshFunc(conn *ec2.EC2, domainID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		multicastDomain, err := DescribeTransitGatewayMulticastDomain(conn, domainID)
		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
			return nil, ec2.TransitGatewayMulticastDomainStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain (%s): %s", domainID, err)
		}

		if multicastDomain == nil {
			return nil, ec2.TransitGatewayMulticastDomainStateDeleted, nil
		}

		return multicastDomain, aws.StringValue(multicastDomain.State), nil
	}
}

func TransitGatewayMulticastDomainAssociationRefreshFunc(conn *ec2.EC2, domainID string, subnetIDs []*string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		associations, err := GetTransitGatewayMulticastDomainAssociations(conn, domainID)
		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain associations: %s", err)
		}

		subnetStates := make(map[string]string)
		for _, subnetID := range subnetIDs {
			subnetStates[aws.StringValue(subnetID)] = ""
		}

		for _, association := range associations {
			if association == nil {
				continue
			}
			subnet := association.Subnet
			subnetID := aws.StringValue(subnet.SubnetId)
			if _, exists := subnetStates[subnetID]; exists {
				subnetStates[subnetID] = aws.StringValue(subnet.State)
				continue
			}
		}

		for subnetID, subnetState := range subnetStates {
			if subnetState == "" {
				// Not found, mark as functionally disassociated
				subnetStates[subnetID] = ec2.AssociationStatusCodeDisassociated
			}
		}

		log.Printf(
			"[DEBUG] Current EC2 Transit Gateway Multicast Domain (%s) states:\n\t%s", domainID, subnetStates)

		// Note: Since we are potentially associating/disassociating multiple subnets here, we will have this refresh
		// function only return "associated" once all of the subnets are associated or "disassociated" once all
		// disassociated
		// if we encounter anything else, return immediately
		// if we encounter mixed "disassociated" and "associated", raise an error
		compoundState := ""
		for _, state := range subnetStates {
			if compoundState == "" {
				compoundState = state
				continue
			}
			switch state {
			case ec2.AssociationStatusCodeAssociationFailed:
			case ec2.AssociationStatusCodeDisassociating:
			case ec2.AssociationStatusCodeAssociating:
				return associations, state, nil
			case ec2.AssociationStatusCodeDisassociated:
			case ec2.AssociationStatusCodeAssociated:
				if compoundState != state {
					return nil, "", fmt.Errorf("received conflicting association states")
				}
			default:
				return nil, "", fmt.Errorf("unhandled association state: %s", state)
			}
		}

		return associations, compoundState, nil
	}
}

func waitForTransitGatewayCreation(conn *ec2.EC2, transitGatewayID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayStatePending},
		Target:  []string{ec2.TransitGatewayStateAvailable},
		Refresh: transitGatewayRefreshFunc(conn, transitGatewayID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway (%s) availability", transitGatewayID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayDeletion(conn *ec2.EC2, transitGatewayID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayStateAvailable,
			ec2.TransitGatewayStateDeleting,
		},
		Target:         []string{ec2.TransitGatewayStateDeleted},
		Refresh:        transitGatewayRefreshFunc(conn, transitGatewayID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway (%s) deletion", transitGatewayID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForTransitGatewayPeeringAttachmentAcceptance(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAttachmentStatePending,
			ec2.TransitGatewayAttachmentStatePendingAcceptance,
		},
		Target:  []string{ec2.TransitGatewayAttachmentStateAvailable},
		Refresh: transitGatewayPeeringAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Peering Attachment (%s) availability", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForTransitGatewayPeeringAttachmentCreation(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAttachmentStateFailing,
			ec2.TransitGatewayAttachmentStatePending,
			"initiatingRequest", // No ENUM currently exists in the SDK for the state given by AWS
		},
		Target: []string{
			ec2.TransitGatewayAttachmentStateAvailable,
			ec2.TransitGatewayAttachmentStatePendingAcceptance,
		},
		Refresh: transitGatewayPeeringAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Peering Attachment (%s) availability", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayPeeringAttachmentDeletion(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAttachmentStateAvailable,
			ec2.TransitGatewayAttachmentStateDeleting,
			ec2.TransitGatewayAttachmentStatePendingAcceptance,
			ec2.TransitGatewayAttachmentStateRejected,
		},
		Target:  []string{ec2.TransitGatewayAttachmentStateDeleted},
		Refresh: transitGatewayPeeringAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Peering Attachment (%s) deletion", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForTransitGatewayRouteTableAssociationCreation(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayAssociationStateAssociating},
		Target:  []string{ec2.TransitGatewayAssociationStateAssociated},
		Refresh: transitGatewayRouteTableAssociationRefreshFunc(conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
		Timeout: 5 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Route Table (%s) association: %s", transitGatewayRouteTableID, transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForTransitGatewayRouteTableAssociationDeletion(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAssociationStateAssociated,
			ec2.TransitGatewayAssociationStateDisassociating,
		},
		Target:         []string{""},
		Refresh:        transitGatewayRouteTableAssociationRefreshFunc(conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
		Timeout:        5 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Route Table (%s) disassociation: %s", transitGatewayRouteTableID, transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForTransitGatewayRouteTableCreation(conn *ec2.EC2, transitGatewayRouteTableID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayRouteTableStatePending},
		Target:  []string{ec2.TransitGatewayRouteTableStateAvailable},
		Refresh: transitGatewayRouteTableRefreshFunc(conn, transitGatewayRouteTableID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Route Table (%s) availability", transitGatewayRouteTableID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForTransitGatewayRouteTableDeletion(conn *ec2.EC2, transitGatewayRouteTableID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayRouteTableStateAvailable,
			ec2.TransitGatewayRouteTableStateDeleting,
		},
		Target:         []string{ec2.TransitGatewayRouteTableStateDeleted},
		Refresh:        transitGatewayRouteTableRefreshFunc(conn, transitGatewayRouteTableID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Route Table (%s) deletion", transitGatewayRouteTableID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForTransitGatewayVPCAttachmentAcceptance(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAttachmentStatePending,
			ec2.TransitGatewayAttachmentStatePendingAcceptance,
		},
		Target:  []string{ec2.TransitGatewayAttachmentStateAvailable},
		Refresh: transitGatewayVPCAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway VPC Attachment (%s) availability", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForTransitGatewayVPCAttachmentCreation(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayAttachmentStatePending},
		Target: []string{
			ec2.TransitGatewayAttachmentStatePendingAcceptance,
			ec2.TransitGatewayAttachmentStateAvailable,
		},
		Refresh: transitGatewayVPCAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway VPC Attachment (%s) availability", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayVPCAttachmentDeletion(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayAttachmentStateAvailable,
			ec2.TransitGatewayAttachmentStateDeleting,
		},
		Target:         []string{ec2.TransitGatewayAttachmentStateDeleted},
		Refresh:        transitGatewayVPCAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway VPC Attachment (%s) deletion", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForTransitGatewayVPCAttachmentUpdate(conn *ec2.EC2, transitGatewayAttachmentID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayAttachmentStateModifying},
		Target:  []string{ec2.TransitGatewayAttachmentStateAvailable},
		Refresh: transitGatewayVPCAttachmentRefreshFunc(conn, transitGatewayAttachmentID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway VPC Attachment (%s) availability", transitGatewayAttachmentID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayMulticastDomainCreation(conn *ec2.EC2, domainID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayMulticastDomainStatePending},
		Target:  []string{ec2.TransitGatewayMulticastDomainStateAvailable},
		Refresh: TransitGatewayMulticastDomainRefreshFunc(conn, domainID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Multicast Domain (%s) availability", domainID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayMulticastDomainDeletion(conn *ec2.EC2, domainID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.TransitGatewayMulticastDomainStateAvailable,
			ec2.TransitGatewayMulticastDomainStateDeleting,
		},
		Target:         []string{ec2.TransitGatewayMulticastDomainStateDeleted},
		Refresh:        TransitGatewayMulticastDomainRefreshFunc(conn, domainID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Multicast Domain (%s) deletion", domainID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func WaitForTransitGatewayMulticastDomainAssociation(conn *ec2.EC2, domainID string, subnetIDs []*string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AssociationStatusCodeAssociating},
		Target:  []string{ec2.AssociationStatusCodeAssociated},
		Refresh: TransitGatewayMulticastDomainAssociationRefreshFunc(conn, domainID, subnetIDs),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Multicast Domain associations")
	_, err := stateConf.WaitForState()

	return err
}

func WaitForTransitGatewayMulticastDomainDisassociation(conn *ec2.EC2, domainID string, subnetIDs []*string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.AssociationStatusCodeAssociated,
			ec2.AssociationStatusCodeDisassociating,
		},
		Target:         []string{ec2.AssociationStatusCodeDisassociated},
		Refresh:        TransitGatewayMulticastDomainAssociationRefreshFunc(conn, domainID, subnetIDs),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for EC2 Transit Gateway Multicast Domain dissasociation(s)")
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func WaitForTransitGatewayMulticastDomainGroupRegister(conn *ec2.EC2, domainID string, groupData map[string]interface{}, member bool) error {
	filters := SearchTransitGatewayMulticastDomainGroupIpFilters(member, groupData["group_ip_address"].(string))
	netIDs := groupData["network_interface_ids"].(*schema.Set)

	log.Printf(
		"[DEBUG] Validating EC2 Transit Gateway Multicast Domain (%s) group was registered successfully",
		domainID)

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		groups, err := SearchTransitGatewayMulticastDomainGroups(conn, domainID, filters)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		// find each net ID for this group
		for _, netID := range netIDs.List() {
			found := false
			for _, group := range groups {
				if aws.StringValue(group.NetworkInterfaceId) == netID {
					found = true
					break
				}
			}

			if !found {
				return resource.RetryableError(fmt.Errorf(
					"EC2 Transit Gateway Multicast Domain (%s) group not available: %s",
					domainID, groupData))
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		return fmt.Errorf(
			"error validating that EC2 Transit Gateway Multicast Domain (%s) group was successfully "+
				"registered: %s", domainID, err)
	}

	return nil
}

func WaitForTransitGatewayMulticastDomainGroupDeregister(conn *ec2.EC2, domainID string, groupData map[string]interface{}, member bool) error {
	filters := SearchTransitGatewayMulticastDomainGroupIpFilters(member, groupData["group_ip_address"].(string))
	netIDs := groupData["network_interface_ids"].(*schema.Set)

	log.Printf(
		"[DEBUG] Validating EC2 Transit Gateway Multicast Domain (%s) group was deregistered successfully",
		domainID)

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		groups, err := SearchTransitGatewayMulticastDomainGroups(conn, domainID, filters)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		// make sure no net IDs from this group are found
		for _, netID := range netIDs.List() {
			for _, group := range groups {
				if aws.StringValue(group.NetworkInterfaceId) == netID {
					return resource.RetryableError(
						fmt.Errorf("EC2 Transit Gateway Multicast Domain (%s) still available: %s",
							domainID, groupData))
				}
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		return fmt.Errorf(
			"error validating that EC2 Transit Gateway Multicast Domain (%s) group was successfully "+
				"deregistered: %s", domainID, err)
	}

	return nil
}
