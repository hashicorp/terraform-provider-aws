package ec2

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.IsIPv4Address,
				),
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.GatewayType_Values(), false),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomerGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	ipAddress := d.Get("ip_address").(string)
	vpnType := d.Get("type").(string)
	bgpAsn := d.Get("bgp_asn").(string)
	deviceName := d.Get("device_name").(string)

	alreadyExists, err := resourceCustomerGatewayExists(vpnType, ipAddress, bgpAsn, deviceName, conn)
	if err != nil {
		return err
	}

	if alreadyExists {
		return fmt.Errorf("An existing customer gateway for IpAddress: %s, VpnType: %s, BGP ASN: %s has been found", ipAddress, vpnType, bgpAsn)
	}

	i64BgpAsn, err := strconv.ParseInt(bgpAsn, 10, 64)
	if err != nil {
		return err
	}

	createOpts := &ec2.CreateCustomerGatewayInput{
		BgpAsn:            aws.Int64(i64BgpAsn),
		Type:              aws.String(vpnType),
		PublicIp:          aws.String(ipAddress),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeCustomerGateway),
	}

	if len(deviceName) != 0 {
		createOpts.DeviceName = aws.String(deviceName)
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		createOpts.CertificateArn = aws.String(v.(string))
	}

	// Create the Customer Gateway.
	log.Printf("[DEBUG] Creating customer gateway")
	resp, err := conn.CreateCustomerGateway(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating customer gateway: %w", err)
	}

	// Store the ID
	customerGateway := resp.CustomerGateway
	cgId := aws.StringValue(customerGateway.CustomerGatewayId)
	d.SetId(cgId)
	log.Printf("[INFO] Customer gateway ID: %s", cgId)

	// Wait for the CustomerGateway to be available.
	_, err = waitCustomerGatewayAvailable(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for customer gateway (%s) to become available: %w", d.Id(), err)
	}

	return resourceCustomerGatewayRead(d, meta)
}

func resourceCustomerGatewayExists(vpnType, ipAddress, bgpAsn, deviceName string, conn *ec2.EC2) (bool, error) {
	filters := []*ec2.Filter{
		{
			Name:   aws.String("ip-address"),
			Values: []*string{aws.String(ipAddress)},
		},
		{
			Name:   aws.String("type"),
			Values: []*string{aws.String(vpnType)},
		},
		{
			Name:   aws.String("bgp-asn"),
			Values: []*string{aws.String(bgpAsn)},
		},
	}

	if len(deviceName) != 0 {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("device-name"),
			Values: []*string{aws.String(deviceName)},
		})
	}

	resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
		Filters: filters,
	})
	if err != nil {
		return false, err
	}

	if len(resp.CustomerGateways) > 0 && aws.StringValue(resp.CustomerGateways[0].State) != CustomerGatewayStateDeleted {
		return true, nil
	}

	return false, nil
}

func resourceCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	customerGateway, err := FindCustomerGatewayById(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Customer Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Customer Gateway (%s): %w", d.Id(), err)
	}

	d.Set("bgp_asn", customerGateway.BgpAsn)
	d.Set("ip_address", customerGateway.IpAddress)
	d.Set("type", customerGateway.Type)
	d.Set("device_name", customerGateway.DeviceName)
	d.Set("certificate_arn", customerGateway.CertificateArn)

	tags := KeyValueTags(customerGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceCustomerGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Customer Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCustomerGatewayRead(d, meta)
}

func resourceCustomerGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	_, err := conn.DeleteCustomerGateway(&ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ErrCodeInvalidCustomerGatewayIDNotFound) {
			return nil
		}

		return fmt.Errorf("[ERROR] Error deleting Customer Gateway (%s): %w", d.Id(), err)
	}

	_, err = waitCustomerGatewayDeleted(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for customer gateway (%s) to become deleted: %w", d.Id(), err)
	}

	return nil
}
