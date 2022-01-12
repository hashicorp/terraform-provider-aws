package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEgressOnlyInternetGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceEgressOnlyInternetGatewayCreate,
		Read:   resourceEgressOnlyInternetGatewayRead,
		Update: resourceEgressOnlyInternetGatewayUpdate,
		Delete: resourceEgressOnlyInternetGatewayDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
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

func resourceEgressOnlyInternetGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateEgressOnlyInternetGatewayInput{
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeEgressOnlyInternetGateway),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Egress-only Internet Gateway: %s", input)
	output, err := conn.CreateEgressOnlyInternetGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Egress-only Internet Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId))

	return resourceEgressOnlyInternetGatewayRead(d, meta)
}

func resourceEgressOnlyInternetGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var req = &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: []*string{aws.String(d.Id())},
	}

	var resp *ec2.DescribeEgressOnlyInternetGatewaysOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.DescribeEgressOnlyInternetGateways(req)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		igw := getEc2EgressOnlyInternetGateway(d.Id(), resp)
		if d.IsNewResource() && igw == nil {
			return resource.RetryableError(fmt.Errorf("Egress Only Internet Gateway (%s) not found.", d.Id()))
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeEgressOnlyInternetGateways(req)
	}

	if err != nil {
		return fmt.Errorf("Error describing egress internet gateway: %s", err)
	}

	igw := getEc2EgressOnlyInternetGateway(d.Id(), resp)
	if igw == nil {
		log.Printf("[Error] Cannot find Egress Only Internet Gateway: %q", d.Id())
		d.SetId("")
		return nil
	}

	if len(igw.Attachments) == 1 && aws.StringValue(igw.Attachments[0].State) == ec2.AttachmentStatusAttached {
		d.Set("vpc_id", igw.Attachments[0].VpcId)
	}

	tags := KeyValueTags(igw.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func getEc2EgressOnlyInternetGateway(id string, resp *ec2.DescribeEgressOnlyInternetGatewaysOutput) *ec2.EgressOnlyInternetGateway {
	if resp != nil && len(resp.EgressOnlyInternetGateways) > 0 {
		for _, igw := range resp.EgressOnlyInternetGateways {
			if aws.StringValue(igw.EgressOnlyInternetGatewayId) == id {
				return igw
			}
		}
	}
	return nil
}

func resourceEgressOnlyInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Egress-only Internet Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceEgressOnlyInternetGatewayRead(d, meta)
}

func resourceEgressOnlyInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Egress-only Internet Gateway: %s", d.Id())
	_, err := conn.DeleteEgressOnlyInternetGateway(&ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Egress-only Internet Gateway (%s): %w", d.Id(), err)
	}

	return nil
}
