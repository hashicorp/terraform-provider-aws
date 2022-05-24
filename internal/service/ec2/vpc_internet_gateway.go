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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceInternetGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceInternetGatewayCreate,
		Read:   resourceInternetGatewayRead,
		Update: resourceInternetGatewayUpdate,
		Delete: resourceInternetGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInternetGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateInternetGatewayInput{
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeInternetGateway),
	}

	log.Printf("[DEBUG] Creating EC2 Internet Gateway: %s", input)
	output, err := conn.CreateInternetGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Internet Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.InternetGateway.InternetGatewayId))

	if v, ok := d.GetOk("vpc_id"); ok {
		if err := attachInternetGateway(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	return resourceInternetGatewayRead(d, meta)
}

func resourceInternetGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindInternetGatewayByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Internet Gateway (%s): %w", d.Id(), err)
	}

	ig := outputRaw.(*ec2.InternetGateway)

	ownerID := aws.StringValue(ig.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("internet-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)
	if len(ig.Attachments) == 0 {
		// Gateway exists but not attached to the VPC.
		d.Set("vpc_id", "")
	} else {
		d.Set("vpc_id", ig.Attachments[0].VpcId)
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

func resourceInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("vpc_id") {
		o, n := d.GetChange("vpc_id")

		if v := o.(string); v != "" {
			if err := detachInternetGateway(conn, d.Id(), v); err != nil {
				return err
			}
		}

		if v := n.(string); v != "" {
			if err := attachInternetGateway(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Internet Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceInternetGatewayRead(d, meta)
}

func resourceInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Detach if it is attached.
	if v, ok := d.GetOk("vpc_id"); ok {
		if err := detachInternetGateway(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	input := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting Internet Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(internetGatewayDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteInternetGateway(input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Internet Gateway (%s): %w", d.Id(), err)
	}

	return nil
}

func attachInternetGateway(conn *ec2.EC2, internetGatewayID, vpcID string) error {
	input := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGatewayID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[INFO] Attaching EC2 Internet Gateway: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, func() (interface{}, error) {
		return conn.AttachInternetGateway(input)
	}, errCodeInvalidInternetGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("error attaching EC2 Internet Gateway (%s) to VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	_, err = WaitInternetGatewayAttached(conn, internetGatewayID, vpcID, internetGatewayAttachedTimeout)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Internet Gateway (%s) to attach to VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	return nil
}

func detachInternetGateway(conn *ec2.EC2, internetGatewayID, vpcID string) error {
	input := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGatewayID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[INFO] Detaching EC2 Internet Gateway: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(internetGatewayDetachedTimeout, func() (interface{}, error) {
		return conn.DetachInternetGateway(input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeGatewayNotAttached) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error detaching EC2 Internet Gateway (%s) from VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	_, err = WaitInternetGatewayDetached(conn, internetGatewayID, vpcID, internetGatewayDetachedTimeout)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Internet Gateway (%s) to detach from VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	return nil
}
