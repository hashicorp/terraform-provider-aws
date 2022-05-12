package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			"ipv4_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				},
				ConflictsWith: []string{"ipv4_prefix_count"},
			},
			"ipv4_prefix_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv4_prefixes"},
			},
			"ipv6_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_list"},
			},
			"ipv6_address_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_count"},
			},
			"ipv6_address_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count", "ipv6_address_list"},
			},
			"ipv6_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
				},
				ConflictsWith: []string{"ipv6_prefix_count"},
			},
			"ipv6_prefix_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_prefixes"},
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
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ip_list"},
			},
			"private_ips_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"private_ip_list"},
			},
			"private_ip_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ips", "private_ips_count"},
			},
			"private_ip_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIf("private_ips", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				privateIPListEnabled := d.Get("private_ip_list_enabled").(bool)
				if privateIPListEnabled {
					return false
				}
				_, new := d.GetChange("private_ips")
				if new != nil {
					oldPrimaryIP := ""
					if v, ok := d.GetOk("private_ip_list"); ok {
						for _, ip := range v.([]interface{}) {
							oldPrimaryIP = ip.(string)
							break
						}
					}
					for _, ip := range new.(*schema.Set).List() {
						// no need for new resource if we still have the primary ip
						if oldPrimaryIP == ip.(string) {
							return false
						}
					}
					// new primary ip requires a new resource
					return true
				} else {
					return false
				}
			}),
			customdiff.ForceNewIf("private_ip_list", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				privateIPListEnabled := d.Get("private_ip_list_enabled").(bool)
				if !privateIPListEnabled {
					return false
				}
				old, new := d.GetChange("private_ip_list")
				if old != nil && new != nil {
					oldPrimaryIP := ""
					newPrimaryIP := ""
					for _, ip := range old.([]interface{}) {
						oldPrimaryIP = ip.(string)
						break
					}
					for _, ip := range new.([]interface{}) {
						newPrimaryIP = ip.(string)
						break
					}

					// change in primary private ip requires a new resource
					return oldPrimaryIP != newPrimaryIP
				} else {
					return false
				}
			}),
			customdiff.ComputedIf("private_ips", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("private_ips_count")
					}
				} else {
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ips_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips_count") {
						return false
					} else {
						// compute the new count if private_ips change
						return diff.HasChange("private_ips")
					}
				} else {
					// compute the new count if private_ip_list changes
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ip_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("private_ip_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if private_ips or private_ips_count changes
					return diff.HasChange("private_ips") || diff.HasChange("private_ips_count") || diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("ipv6_addresses", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("ipv6_address_count")
					}
				} else {
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("ipv6_address_count") {
						return false
					} else {
						// compute the new count if ipv6_addresses change
						return diff.HasChange("ipv6_addresses")
					}
				} else {
					// compute the new count if ipv6_address_list changes
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("ipv6_address_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if anything changes
					return diff.HasChange("ipv6_addresses") || diff.HasChange("ipv6_address_count") || diff.HasChange("ipv6_address_list")
				}
			}),
		),
	}
}

func resourceNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	ipv4PrefixesSpecified := false
	ipv6PrefixesSpecified := false

	input := &ec2.CreateNetworkInterfaceInput{
		SubnetId: aws.String(d.Get("subnet_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("interface_type"); ok {
		input.InterfaceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv4PrefixesSpecified = true
		input.Ipv4Prefixes = expandIPv4PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv4_prefix_count"); ok {
		input.Ipv4PrefixCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_address_count"); ok {
		input.Ipv6AddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Ipv6Addresses = expandInstanceIPv6Addresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv6PrefixesSpecified = true
		input.Ipv6Prefixes = expandIPv6PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefix_count"); ok {
		input.Ipv6PrefixCount = aws.Int64(int64(v.(int)))
	}

	if d.Get("private_ip_list_enabled").(bool) {
		if v, ok := d.GetOk("private_ip_list"); ok && len(v.([]interface{})) > 0 {
			input.PrivateIpAddresses = expandPrivateIPAddressSpecifications(v.([]interface{}))
		}
	} else {
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			privateIPs := v.(*schema.Set).List()
			// total includes the primary
			totalPrivateIPs := len(privateIPs)
			// private_ips_count is for secondaries
			if v, ok := d.GetOk("private_ips_count"); ok {
				// reduce total count if necessary
				if v.(int)+1 < totalPrivateIPs {
					totalPrivateIPs = v.(int) + 1
				}
			}
			// truncate the list
			countLimitedIPs := make([]interface{}, totalPrivateIPs)
			for i, ip := range privateIPs {
				countLimitedIPs[i] = ip.(string)
				if i == totalPrivateIPs-1 {
					break
				}
			}
			input.PrivateIpAddresses = expandPrivateIPAddressSpecifications(countLimitedIPs)
		} else {
			if v, ok := d.GetOk("private_ips_count"); ok {
				input.SecondaryPrivateIpAddressCount = aws.Int64(int64(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.Groups = flex.ExpandStringSet(v.(*schema.Set))
	}

	// If IPv4 or IPv6 prefixes are specified, tag after create.
	// Otherwise "An error occurred (InternalError) when calling the CreateNetworkInterface operation".
	if len(tags) > 0 && !(ipv4PrefixesSpecified || ipv6PrefixesSpecified) {
		input.TagSpecifications = ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkInterface)
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

	if !d.Get("private_ip_list_enabled").(bool) {
		// add more ips to match the count
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			totalPrivateIPs := v.(*schema.Set).Len()
			if privateIPsCount, ok := d.GetOk("private_ips_count"); ok {
				if privateIPsCount.(int)+1 > totalPrivateIPs {
					input := &ec2.AssignPrivateIpAddressesInput{
						NetworkInterfaceId:             aws.String(d.Id()),
						SecondaryPrivateIpAddressCount: aws.Int64(int64(privateIPsCount.(int) + 1 - totalPrivateIPs)),
					}
					_, err := conn.AssignPrivateIpAddresses(input)
					if err != nil {
						return fmt.Errorf("Failure to assign Private IPs: %s", err)
					}
				}
			}
		}
	}

	if len(tags) > 0 && (ipv4PrefixesSpecified || ipv6PrefixesSpecified) {
		if err := UpdateTags(conn, d.Id(), nil, tags); err != nil {
			return fmt.Errorf("error updating EC2 Network Interface (%s) tags: %w", d.Id(), err)
		}
	}

	// Default value is enabled.
	if !d.Get("source_dest_check").(bool) {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
		}

		log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
		_, err := conn.ModifyNetworkInterfaceAttribute(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 Network Interface (%s) SourceDestCheck: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		_, err := attachNetworkInterface(conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), networkInterfaceAttachedTimeout)

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

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
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

	ownerID := aws.StringValue(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-interface/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	if eni.Attachment != nil {
		if err := d.Set("attachment", []interface{}{flattenNetworkInterfaceAttachment(eni.Attachment)}); err != nil {
			return fmt.Errorf("error setting attachment: %w", err)
		}
	} else {
		d.Set("attachment", nil)
	}

	d.Set("description", eni.Description)
	d.Set("interface_type", eni.InterfaceType)

	if err := d.Set("ipv4_prefixes", flattenIPv4PrefixSpecifications(eni.Ipv4Prefixes)); err != nil {
		return fmt.Errorf("error setting ipv4_prefixes: %w", err)
	}

	d.Set("ipv4_prefix_count", len(eni.Ipv4Prefixes))

	d.Set("ipv6_address_count", len(eni.Ipv6Addresses))

	if err := d.Set("ipv6_address_list", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6 address list: %s", err)
	}

	if err := d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6_addresses: %w", err)
	}

	if err := d.Set("ipv6_prefixes", flattenIPv6PrefixSpecifications(eni.Ipv6Prefixes)); err != nil {
		return fmt.Errorf("error setting ipv6_prefixes: %w", err)
	}

	d.Set("ipv6_prefix_count", len(eni.Ipv6Prefixes))

	d.Set("mac_address", eni.MacAddress)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set("owner_id", ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)

	if err := d.Set("private_ips", FlattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return fmt.Errorf("error setting private_ips: %w", err)
	}

	d.Set("private_ips_count", len(eni.PrivateIpAddresses)-1)

	if err := d.Set("private_ip_list", FlattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return fmt.Errorf("error setting private_ip_list: %s", err)
	}

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

func resourceNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	privateIPsNetChange := 0

	if d.HasChange("attachment") {
		oa, na := d.GetChange("attachment")

		if oa != nil && oa.(*schema.Set).Len() > 0 {
			attachment := oa.(*schema.Set).List()[0].(map[string]interface{})

			err := DetachNetworkInterface(conn, d.Id(), attachment["attachment_id"].(string), NetworkInterfaceDetachedTimeout)

			if err != nil {
				return err
			}
		}

		if na != nil && na.(*schema.Set).Len() > 0 {
			attachment := na.(*schema.Set).List()[0].(map[string]interface{})

			_, err := attachNetworkInterface(conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), networkInterfaceAttachedTimeout)

			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("private_ips") && !d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ips")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IP addresses.
		unassignIPs := os.Difference(ns)
		if unassignIPs.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(unassignIPs),
			}

			log.Printf("[INFO] Unassigning private IPv4 addresses: %s", input)
			_, err := conn.UnassignPrivateIpAddresses(input)

			if err != nil {
				return fmt.Errorf("error unassigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
			}

			privateIPsNetChange -= unassignIPs.Len()
		}

		// Assign new IP addresses.
		assignIPs := ns.Difference(os)
		if assignIPs.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(assignIPs),
			}

			log.Printf("[INFO] Assigning private IPv4 addresses: %s", input)
			_, err := conn.AssignPrivateIpAddresses(input)

			if err != nil {
				return fmt.Errorf("error assigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
			}
			privateIPsNetChange += assignIPs.Len()
		}
	}

	if d.HasChange("private_ip_list") && d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ip_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}
		if len(o.([]interface{}))-1 > 0 {
			privateIPsToUnassign := make([]interface{}, len(o.([]interface{}))-1)
			idx := 0
			for i, ip := range o.([]interface{}) {
				// skip primary private ip address
				if i == 0 {
					continue
				}
				privateIPsToUnassign[idx] = ip
				log.Printf("[INFO] Unassigning private ip %s", ip)
				idx += 1
			}

			// Unassign the secondary IP addresses
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringList(privateIPsToUnassign),
			}
			_, err := conn.UnassignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to unassign Private IPs: %s", err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for i, ip := range n.([]interface{}) {
			// skip primary private ip address
			if i == 0 {
				continue
			}
			privateIPToAssign := []interface{}{ip}
			log.Printf("[INFO] Assigning private ip %s", ip)

			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringList(privateIPToAssign),
			}
			_, err := conn.AssignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign Private IPs: %s", err)
			}
		}
	}

	if d.HasChange("private_ips_count") && !d.Get("private_ip_list_enabled").(bool) {
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
			if diff := n.(int) - o.(int) - privateIPsNetChange; diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId:             aws.String(d.Id()),
					SecondaryPrivateIpAddressCount: aws.Int64(int64(diff)),
				}

				log.Printf("[INFO] Assigning private IPv4 addresses: %s", input)
				_, err := conn.AssignPrivateIpAddresses(input)

				if err != nil {
					return fmt.Errorf("error assigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringList(privateIPsFiltered[0:-diff]),
				}

				log.Printf("[INFO] Unassigning private IPv4 addresses: %s", input)
				_, err := conn.UnassignPrivateIpAddresses(input)

				if err != nil {
					return fmt.Errorf("error unassigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv4_prefix_count") {
		o, n := d.GetChange("ipv4_prefix_count")
		ipv4Prefixes := d.Get("ipv4_prefixes").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv4Prefixes) {
			if diff := n.(int) - o.(int); diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv4PrefixCount:    aws.Int64(int64(diff)),
				}

				log.Printf("[INFO] Assigning private IPv4 addresses: %s", input)
				_, err := conn.AssignPrivateIpAddresses(input)

				if err != nil {
					return fmt.Errorf("error assigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv4Prefixes:       flex.ExpandStringList(ipv4Prefixes[0:-diff]),
				}

				log.Printf("[INFO] Unassigning private IPv4 addresses: %s", input)
				_, err := conn.UnassignPrivateIpAddresses(input)

				if err != nil {
					return fmt.Errorf("error unassigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv4_prefixes") {
		o, n := d.GetChange("ipv4_prefixes")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV4 prefixes.
		unassignPrefixes := os.Difference(ns)
		if unassignPrefixes.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv4Prefixes:       flex.ExpandStringSet(unassignPrefixes),
			}

			log.Printf("[INFO] Unassigning private IPv4 addresses: %s", input)
			_, err := conn.UnassignPrivateIpAddresses(input)

			if err != nil {
				return fmt.Errorf("error unassigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
			}
		}

		// Assign new IPV4 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv4Prefixes:       flex.ExpandStringSet(assignPrefixes),
			}

			log.Printf("[INFO] Assigning private IPv4 addresses: %s", input)
			_, err := conn.AssignPrivateIpAddresses(input)

			if err != nil {
				return fmt.Errorf("error assigning EC2 Network Interface (%s) private IPv4 addresses: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_addresses") && !d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_addresses")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 addresses.
		unassignIPs := os.Difference(ns)
		if unassignIPs.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringSet(unassignIPs),
			}

			log.Printf("[INFO] Unassigning IPv6 addresses: %s", input)
			_, err := conn.UnassignIpv6Addresses(input)

			if err != nil {
				return fmt.Errorf("error unassigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
			}
		}

		// Assign new IPV6 addresses,
		assignIPs := ns.Difference(os)
		if assignIPs.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringSet(assignIPs),
			}

			log.Printf("[INFO] Assigning IPv6 addresses: %s", input)
			_, err := conn.AssignIpv6Addresses(input)

			if err != nil {
				return fmt.Errorf("error assigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_address_count") && !d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_address_count")
		ipv6Addresses := d.Get("ipv6_addresses").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv6Addresses) {
			if diff := n.(int) - o.(int); diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6AddressCount:   aws.Int64(int64(diff)),
				}

				log.Printf("[INFO] Assigning IPv6 addresses: %s", input)
				_, err := conn.AssignIpv6Addresses(input)

				if err != nil {
					return fmt.Errorf("error assigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Addresses:      flex.ExpandStringList(ipv6Addresses[0:-diff]),
				}

				log.Printf("[INFO] Unassigning IPv6 addresses: %s", input)
				_, err := conn.UnassignIpv6Addresses(input)

				if err != nil {
					return fmt.Errorf("error unassigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv6_address_list") && d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_address_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}

		// Unassign old IPV6 addresses
		if len(o.([]interface{})) > 0 {
			unassignIPs := make([]interface{}, len(o.([]interface{})))
			for i, ip := range o.([]interface{}) {
				unassignIPs[i] = ip
				log.Printf("[INFO] Unassigning ipv6 address %s", ip)
			}

			log.Printf("[INFO] Unassigning ipv6 addresses")
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringList(unassignIPs),
			}
			_, err := conn.UnassignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for _, ip := range n.([]interface{}) {
			privateIPToAssign := []interface{}{ip}
			log.Printf("[INFO] Assigning ipv6 address %s", ip)

			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringList(privateIPToAssign),
			}
			_, err := conn.AssignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign IPV6 Addresses: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_prefixes") {
		o, n := d.GetChange("ipv6_prefixes")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 prefixes.
		unassignPrefixes := os.Difference(ns)
		if unassignPrefixes.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Prefixes:       flex.ExpandStringSet(unassignPrefixes),
			}

			log.Printf("[INFO] Unassigning IPv6 addresses: %s", input)
			_, err := conn.UnassignIpv6Addresses(input)

			if err != nil {
				return fmt.Errorf("error unassigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
			}
		}

		// Assign new IPV6 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Prefixes:       flex.ExpandStringSet(assignPrefixes),
			}

			log.Printf("[INFO] Assigning IPv6 addresses: %s", input)
			_, err := conn.AssignIpv6Addresses(input)

			if err != nil {
				return fmt.Errorf("error assigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_prefix_count") {
		o, n := d.GetChange("ipv6_prefix_count")
		ipv6Prefixes := d.Get("ipv6_prefixes").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv6Prefixes) {
			if diff := n.(int) - o.(int); diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6PrefixCount:    aws.Int64(int64(diff)),
				}

				log.Printf("[INFO] Assigning IPv6 addresses: %s", input)
				_, err := conn.AssignIpv6Addresses(input)

				if err != nil {
					return fmt.Errorf("error assigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Prefixes:       flex.ExpandStringList(ipv6Prefixes[0:-diff]),
				}

				log.Printf("[INFO] Unassigning IPv6 addresses: %s", input)
				_, err := conn.UnassignIpv6Addresses(input)

				if err != nil {
					return fmt.Errorf("error unassigning EC2 Network Interface (%s) IPv6 addresses: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("source_dest_check") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(d.Get("source_dest_check").(bool))},
		}

		log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
		_, err := conn.ModifyNetworkInterfaceAttribute(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 Network Interface (%s) SourceDestCheck: %w", d.Id(), err)
		}
	}

	if d.HasChange("security_groups") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Groups:             flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
		_, err := conn.ModifyNetworkInterfaceAttribute(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 Network Interface (%s) Groups: %w", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Description:        &ec2.AttributeValue{Value: aws.String(d.Get("description").(string))},
		}

		log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
		_, err := conn.ModifyNetworkInterfaceAttribute(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 Network Interface (%s) Description: %w", d.Id(), err)
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

		err := DetachNetworkInterface(conn, d.Id(), attachment["attachment_id"].(string), NetworkInterfaceDetachedTimeout)

		if err != nil {
			return err
		}
	}

	return DeleteNetworkInterface(conn, d.Id())
}

func attachNetworkInterface(conn *ec2.EC2, networkInterfaceID, instanceID string, deviceIndex int, timeout time.Duration) (string, error) {
	input := &ec2.AttachNetworkInterfaceInput{
		DeviceIndex:        aws.Int64(int64(deviceIndex)),
		InstanceId:         aws.String(instanceID),
		NetworkInterfaceId: aws.String(networkInterfaceID),
	}

	log.Printf("[INFO] Attaching EC2 Network Interface: %s", input)
	output, err := conn.AttachNetworkInterface(input)

	if err != nil {
		return "", fmt.Errorf("error attaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, instanceID, err)
	}

	attachmentID := aws.StringValue(output.AttachmentId)

	_, err = WaitNetworkInterfaceAttached(conn, attachmentID, timeout)

	if err != nil {
		return attachmentID, fmt.Errorf("error waiting for EC2 Network Interface (%s/%s) attach: %w", networkInterfaceID, attachmentID, err)
	}

	return attachmentID, nil
}

func DeleteNetworkInterface(conn *ec2.EC2, networkInterfaceID string) error {
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

func DetachNetworkInterface(conn *ec2.EC2, networkInterfaceID, attachmentID string, timeout time.Duration) error {
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

func flattenNetworkInterfaceAssociation(apiObject *ec2.NetworkInterfaceAssociation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationId; v != nil {
		tfMap["allocation_id"] = aws.StringValue(v)
	}

	if v := apiObject.AssociationId; v != nil {
		tfMap["association_id"] = aws.StringValue(v)
	}

	if v := apiObject.CarrierIp; v != nil {
		tfMap["carrier_ip"] = aws.StringValue(v)
	}

	if v := apiObject.CustomerOwnedIp; v != nil {
		tfMap["customer_owned_ip"] = aws.StringValue(v)
	}

	if v := apiObject.IpOwnerId; v != nil {
		tfMap["ip_owner_id"] = aws.StringValue(v)
	}

	if v := apiObject.PublicDnsName; v != nil {
		tfMap["public_dns_name"] = aws.StringValue(v)
	}

	if v := apiObject.PublicIp; v != nil {
		tfMap["public_ip"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenNetworkInterfaceAttachment(apiObject *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttachmentId; v != nil {
		tfMap["attachment_id"] = aws.StringValue(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.Int64Value(v)
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance"] = aws.StringValue(v)
	}

	return tfMap
}

func expandPrivateIPAddressSpecification(tfString string) *ec2.PrivateIpAddressSpecification {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.PrivateIpAddressSpecification{
		PrivateIpAddress: aws.String(tfString),
	}

	return apiObject
}

func expandPrivateIPAddressSpecifications(tfList []interface{}) []*ec2.PrivateIpAddressSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.PrivateIpAddressSpecification

	for i, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandPrivateIPAddressSpecification(tfString)

		if apiObject == nil {
			continue
		}

		if i == 0 {
			apiObject.Primary = aws.Bool(true)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInstanceIPv6Address(tfString string) *ec2.InstanceIpv6Address {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.InstanceIpv6Address{
		Ipv6Address: aws.String(tfString),
	}

	return apiObject
}

func expandInstanceIPv6Addresses(tfList []interface{}) []*ec2.InstanceIpv6Address {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.InstanceIpv6Address

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandInstanceIPv6Address(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenNetworkInterfacePrivateIPAddress(apiObject *ec2.NetworkInterfacePrivateIpAddress) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.PrivateIpAddress; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func FlattenNetworkInterfacePrivateIPAddresses(apiObjects []*ec2.NetworkInterfacePrivateIpAddress) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterfacePrivateIPAddress(apiObject))
	}

	return tfList
}

func flattenNetworkInterfaceIPv6Address(apiObject *ec2.NetworkInterfaceIpv6Address) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Address; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenNetworkInterfaceIPv6Addresses(apiObjects []*ec2.NetworkInterfaceIpv6Address) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterfaceIPv6Address(apiObject))
	}

	return tfList
}

func expandIPv4PrefixSpecificationRequest(tfString string) *ec2.Ipv4PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.Ipv4PrefixSpecificationRequest{
		Ipv4Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv4PrefixSpecificationRequests(tfList []interface{}) []*ec2.Ipv4PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.Ipv4PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv4PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIPv6PrefixSpecificationRequest(tfString string) *ec2.Ipv6PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.Ipv6PrefixSpecificationRequest{
		Ipv6Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv6PrefixSpecificationRequests(tfList []interface{}) []*ec2.Ipv6PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.Ipv6PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv6PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenIPv4PrefixSpecification(apiObject *ec2.Ipv4PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv4Prefix; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenIPv4PrefixSpecifications(apiObjects []*ec2.Ipv4PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIPv4PrefixSpecification(apiObject))
	}

	return tfList
}

func flattenIPv6PrefixSpecification(apiObject *ec2.Ipv6PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Prefix; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenIPv6PrefixSpecifications(apiObjects []*ec2.Ipv6PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIPv6PrefixSpecification(apiObject))
	}

	return tfList
}
