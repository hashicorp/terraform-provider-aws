package ec2

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceEIPAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceEIPAssociationCreate,
		Read:   resourceEIPAssociationRead,
		Delete: resourceEIPAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"allow_reassociation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEIPAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.AssociateAddressInput{}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("allow_reassociation"); ok {
		input.AllowReassociation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_id"); ok {
		input.InstanceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_ip_address"); ok {
		input.PrivateIpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_ip"); ok {
		input.PublicIp = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 EIP Association: %s", input)
	output, err := conn.AssociateAddress(input)

	if err != nil {
		return fmt.Errorf("creating EC2 EIP Association: %w", err)
	}

	if output.AssociationId != nil {
		d.SetId(aws.StringValue(output.AssociationId))
	} else {
		// EC2-Classic.
		d.SetId(aws.StringValue(input.PublicIp))

		if err := waitForAddressAssociationClassic(conn, aws.StringValue(input.PublicIp), aws.StringValue(input.InstanceId)); err != nil {
			return fmt.Errorf("waiting for EC2 EIP (%s) to associate with EC2-Classic Instance (%s): %w", aws.StringValue(input.PublicIp), aws.StringValue(input.InstanceId), err)
		}
	}

	return resourceEIPAssociationRead(d, meta)
}

func resourceEIPAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	request, err := DescribeAddressesByID(d.Id(), meta.(*conns.AWSClient).SupportedPlatforms)
	if err != nil {
		return err
	}

	var response *ec2.DescribeAddressesOutput
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		response, err = conn.DescribeAddresses(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidAssociationID.NotFound") {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && (response.Addresses == nil || len(response.Addresses) == 0) {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		response, err = conn.DescribeAddresses(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidAssociationID.NotFound") {
		log.Printf("[WARN] EIP Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading EC2 Elastic IP %s: %#v", d.Get("allocation_id").(string), err)
	}

	if response.Addresses == nil || len(response.Addresses) == 0 {
		log.Printf("[INFO] EIP Association ID Not Found. Refreshing from state")
		d.SetId("")
		return nil
	}

	return readEIPAssociation(d, response.Addresses[0])
}

func resourceEIPAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DisassociateAddressInput{}

	if eipAssociationID(d.Id()).IsVPC() {
		input.AssociationId = aws.String(d.Id())
	} else {
		input.PublicIp = aws.String(d.Id())
	}

	log.Printf("[DEBUG] Deleting EC2 EIP Association: %s", d.Id())
	_, err := conn.DisassociateAddress(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 EIP Association (%s): %w", d.Id(), err)
	}

	return nil
}

func readEIPAssociation(d *schema.ResourceData, address *ec2.Address) error {
	if err := d.Set("allocation_id", address.AllocationId); err != nil {
		return err
	}
	if err := d.Set("instance_id", address.InstanceId); err != nil {
		return err
	}
	if err := d.Set("network_interface_id", address.NetworkInterfaceId); err != nil {
		return err
	}
	if err := d.Set("private_ip_address", address.PrivateIpAddress); err != nil {
		return err
	}
	if err := d.Set("public_ip", address.PublicIp); err != nil {
		return err
	}

	return nil
}

func DescribeAddressesByID(id string, supportedPlatforms []string) (*ec2.DescribeAddressesInput, error) {
	// We assume EC2 Classic if ID is a valid IPv4 address
	ip := net.ParseIP(id)
	if ip != nil {
		if len(supportedPlatforms) > 0 && !conns.HasEC2Classic(supportedPlatforms) {
			return nil, fmt.Errorf("Received IPv4 address as ID in account that doesn't support EC2 Classic (%q)",
				supportedPlatforms)
		}

		return &ec2.DescribeAddressesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("public-ip"),
					Values: []*string{aws.String(id)},
				},
				{
					Name:   aws.String("domain"),
					Values: []*string{aws.String("standard")},
				},
			},
		}, nil
	}

	return &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("association-id"),
				Values: []*string{aws.String(id)},
			},
		},
	}, nil
}

type eipAssociationID string

// IsVPC returns whether or not the associated EIP is in the VPC domain.
func (id eipAssociationID) IsVPC() bool {
	return strings.HasPrefix(string(id), "eipassoc-")
}
