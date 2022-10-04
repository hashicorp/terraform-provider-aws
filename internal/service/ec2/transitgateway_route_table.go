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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
		return fmt.Errorf("creating EC2 Transit Gateway Route Table: %w", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayRouteTable.TransitGatewayRouteTableId))

	if _, err := WaitTransitGatewayRouteTableCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayRouteTableRead(d, meta)
}

func resourceTransitGatewayRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayRouteTable, err := FindTransitGatewayRouteTableByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Route Table (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-route-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("default_association_route_table", transitGatewayRouteTable.DefaultAssociationRouteTable)
	d.Set("default_propagation_route_table", transitGatewayRouteTable.DefaultPropagationRouteTable)
	d.Set("transit_gateway_id", transitGatewayRouteTable.TransitGatewayId)

	tags := KeyValueTags(transitGatewayRouteTable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway Route Table (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table: %s", d.Id())
	_, err := conn.DeleteTransitGatewayRouteTable(&ec2.DeleteTransitGatewayRouteTableInput{
		TransitGatewayRouteTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Route Table (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayRouteTableDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table (%s) delete: %w", d.Id(), err)
	}

	return nil
}
