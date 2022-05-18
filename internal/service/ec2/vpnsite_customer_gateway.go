package ec2

import (
	"fmt"
	"log"
	"strconv"

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

func ResourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomerGatewayCreate,
		Read:   resourceCustomerGatewayRead,
		Update: resourceCustomerGatewayUpdate,
		Delete: resourceCustomerGatewayDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: valid4ByteASN,
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"device_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.GatewayType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomerGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	i64BgpAsn, err := strconv.ParseInt(d.Get("bgp_asn").(string), 10, 64)

	if err != nil {
		return err
	}

	input := &ec2.CreateCustomerGatewayInput{
		BgpAsn:            aws.Int64(i64BgpAsn),
		PublicIp:          aws.String(d.Get("ip_address").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeCustomerGateway),
		Type:              aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("device_name"); ok {
		input.DeviceName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Customer Gateway: %s", input)
	output, err := conn.CreateCustomerGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Customer Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.CustomerGateway.CustomerGatewayId))

	if _, err := WaitCustomerGatewayCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Customer Gateway (%s) create: %w", d.Id(), err)
	}

	return resourceCustomerGatewayRead(d, meta)
}

func resourceCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	customerGateway, err := FindCustomerGatewayByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Customer Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Customer Gateway (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("bgp_asn", customerGateway.BgpAsn)
	d.Set("certificate_arn", customerGateway.CertificateArn)
	d.Set("device_name", customerGateway.DeviceName)
	d.Set("ip_address", customerGateway.IpAddress)
	d.Set("type", customerGateway.Type)

	tags := KeyValueTags(customerGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCustomerGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Customer Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceCustomerGatewayRead(d, meta)
}

func resourceCustomerGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Customer Gateway: %s", d.Id())
	_, err := conn.DeleteCustomerGateway(&ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Customer Gateway (%s): %w", d.Id(), err)
	}

	if _, err := WaitCustomerGatewayDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Customer Gateway (%s) delete: %w", d.Id(), err)
	}

	return nil
}
