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

func ResourceTransitGatewayPolicyTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayPolicyTableCreate,
		Read:   resourceTransitGatewayPolicyTableRead,
		Update: resourceTransitGatewayPolicyTableUpdate,
		Delete: resourceTransitGatewayPolicyTableDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
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

func resourceTransitGatewayPolicyTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	transitGatewayID := d.Get("transit_gateway_id").(string)
	input := &ec2.CreateTransitGatewayPolicyTableInput{
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayPolicyTable),
		TransitGatewayId:  aws.String(transitGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Policy Table: %s", input)
	output, err := conn.CreateTransitGatewayPolicyTable(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway (%s) Policy Table: %w", transitGatewayID, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPolicyTable.TransitGatewayPolicyTableId))

	if _, err := WaitTransitGatewayPolicyTableCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Policy Table (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayPolicyTableRead(d, meta)
}

func resourceTransitGatewayPolicyTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayPolicyTable, err := FindTransitGatewayPolicyTableByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Policy Table (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-policy-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("state", transitGatewayPolicyTable.State)
	d.Set("transit_gateway_id", transitGatewayPolicyTable.TransitGatewayId)

	tags := KeyValueTags(transitGatewayPolicyTable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayPolicyTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway Policy Table (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayPolicyTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table: %s", d.Id())
	_, err := conn.DeleteTransitGatewayPolicyTable(&ec2.DeleteTransitGatewayPolicyTableInput{
		TransitGatewayPolicyTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Policy Table (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPolicyTableDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Policy Table (%s) delete: %w", d.Id(), err)
	}

	return nil
}
