package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomerGatewayCreate,
		ReadWithoutTimeout:   resourceCustomerGatewayRead,
		UpdateWithoutTimeout: resourceCustomerGatewayUpdate,
		DeleteWithoutTimeout: resourceCustomerGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ValidateFunc: verify.Valid4ByteASN,
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
				Optional:     true,
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

func resourceCustomerGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateCustomerGatewayInput{
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeCustomerGateway),
		Type:              aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("bgp_asn"); ok {
		v, err := strconv.ParseInt(v.(string), 10, 64)

		if err != nil {
			return diag.FromErr(err)
		}

		input.BgpAsn = aws.Int64(v)
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("device_name"); ok {
		input.DeviceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}

	output, err := conn.CreateCustomerGatewayWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating EC2 Customer Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.CustomerGateway.CustomerGatewayId))

	if _, err := WaitCustomerGatewayCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 Customer Gateway (%s) create: %s", d.Id(), err)
	}

	return resourceCustomerGatewayRead(ctx, d, meta)
}

func resourceCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	customerGateway, err := FindCustomerGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Customer Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 Customer Gateway (%s): %s", d.Id(), err)
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
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceCustomerGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating EC2 Customer Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCustomerGatewayRead(ctx, d, meta)
}

func resourceCustomerGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 Customer Gateway: %s", d.Id())
	_, err := conn.DeleteCustomerGatewayWithContext(ctx, &ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 Customer Gateway (%s): %s", d.Id(), err)
	}

	if _, err := WaitCustomerGatewayDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 Customer Gateway (%s) delete: %s", d.Id(), err)
	}

	return nil
}
