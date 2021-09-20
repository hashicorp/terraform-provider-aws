package elbv2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerCreate,
		Read:   resourceLoadBalancerRead,
		Update: resourceLoadBalancerUpdate,
		Delete: resourceLoadBalancerDelete,
		// Subnets are ForceNew for Network Load Balancers
		CustomizeDiff: customdiff.Sequence(
			customizeDiffNLBSubnets,
			verify.SetTagsDiff,
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(loadBalancerCreateTimeout),
			Update: schema.DefaultTimeout(loadBalancerUpdateTimeout),
			Delete: schema.DefaultTimeout(loadBalancerDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validName,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validNamePrefix,
			},

			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"load_balancer_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      elbv2.LoadBalancerTypeEnumApplication,
				ValidateFunc: validation.StringInSlice(elbv2.LoadBalancerTypeEnum_Values(), false),
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
				Set:      schema.HashString,
			},

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},

			"subnet_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"ipv6_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsIPv6Address,
						},
						"outpost_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allocation_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"private_ipv4_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
					if m["allocation_id"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["allocation_id"].(string)))
					}
					if m["private_ipv4_address"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["private_ipv4_address"].(string)))
					}
					return create.StringHashcode(buf.String())
				},
			},

			"access_logs": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"idle_timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          60,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumNetwork),
			},

			"drop_invalid_header_fields": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType("network"),
			},

			"enable_cross_zone_load_balancing": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumApplication),
			},

			"enable_http2": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumNetwork),
			},

			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					elbv2.IpAddressTypeIpv4,
					elbv2.IpAddressTypeDualstack,
				}, false),
			},

			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func suppressIfLBType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return d.Get("load_balancer_type").(string) == t
	}
}

func resourceLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.PrefixedUniqueId("tf-lb-")
	}
	d.Set("name", name)

	elbOpts := &elbv2.CreateLoadBalancerInput{
		Name: aws.String(name),
		Type: aws.String(d.Get("load_balancer_type").(string)),
	}

	if len(tags) > 0 {
		elbOpts.Tags = tags.IgnoreAws().Elbv2Tags()
	}

	if _, ok := d.GetOk("internal"); ok {
		elbOpts.Scheme = aws.String("internal")
	}

	if v, ok := d.GetOk("security_groups"); ok {
		elbOpts.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnets"); ok {
		elbOpts.Subnets = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_mapping"); ok {
		rawMappings := v.(*schema.Set).List()
		elbOpts.SubnetMappings = make([]*elbv2.SubnetMapping, len(rawMappings))
		for i, mapping := range rawMappings {
			subnetMap := mapping.(map[string]interface{})

			elbOpts.SubnetMappings[i] = &elbv2.SubnetMapping{
				SubnetId: aws.String(subnetMap["subnet_id"].(string)),
			}

			if subnetMap["allocation_id"].(string) != "" {
				elbOpts.SubnetMappings[i].AllocationId = aws.String(subnetMap["allocation_id"].(string))
			}

			if subnetMap["private_ipv4_address"].(string) != "" {
				elbOpts.SubnetMappings[i].PrivateIPv4Address = aws.String(subnetMap["private_ipv4_address"].(string))
			}

			if subnetMap["ipv6_address"].(string) != "" {
				elbOpts.SubnetMappings[i].IPv6Address = aws.String(subnetMap["ipv6_address"].(string))
			}
		}
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		elbOpts.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		elbOpts.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	log.Printf("[DEBUG] ALB create configuration: %#v", elbOpts)

	resp, err := conn.CreateLoadBalancer(elbOpts)
	if err != nil {
		return fmt.Errorf("error creating %s Load Balancer: %w", d.Get("load_balancer_type").(string), err)
	}

	if len(resp.LoadBalancers) != 1 {
		return fmt.Errorf("no load balancers returned following creation of %s", d.Get("name").(string))
	}

	lb := resp.LoadBalancers[0]
	d.SetId(aws.StringValue(lb.LoadBalancerArn))
	log.Printf("[INFO] LB ID: %s", d.Id())

	_, err = waitLoadBalancerActive(conn, aws.StringValue(lb.LoadBalancerArn), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error waiting for Load Balancer (%s) to be active: %w", d.Get("name").(string), err)
	}

	return resourceLoadBalancerUpdate(d, meta)
}

func resourceLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	lb, err := FindLoadBalancerByARN(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
		// The ALB is gone now, so just remove it from the state
		log.Printf("[WARN] ALB %s not found in AWS, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving ALB (%s): %w", d.Id(), err)
	}

	if lb == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error retrieving ALB (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] ALB %s not found in AWS, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return flattenAwsLbResource(d, meta, lb)
}

func resourceLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := resource.Retry(loadBalancerTagPropagationTimeout, func() *resource.RetryError {
			err := tftags.Elbv2UpdateTags(conn, d.Id(), o, n)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
				log.Printf("[DEBUG] Retrying tagging of LB (%s) after error: %s", d.Id(), err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			err = tftags.Elbv2UpdateTags(conn, d.Id(), o, n)
		}

		if err != nil {
			return fmt.Errorf("error updating LB (%s) tags: %w", d.Id(), err)
		}
	}

	attributes := make([]*elbv2.LoadBalancerAttribute, 0)

	if d.HasChange("access_logs") {
		logs := d.Get("access_logs").([]interface{})

		if len(logs) == 1 && logs[0] != nil {
			log := logs[0].(map[string]interface{})

			enabled := log["enabled"].(bool)

			attributes = append(attributes,
				&elbv2.LoadBalancerAttribute{
					Key:   aws.String("access_logs.s3.enabled"),
					Value: aws.String(strconv.FormatBool(enabled)),
				})
			if enabled {
				attributes = append(attributes,
					&elbv2.LoadBalancerAttribute{
						Key:   aws.String("access_logs.s3.bucket"),
						Value: aws.String(log["bucket"].(string)),
					},
					&elbv2.LoadBalancerAttribute{
						Key:   aws.String("access_logs.s3.prefix"),
						Value: aws.String(log["prefix"].(string)),
					})
			}
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("access_logs.s3.enabled"),
				Value: aws.String("false"),
			})
		}
	}

	switch d.Get("load_balancer_type").(string) {
	case elbv2.LoadBalancerTypeEnumApplication:
		if d.HasChange("idle_timeout") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("idle_timeout.timeout_seconds"),
				Value: aws.String(fmt.Sprintf("%d", d.Get("idle_timeout").(int))),
			})
		}
		if d.HasChange("enable_http2") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http2.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("enable_http2").(bool))),
			})
		}

		if d.HasChange("drop_invalid_header_fields") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("routing.http.drop_invalid_header_fields.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("drop_invalid_header_fields").(bool))),
			})
		}

	case elbv2.LoadBalancerTypeEnumGateway, elbv2.LoadBalancerTypeEnumNetwork:
		if d.HasChange("enable_cross_zone_load_balancing") || d.IsNewResource() {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String("load_balancing.cross_zone.enabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("enable_cross_zone_load_balancing").(bool))),
			})
		}
	}

	if d.HasChange("enable_deletion_protection") || d.IsNewResource() {
		attributes = append(attributes, &elbv2.LoadBalancerAttribute{
			Key:   aws.String("deletion_protection.enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("enable_deletion_protection").(bool))),
		})
	}

	if len(attributes) != 0 {
		input := &elbv2.ModifyLoadBalancerAttributesInput{
			LoadBalancerArn: aws.String(d.Id()),
			Attributes:      attributes,
		}

		log.Printf("[DEBUG] ALB Modify Load Balancer Attributes Request: %#v", input)
		_, err := conn.ModifyLoadBalancerAttributes(input)
		if err != nil {
			return fmt.Errorf("failure configuring LB attributes: %w", err)
		}
	}

	if d.HasChange("security_groups") {
		sgs := flex.ExpandStringSet(d.Get("security_groups").(*schema.Set))

		params := &elbv2.SetSecurityGroupsInput{
			LoadBalancerArn: aws.String(d.Id()),
			SecurityGroups:  sgs,
		}
		_, err := conn.SetSecurityGroups(params)
		if err != nil {
			return fmt.Errorf("failure Setting LB Security Groups: %w", err)
		}

	}

	// subnets are assigned at Create; the 'change' here is an empty map for old
	// and current subnets for new, so this change is redundant when the
	// resource is just created, so we don't attempt if it is a newly created
	// resource.
	if d.HasChange("subnets") && !d.IsNewResource() {
		subnets := flex.ExpandStringSet(d.Get("subnets").(*schema.Set))

		params := &elbv2.SetSubnetsInput{
			LoadBalancerArn: aws.String(d.Id()),
			Subnets:         subnets,
		}

		_, err := conn.SetSubnets(params)
		if err != nil {
			return fmt.Errorf("failure Setting LB Subnets: %w", err)
		}
	}

	if d.HasChange("ip_address_type") {

		params := &elbv2.SetIpAddressTypeInput{
			LoadBalancerArn: aws.String(d.Id()),
			IpAddressType:   aws.String(d.Get("ip_address_type").(string)),
		}

		_, err := conn.SetIpAddressType(params)
		if err != nil {
			return fmt.Errorf("failure Setting LB IP Address Type: %w", err)
		}
	}

	_, err := waitLoadBalancerActive(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for Load Balancer (%s) to be active: %w", d.Get("name").(string), err)
	}

	return resourceLoadBalancerRead(d, meta)
}

func resourceLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	log.Printf("[INFO] Deleting LB: %s", d.Id())

	// Destroy the load balancer
	deleteElbOpts := elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(d.Id()),
	}
	if _, err := conn.DeleteLoadBalancer(&deleteElbOpts); err != nil {
		return fmt.Errorf("error deleting LB: %w", err)
	}

	ec2conn := meta.(*conns.AWSClient).EC2Conn

	err := cleanupLBNetworkInterfaces(ec2conn, d.Id())
	if err != nil {
		log.Printf("[WARN] Failed to cleanup ENIs for ALB %q: %#v", d.Id(), err)
	}

	err = waitForNLBNetworkInterfacesToDetach(ec2conn, d.Id())
	if err != nil {
		log.Printf("[WARN] Failed to wait for ENIs to disappear for NLB %q: %#v", d.Id(), err)
	}

	return nil
}

// ALB automatically creates ENI(s) on creation
// but the cleanup is asynchronous and may take time
// which then blocks IGW, SG or VPC on deletion
// So we make the cleanup "synchronous" here
func cleanupLBNetworkInterfaces(conn *ec2.EC2, lbArn string) error {
	name, err := getLbNameFromArn(lbArn)
	if err != nil {
		return err
	}

	out, err := conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-owner-id"),
				Values: []*string{aws.String("amazon-elb")},
			},
			{
				Name:   aws.String("description"),
				Values: []*string{aws.String("ELB " + name)},
			},
		},
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found %d ENIs to cleanup for LB %q",
		len(out.NetworkInterfaces), name)

	if len(out.NetworkInterfaces) == 0 {
		// Nothing to cleanup
		return nil
	}

	err = detachNetworkInterfaces(conn, out.NetworkInterfaces)
	if err != nil {
		return err
	}

	err = deleteNetworkInterfaces(conn, out.NetworkInterfaces)

	return err
}

func waitForNLBNetworkInterfacesToDetach(conn *ec2.EC2, lbArn string) error {
	name, err := getLbNameFromArn(lbArn)
	if err != nil {
		return err
	}

	// We cannot cleanup these ENIs ourselves as that would result in
	// OperationNotPermitted: You are not allowed to manage 'ela-attach' attachments.
	// yet presence of these ENIs may prevent us from deleting EIPs associated w/ the NLB
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-owner-id"),
				Values: []*string{aws.String("amazon-aws")},
			},
			{
				Name:   aws.String("attachment.attachment-id"),
				Values: []*string{aws.String("ela-attach-*")},
			},
			{
				Name:   aws.String("description"),
				Values: []*string{aws.String("ELB " + name)},
			},
		},
	}
	var out *ec2.DescribeNetworkInterfacesOutput
	err = resource.Retry(loadBalancerNetworkInterfaceDetachTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeNetworkInterfaces(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		niCount := len(out.NetworkInterfaces)
		if niCount > 0 {
			log.Printf("[DEBUG] Found %d ENIs to cleanup for NLB %q", niCount, lbArn)
			return resource.RetryableError(fmt.Errorf("waiting for %d ENIs of %q to clean up", niCount, lbArn))
		}
		log.Printf("[DEBUG] ENIs gone for NLB %q", lbArn)

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.DescribeNetworkInterfaces(input)
		if err != nil {
			return fmt.Errorf("error describing network inferfaces: %w", err)
		}

		niCount := len(out.NetworkInterfaces)
		if niCount > 0 {
			return fmt.Errorf("error waiting for %d ENIs of %q to clean up", niCount, lbArn)
		}
	}

	if err != nil {
		return fmt.Errorf("error describing network inferfaces: %w", err)
	}

	return nil
}

