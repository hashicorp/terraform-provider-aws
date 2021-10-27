package ec2

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkInterfaceCreate,
		Read:   resourceNetworkInterfaceRead,
		Update: resourceNetworkInterfaceUpdate,
		Delete: resourceNetworkInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"instance": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interface_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.NetworkInterfaceCreationType_Values(), false),
			},
			"ipv6_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_addresses"},
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count"},
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"private_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ips_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateNetworkInterfaceInput{
		SubnetId:          aws.String(d.Get("subnet_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkInterface),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("interface_type"); ok {
		input.InterfaceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_address_count"); ok {
		input.Ipv6AddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Ipv6Addresses = expandIP6Addresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
		input.PrivateIpAddresses = ExpandPrivateIPAddresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("private_ips_count"); ok {
		input.SecondaryPrivateIpAddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.Groups = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 Network Interface: %s", input)
	output, err := conn.CreateNetworkInterface(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Network Interface: %w", err)
	}

	d.SetId(aws.StringValue(output.NetworkInterface.NetworkInterfaceId))

	if _, err := WaitNetworkInterfaceCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for EC2 Network Interface (%s) create: %w", d.Id(), err)
	}

	//Default value is enabled
	if !d.Get("source_dest_check").(bool) {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(input)

		if err != nil {
			return fmt.Errorf("error setting EC2 Network Interface (%s) SourceDestCheck: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		err := attachNetworkInterface(conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), 5*time.Minute)

		if err != nil {
			return err
		}
	}

	return resourceNetworkInterfaceRead(d, meta)
}

func resourceNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(1*time.Minute, func() (interface{}, error) {
		return FindNetworkInterfaceByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s): %w", d.Id(), err)
	}

	eni := outputRaw.(*ec2.NetworkInterface)

	attachment := []map[string]interface{}{}

	if eni.Attachment != nil {
		attachment = []map[string]interface{}{FlattenAttachment(eni.Attachment)}
	}

	if err := d.Set("attachment", attachment); err != nil {
		return fmt.Errorf("error setting attachment: %w", err)
	}

	ownerID := aws.StringValue(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-interface/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("description", eni.Description)
	d.Set("interface_type", eni.InterfaceType)

	d.Set("ipv6_address_count", len(eni.Ipv6Addresses))

	if err := d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Address(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6_addresses: %w", err)
	}

	d.Set("mac_address", eni.MacAddress)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set("owner_id", ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)

	if err := d.Set("private_ips", FlattenNetworkInterfacesPrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return fmt.Errorf("error setting private_ips: %w", err)
	}

	d.Set("private_ips_count", len(eni.PrivateIpAddresses)-1)

	if err := d.Set("security_groups", FlattenGroupIdentifiers(eni.Groups)); err != nil {
		return fmt.Errorf("error setting security_groups: %w", err)
	}

	d.Set("source_dest_check", eni.SourceDestCheck)
	d.Set("subnet_id", eni.SubnetId)

	tags := KeyValueTags(eni.TagSet).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func networkInterfaceAttachmentRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(id)},
		}
		describeResp, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

		if err != nil {
			log.Printf("[ERROR] Could not find network interface %s. %s", id, err)
			return nil, "", err
		}

		eni := describeResp.NetworkInterfaces[0]
		hasAttachment := strconv.FormatBool(eni.Attachment != nil)
		log.Printf("[DEBUG] ENI %s has attachment state %s", id, hasAttachment)
		return eni, hasAttachment, nil
	}
}

func resourceNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("attachment") {
		oa, na := d.GetChange("attachment")

		if oa != nil && oa.(*schema.Set).Len() > 0 {
			attachment := oa.(*schema.Set).List()[0].(map[string]interface{})

			err := detachNetworkInterface(conn, d.Id(), attachment["attachment_id"].(string), 10*time.Minute)

			if err != nil {
				return err
			}
		}

		// if there is a new attachment, attach it
		if na != nil && na.(*schema.Set).Len() > 0 {
			attachment := na.(*schema.Set).List()[0].(map[string]interface{})

			err := attachNetworkInterface(conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), 5*time.Minute)

			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("private_ips") {
		o, n := d.GetChange("private_ips")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IP addresses
		unassignIps := os.Difference(ns)
		if unassignIps.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(unassignIps),
			}
			_, err := conn.UnassignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to unassign Private IPs: %s", err)
			}
		}

		// Assign new IP addresses
		assignIps := ns.Difference(os)
		if assignIps.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(assignIps),
			}
			_, err := conn.AssignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign Private IPs: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_addresses") {
		o, n := d.GetChange("ipv6_addresses")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 addresses
		unassignIps := os.Difference(ns)
		if unassignIps.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringSet(unassignIps),
			}
			_, err := conn.UnassignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
			}
		}

		// Assign new IPV6 addresses
		assignIps := ns.Difference(os)
		if assignIps.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringSet(assignIps),
			}
			_, err := conn.AssignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign IPV6 Addresses: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_address_count") {
		o, n := d.GetChange("ipv6_address_count")
		ipv6Addresses := d.Get("ipv6_addresses").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv6Addresses) {

			diff := n.(int) - o.(int)

			// Surplus of IPs, add the diff
			if diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6AddressCount:   aws.Int64(int64(diff)),
				}
				_, err := conn.AssignIpv6Addresses(input)
				if err != nil {
					return fmt.Errorf("failure to assign IPV6 Addresses: %s", err)
				}
			}

			if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Addresses:      flex.ExpandStringList(ipv6Addresses[0:int(math.Abs(float64(diff)))]),
				}
				_, err := conn.UnassignIpv6Addresses(input)
				if err != nil {
					return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
				}
			}
		}
	}

	if d.HasChange("source_dest_check") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(d.Get("source_dest_check").(bool))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("failure updating Source Dest Check on ENI: %s", err)
		}
	}

	if d.HasChange("private_ips_count") {
		o, n := d.GetChange("private_ips_count")
		privateIPs := d.Get("private_ips").(*schema.Set).List()
		privateIPsFiltered := privateIPs[:0]
		primaryIP := d.Get("private_ip")

		for _, ip := range privateIPs {
			if ip != primaryIP {
				privateIPsFiltered = append(privateIPsFiltered, ip)
			}
		}

		if o != nil && n != nil && n != len(privateIPsFiltered) {

			diff := n.(int) - o.(int)

			// Surplus of IPs, add the diff
			if diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId:             aws.String(d.Id()),
					SecondaryPrivateIpAddressCount: aws.Int64(int64(diff)),
				}
				_, err := conn.AssignPrivateIpAddresses(input)
				if err != nil {
					return fmt.Errorf("Failure to assign Private IPs: %s", err)
				}
			}

			if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringList(privateIPsFiltered[0:int(math.Abs(float64(diff)))]),
				}
				_, err := conn.UnassignPrivateIpAddresses(input)
				if err != nil {
					return fmt.Errorf("Failure to unassign Private IPs: %s", err)
				}
			}
		}
	}

	if d.HasChange("security_groups") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Groups:             flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("Failure updating ENI: %s", err)
		}
	}

	if d.HasChange("description") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Description:        &ec2.AttributeValue{Value: aws.String(d.Get("description").(string))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("Failure updating ENI: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Network Interface (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceNetworkInterfaceRead(d, meta)
}

func resourceNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		err := detachNetworkInterface(conn, d.Id(), attachment["attachment_id"].(string), 10*time.Minute)

		if err != nil {
			return err
		}
	}

	return deleteNetworkInterface(conn, d.Id())
}

func attachNetworkInterface(conn *ec2.EC2, networkInterfaceID, instanceID string, deviceIndex int, timeout time.Duration) error {
	input := &ec2.AttachNetworkInterfaceInput{
		DeviceIndex:        aws.Int64(int64(deviceIndex)),
		InstanceId:         aws.String(instanceID),
		NetworkInterfaceId: aws.String(networkInterfaceID),
	}

	log.Printf("[INFO] Attaching EC2 Network Interface: %s", input)
	output, err := conn.AttachNetworkInterface(input)

	if err != nil {
		return fmt.Errorf("error attaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, instanceID, err)
	}

	attachmentID := aws.StringValue(output.AttachmentId)

	_, err = WaitNetworkInterfaceAttached(conn, attachmentID, timeout)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Network Interface (%s/%s) attach: %w", networkInterfaceID, attachmentID, err)
	}

	return nil
}

func deleteNetworkInterface(conn *ec2.EC2, networkInterfaceID string) error {
	log.Printf("[INFO] Deleting EC2 Network Interface: %s", networkInterfaceID)
	_, err := conn.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInterfaceIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	return nil
}

func detachNetworkInterface(conn *ec2.EC2, networkInterfaceID, attachmentID string, timeout time.Duration) error {
	input := &ec2.DetachNetworkInterfaceInput{
		AttachmentId: aws.String(attachmentID),
		Force:        aws.Bool(true),
	}

	log.Printf("[INFO] Detaching EC2 Network Interface: %s", input)
	_, err := conn.DetachNetworkInterface(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error detaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, attachmentID, err)
	}

	_, err = WaitNetworkInterfaceDetached(conn, attachmentID, timeout)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Network Interface (%s/%s) detach: %w", networkInterfaceID, attachmentID, err)
	}

	return nil
}
