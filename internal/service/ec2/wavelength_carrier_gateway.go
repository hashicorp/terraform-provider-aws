package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCarrierGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCarrierGatewayCreate,
		Read:   resourceCarrierGatewayRead,
		Update: resourceCarrierGatewayUpdate,
		Delete: resourceCarrierGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceCarrierGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateCarrierGatewayInput{
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, "carrier-gateway"),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Carrier Gateway: %s", input)
	output, err := conn.CreateCarrierGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Carrier Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.CarrierGateway.CarrierGatewayId))

	_, err = WaitCarrierGatewayAvailable(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Carrier Gateway (%s) to become available: %w", d.Id(), err)
	}

	return resourceCarrierGatewayRead(d, meta)
}

func resourceCarrierGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	carrierGateway, err := FindCarrierGatewayByID(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidCarrierGatewayIDNotFound) {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Carrier Gateway (%s): %w", d.Id(), err)
	}

	if carrierGateway == nil || aws.StringValue(carrierGateway.State) == ec2.CarrierGatewayStateDeleted {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(carrierGateway.OwnerId),
		Resource:  fmt.Sprintf("carrier-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", carrierGateway.OwnerId)
	d.Set("vpc_id", carrierGateway.VpcId)

	tags := KeyValueTags(carrierGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCarrierGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Carrier Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceCarrierGatewayRead(d, meta)
}

func resourceCarrierGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Carrier Gateway (%s)", d.Id())
	_, err := conn.DeleteCarrierGateway(&ec2.DeleteCarrierGatewayInput{
		CarrierGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidCarrierGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Carrier Gateway (%s): %w", d.Id(), err)
	}

	_, err = WaitCarrierGatewayDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Carrier Gateway (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