func getLbNameFromArn(arn string) (string, error) {
	re := regexp.MustCompile("([^/]+/[^/]+/[^/]+)$")
	matches := re.FindStringSubmatch(arn)
	if len(matches) != 2 {
		return "", fmt.Errorf("unexpected ARN format: %q", arn)
	}

	// e.g. app/example-alb/b26e625cdde161e6
	return matches[1], nil
}

// flattenSubnetsFromAvailabilityZones creates a slice of strings containing the subnet IDs
// for the ALB based on the AvailabilityZones structure returned by the API.
func flattenSubnetsFromAvailabilityZones(availabilityZones []*elbv2.AvailabilityZone) []string {
	var result []string
	for _, az := range availabilityZones {
		result = append(result, aws.StringValue(az.SubnetId))
	}
	return result
}

func flattenSubnetMappingsFromAvailabilityZones(availabilityZones []*elbv2.AvailabilityZone) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)
	for _, availabilityZone := range availabilityZones {
		m := make(map[string]interface{})
		m["subnet_id"] = aws.StringValue(availabilityZone.SubnetId)
		m["outpost_id"] = aws.StringValue(availabilityZone.OutpostId)

		for _, loadBalancerAddress := range availabilityZone.LoadBalancerAddresses {
			m["allocation_id"] = aws.StringValue(loadBalancerAddress.AllocationId)
			m["private_ipv4_address"] = aws.StringValue(loadBalancerAddress.PrivateIPv4Address)
			m["ipv6_address"] = aws.StringValue(loadBalancerAddress.IPv6Address)
		}

		l = append(l, m)
	}
	return l
}

