package ec2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("default_route_table_association", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource.
				return old.(string) == ec2.DefaultRouteTableAssociationValueDisable && new.(string) == ec2.DefaultRouteTableAssociationValueEnable
			}),
			customdiff.ForceNewIfChange("default_route_table_propagation", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource.
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
			"transit_gateway_cidr_blocks": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: verify.IsIPv4CIDRBlockOrIPv6CIDRBlock(
						validation.All(
							validation.IsCIDRNetwork(0, 24),
							validation.StringDoesNotMatch(regexp.MustCompile(`^169\.254\.`), "must not be from range 169.254.0.0/16"),
						),
						validation.IsCIDRNetwork(0, 64),
					),
				},
			},
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
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGateway),
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		input.Options.AmazonSideAsn = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_cidr_blocks"); ok && v.(*schema.Set).Len() > 0 {
		input.Options.TransitGatewayCidrBlocks = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway: %s", input)
	output, err := conn.CreateTransitGateway(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.TransitGateway.TransitGatewayId))

	if _, err := WaitTransitGatewayCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayRead(d, meta)
}

func resourceTransitGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGateway, err := FindTransitGatewayByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", d.Id(), err)
	}

	d.Set("amazon_side_asn", transitGateway.Options.AmazonSideAsn)
	d.Set("arn", transitGateway.TransitGatewayArn)
	d.Set("association_default_route_table_id", transitGateway.Options.AssociationDefaultRouteTableId)
	d.Set("auto_accept_shared_attachments", transitGateway.Options.AutoAcceptSharedAttachments)
	d.Set("default_route_table_association", transitGateway.Options.DefaultRouteTableAssociation)
	d.Set("default_route_table_propagation", transitGateway.Options.DefaultRouteTablePropagation)
	d.Set("description", transitGateway.Description)
	d.Set("dns_support", transitGateway.Options.DnsSupport)
	d.Set("multicast_support", transitGateway.Options.MulticastSupport)
	d.Set("owner_id", transitGateway.OwnerId)
	d.Set("propagation_default_route_table_id", transitGateway.Options.PropagationDefaultRouteTableId)
	d.Set("transit_gateway_cidr_blocks", aws.StringValueSlice(transitGateway.Options.TransitGatewayCidrBlocks))
	d.Set("vpn_ecmp_support", transitGateway.Options.VpnEcmpSupport)

	tags := KeyValueTags(transitGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyTransitGatewayInput{
			Options:          &ec2.ModifyTransitGatewayOptions{},
			TransitGatewayId: aws.String(d.Id()),
		}

		if d.HasChange("auto_accept_shared_attachments") {
			input.Options.AutoAcceptSharedAttachments = aws.String(d.Get("auto_accept_shared_attachments").(string))
		}

		if d.HasChange("default_route_table_association") {
			input.Options.DefaultRouteTableAssociation = aws.String(d.Get("default_route_table_association").(string))
		}

		if d.HasChange("default_route_table_propagation") {
			input.Options.DefaultRouteTablePropagation = aws.String(d.Get("default_route_table_propagation").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("dns_support") {
			input.Options.DnsSupport = aws.String(d.Get("dns_support").(string))
		}

		if d.HasChange("transit_gateway_cidr_blocks") {
			oRaw, nRaw := d.GetChange("transit_gateway_cidr_blocks")
			o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

			if add := n.Difference(o); add.Len() > 0 {
				input.Options.AddTransitGatewayCidrBlocks = flex.ExpandStringSet(add)
			}

			if del := o.Difference(n); del.Len() > 0 {
				input.Options.RemoveTransitGatewayCidrBlocks = flex.ExpandStringSet(del)
			}
		}

		if d.HasChange("vpn_ecmp_support") {
			input.Options.VpnEcmpSupport = aws.String(d.Get("vpn_ecmp_support").(string))
		}

		if _, err := conn.ModifyTransitGateway(input); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway (%s): %w", d.Id(), err)
		}

		if _, err := WaitTransitGatewayUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for EC2 Transit Gateway (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(TransitGatewayIncorrectStateTimeout, func() (interface{}, error) {
		return conn.DeleteTransitGateway(&ec2.DeleteTransitGatewayInput{
			TransitGatewayId: aws.String(d.Id()),
		})
	}, errCodeIncorrectState)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway (%s) delete: %w", d.Id(), err)
	}

	return nil
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

func transitGatewayRouteTableAssociationRefreshFunc(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayAssociation, err := DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
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
		// deassociation must be done only on already associated state
		if err := waitForTransitGatewayRouteTableAssociationCreation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) association before deletion (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
		}

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
	transitGatewayRouteTablePropagation, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)
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
