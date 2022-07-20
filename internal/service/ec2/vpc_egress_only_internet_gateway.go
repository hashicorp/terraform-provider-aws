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
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeEgressOnlyInternetGateway),
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

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindEgressOnlyInternetGatewayByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Egress-only Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Egress-only Internet Gateway (%s): %w", d.Id(), err)
	}

	ig := outputRaw.(*ec2.EgressOnlyInternetGateway)

	if len(ig.Attachments) == 1 && aws.StringValue(ig.Attachments[0].State) == ec2.AttachmentStatusAttached {
		d.Set("vpc_id", ig.Attachments[0].VpcId)
	} else {
		d.Set("vpc_id", nil)
	}

	tags := KeyValueTags(ig.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
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

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Egress-only Internet Gateway (%s): %w", d.Id(), err)
	}

	return nil
}