func lbSuffixFromARN(arn *string) string {
	if arn == nil {
		return ""
	}

	if arnComponents := regexp.MustCompile(`arn:.*:loadbalancer/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return arnComponents[0][1]
		}
	}

	return ""
}

// flattenAwsLbResource takes a *elbv2.LoadBalancer and populates all respective resource fields.
func flattenAwsLbResource(d *schema.ResourceData, meta interface{}, lb *elbv2.LoadBalancer) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	d.Set("arn", lb.LoadBalancerArn)
	d.Set("arn_suffix", lbSuffixFromARN(lb.LoadBalancerArn))
	d.Set("name", lb.LoadBalancerName)
	d.Set("internal", lb.Scheme != nil && aws.StringValue(lb.Scheme) == "internal")
	d.Set("security_groups", flex.FlattenStringList(lb.SecurityGroups))
	d.Set("vpc_id", lb.VpcId)
	d.Set("zone_id", lb.CanonicalHostedZoneId)
	d.Set("dns_name", lb.DNSName)
	d.Set("ip_address_type", lb.IpAddressType)
	d.Set("load_balancer_type", lb.Type)
	d.Set("customer_owned_ipv4_pool", lb.CustomerOwnedIpv4Pool)

	if err := d.Set("subnets", flattenSubnetsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("error setting subnets: %w", err)
	}

	if err := d.Set("subnet_mapping", flattenSubnetMappingsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("error setting subnet_mapping: %w", err)
	}

	tags, err := tftags.Elbv2ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	attributesResp, err := conn.DescribeLoadBalancerAttributes(&elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error retrieving LB Attributes: %w", err)
	}

	accessLogMap := map[string]interface{}{
		"bucket":  "",
		"enabled": false,
		"prefix":  "",
	}

	for _, attr := range attributesResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "access_logs.s3.enabled":
			accessLogMap["enabled"] = aws.StringValue(attr.Value) == "true"
		case "access_logs.s3.bucket":
			accessLogMap["bucket"] = aws.StringValue(attr.Value)
		case "access_logs.s3.prefix":
			accessLogMap["prefix"] = aws.StringValue(attr.Value)
		case "idle_timeout.timeout_seconds":
			timeout, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error parsing ALB timeout: %w", err)
			}
			log.Printf("[DEBUG] Setting ALB Timeout Seconds: %d", timeout)
			d.Set("idle_timeout", timeout)
		case "routing.http.drop_invalid_header_fields.enabled":
			dropInvalidHeaderFieldsEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting LB Invalid Header Fields Enabled: %t", dropInvalidHeaderFieldsEnabled)
			d.Set("drop_invalid_header_fields", dropInvalidHeaderFieldsEnabled)
		case "deletion_protection.enabled":
			protectionEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting LB Deletion Protection Enabled: %t", protectionEnabled)
			d.Set("enable_deletion_protection", protectionEnabled)
		case "routing.http2.enabled":
			http2Enabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting ALB HTTP/2 Enabled: %t", http2Enabled)
			d.Set("enable_http2", http2Enabled)
		case "load_balancing.cross_zone.enabled":
			crossZoneLbEnabled := aws.StringValue(attr.Value) == "true"
			log.Printf("[DEBUG] Setting NLB Cross Zone Load Balancing Enabled: %t", crossZoneLbEnabled)
			d.Set("enable_cross_zone_load_balancing", crossZoneLbEnabled)
		}
	}

	if err := d.Set("access_logs", []interface{}{accessLogMap}); err != nil {
		return fmt.Errorf("error setting access_logs: %w", err)
	}

	return nil
}

// Load balancers of type 'network' cannot have their subnets updated at
// this time. If the type is 'network' and subnets have changed, mark the
// diff as a ForceNew operation
func customizeDiffNLBSubnets(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// The current criteria for determining if the operation should be ForceNew:
	// - lb of type "network"
	// - existing resource (id is not "")
	// - there are actual changes to be made in the subnets
	//
	// Any other combination should be treated as normal. At this time, subnet
	// handling is the only known difference between Network Load Balancers and
	// Application Load Balancers, so the logic below is simple individual checks.
	// If other differences arise we'll want to refactor to check other
	// conditions in combinations, but for now all we handle is subnets
	if lbType := diff.Get("load_balancer_type").(string); lbType != elbv2.LoadBalancerTypeEnumNetwork {
		return nil
	}

	if diff.Id() == "" {
		return nil
	}

	o, n := diff.GetChange("subnets")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	remove := os.Difference(ns).List()
	add := ns.Difference(os).List()
	if len(remove) > 0 || len(add) > 0 {
		if err := diff.SetNew("subnets", n); err != nil {
			return err
		}

		if err := diff.ForceNew("subnets"); err != nil {
			return err
		}
	}
	return nil
}

func deleteNetworkInterfaces(conn *ec2.EC2, nis []*ec2.NetworkInterface) error {
	log.Printf("[DEBUG] Trying to delete %d leftover ENIs", len(nis))
	for _, ni := range nis {
		_, err := conn.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: ni.NetworkInterfaceId,
		})
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "InvalidNetworkInterfaceID.NotFound" {
				log.Printf("[DEBUG] ENI %s is already deleted", *ni.NetworkInterfaceId)
				continue
			}
			return err
		}
	}
	return nil
}

func detachNetworkInterfaces(conn *ec2.EC2, nis []*ec2.NetworkInterface) error {
	log.Printf("[DEBUG] Trying to detach %d leftover ENIs", len(nis))
	for _, ni := range nis {
		if ni.Attachment == nil {
			log.Printf("[DEBUG] ENI %s is already detached", *ni.NetworkInterfaceId)
			continue
		}
		_, err := conn.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: ni.Attachment.AttachmentId,
			Force:        aws.Bool(true),
		})
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "InvalidAttachmentID.NotFound" {
				log.Printf("[DEBUG] ENI %s is already detached", *ni.NetworkInterfaceId)
				continue
			}
			return err
		}

		log.Printf("[DEBUG] Waiting for ENI (%s) to become detached", *ni.NetworkInterfaceId)
		stateConf := &resource.StateChangeConf{
			Pending: []string{"true"},
			Target:  []string{"false"},
			Refresh: networkInterfaceAttachmentRefreshFunc(conn, *ni.NetworkInterfaceId),
			Timeout: 10 * time.Minute,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "InvalidNetworkInterfaceID.NotFound" {
				continue
			}
			return fmt.Errorf(
				"Error waiting for ENI (%s) to become detached: %s", *ni.NetworkInterfaceId, err)
		}
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
