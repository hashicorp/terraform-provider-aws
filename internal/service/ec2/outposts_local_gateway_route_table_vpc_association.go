package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Missing constant in AWS Go SDK
	ec2ResourceTypeLocalGatewayRouteTableVpcAssociation = "local-gateway-route-table-vpc-association"
)

func ResourceLocalGatewayRouteTableVPCAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalGatewayRouteTableVPCAssociationCreate,
		Read:   resourceLocalGatewayRouteTableVPCAssociationRead,
		Update: resourceLocalGatewayRouteTableVPCAssociationUpdate,
		Delete: resourceLocalGatewayRouteTableVPCAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"local_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLocalGatewayRouteTableVPCAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &ec2.CreateLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableId: aws.String(d.Get("local_gateway_route_table_id").(string)),
		TagSpecifications:        ec2TagSpecificationsFromKeyValueTags(tags, ec2ResourceTypeLocalGatewayRouteTableVpcAssociation),
		VpcId:                    aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateLocalGatewayRouteTableVpcAssociation(req)

	if err != nil {
		return fmt.Errorf("error creating EC2 Local Gateway Route Table VPC Association: %w", err)
	}

	d.SetId(aws.StringValue(output.LocalGatewayRouteTableVpcAssociation.LocalGatewayRouteTableVpcAssociationId))

	if _, err := WaitLocalGatewayRouteTableVPCAssociationAssociated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Local Gateway Route Table VPC Association (%s) to associate: %w", d.Id(), err)
	}

	return resourceLocalGatewayRouteTableVPCAssociationRead(d, meta)
}

func resourceLocalGatewayRouteTableVPCAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	association, err := GetLocalGatewayRouteTableVPCAssociation(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Route Table VPC Association (%s): %w", d.Id(), err)
	}

	if association == nil {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeAssociated {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) status (%s), removing from state", d.Id(), aws.StringValue(association.State))
		d.SetId("")
		return nil
	}

	d.Set("local_gateway_id", association.LocalGatewayId)
	d.Set("local_gateway_route_table_id", association.LocalGatewayRouteTableId)

	tags := KeyValueTags(association.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("vpc_id", association.VpcId)

	return nil
}

func resourceLocalGatewayRouteTableVPCAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Local Gateway Route Table VPC Association (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLocalGatewayRouteTableVPCAssociationRead(d, meta)
}

func resourceLocalGatewayRouteTableVPCAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableVpcAssociationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteLocalGatewayRouteTableVpcAssociation(input)

	if tfawserr.ErrCodeEquals(err, "InvalidLocalGatewayRouteTableVpcAssociationID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Local Gateway Route Table VPC Association (%s): %w", d.Id(), err)
	}

	if _, err := WaitLocalGatewayRouteTableVPCAssociationDisassociated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Local Gateway Route Table VPC Association (%s) to disassociate: %w", d.Id(), err)
	}

	return nil
}

func GetLocalGatewayRouteTableVPCAssociation(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
		LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
	}

	output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, fmt.Errorf("empty response")
	}

	var association *ec2.LocalGatewayRouteTableVpcAssociation

	for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
		if outputAssociation == nil {
			continue
		}

		if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
			association = outputAssociation
			break
		}
	}

	return association, nil
}
