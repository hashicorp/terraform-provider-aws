package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayRouteTableCreate,
		Read:   resourceTransitGatewayRouteTableRead,
		Update: resourceTransitGatewayRouteTableUpdate,
		Delete: resourceTransitGatewayRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_association_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_propagation_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTransitGatewayRouteTableInput{
		TransitGatewayId:  aws.String(d.Get("transit_gateway_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayRouteTable),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Route Table: %s", input)
	output, err := conn.CreateTransitGatewayRouteTable(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Route Table: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayRouteTable.TransitGatewayRouteTableId))

	if err := waitForTransitGatewayRouteTableCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) availability: %s", d.Id(), err)
	}

	return resourceTransitGatewayRouteTableRead(d, meta)
}

func resourceTransitGatewayRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayRouteTable, err := DescribeTransitGatewayRouteTable(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table: %s", err)
	}

	if transitGatewayRouteTable == nil {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayRouteTable.State) == ec2.TransitGatewayRouteTableStateDeleting || aws.StringValue(transitGatewayRouteTable.State) == ec2.TransitGatewayRouteTableStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayRouteTable.State))
		d.SetId("")
		return nil
	}

	d.Set("default_association_route_table", transitGatewayRouteTable.DefaultAssociationRouteTable)
	d.Set("default_propagation_route_table", transitGatewayRouteTable.DefaultPropagationRouteTable)

	tags := KeyValueTags(transitGatewayRouteTable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("transit_gateway_id", transitGatewayRouteTable.TransitGatewayId)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-route-table/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceTransitGatewayRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Route Table (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteTransitGatewayRouteTableInput{
		TransitGatewayRouteTableId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayRouteTable(input)

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Route Table: %s", err)
	}

	if err := waitForTransitGatewayRouteTableDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
