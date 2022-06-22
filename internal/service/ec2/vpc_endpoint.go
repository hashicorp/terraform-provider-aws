package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for VPC Endpoint creation
	VPCEndpointCreationTimeout = 10 * time.Minute
)

func ResourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointCreate,
		Read:   resourceVPCEndpointRead,
		Update: resourceVPCEndpointUpdate,
		Delete: resourceVPCEndpointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_accept": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_options": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_record_ip_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.DnsRecordIpType_Values(), false),
						},
					},
				},
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.IpAddressType_Values(), false),
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.VpcEndpointTypeGateway,
				ValidateFunc: validation.StringInSlice(ec2.VpcEndpointType_Values(), false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(VPCEndpointCreationTimeout),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	serviceName := d.Get("service_name").(string)
	input := &ec2.CreateVpcEndpointInput{
		PrivateDnsEnabled: aws.Bool(d.Get("private_dns_enabled").(bool)),
		ServiceName:       aws.String(serviceName),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpcEndpoint),
		VpcEndpointType:   aws.String(d.Get("vpc_endpoint_type").(string)),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("dns_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DnsOptions = expandDNSOptionsSpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v)

		if err != nil {
			return fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		input.PolicyDocument = aws.String(policy)
	}

	if v, ok := d.GetOk("route_table_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.RouteTableIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 VPC Endpoint: %s", input)
	output, err := conn.CreateVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("creating EC2 VPC Endpoint (%s): %w", serviceName, err)
	}

	vpce := output.VpcEndpoint
	d.SetId(aws.StringValue(vpce.VpcEndpointId))

	if d.Get("auto_accept").(bool) && aws.StringValue(vpce.State) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(conn, d.Id(), aws.StringValue(vpce.ServiceName), d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}

	if _, err = WaitVPCEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EC2 VPC Endpoint (%s) create: %w", d.Id(), err)
	}

	return resourceVPCEndpointRead(d, meta)
}

func resourceVPCEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpce, err := FindVPCEndpointByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(vpce.OwnerId),
		Resource:  fmt.Sprintf("vpc-endpoint/%s", d.Id()),
	}.String()
	serviceName := aws.StringValue(vpce.ServiceName)

	d.Set("arn", arn)
	if err := d.Set("dns_entry", flattenDNSEntries(vpce.DnsEntries)); err != nil {
		return fmt.Errorf("setting dns_entry: %w", err)
	}
	if vpce.DnsOptions != nil {
		if err := d.Set("dns_options", []interface{}{flattenDNSOptions(vpce.DnsOptions)}); err != nil {
			return fmt.Errorf("setting dns_options: %w", err)
		}
	} else {
		d.Set("dns_options", nil)
	}
	d.Set("ip_address_type", vpce.IpAddressType)
	d.Set("network_interface_ids", aws.StringValueSlice(vpce.NetworkInterfaceIds))
	d.Set("owner_id", vpce.OwnerId)
	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	d.Set("requester_managed", vpce.RequesterManaged)
	d.Set("route_table_ids", aws.StringValueSlice(vpce.RouteTableIds))
	d.Set("security_group_ids", flattenSecurityGroupIdentifiers(vpce.Groups))
	d.Set("service_name", serviceName)
	d.Set("state", vpce.State)
	d.Set("subnet_ids", aws.StringValueSlice(vpce.SubnetIds))
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if v := aws.StringValue(vpce.VpcEndpointType); v == "" {
		d.Set("vpc_endpoint_type", ec2.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", v)
	}
	d.Set("vpc_id", vpce.VpcId)

	if pl, err := FindPrefixListByName(conn, serviceName); err != nil {
		if tfresource.NotFound(err) {
			d.Set("cidr_blocks", nil)
		} else {
			return fmt.Errorf("reading EC2 Prefix List (%s): %w", serviceName, err)
		}
	} else {
		d.Set("cidr_blocks", aws.StringValueSlice(pl.Cidrs))
		d.Set("prefix_list_id", pl.PrefixListId)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(vpce.PolicyDocument))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	tags := KeyValueTags(vpce.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("auto_accept") && d.Get("auto_accept").(bool) && d.Get("state").(string) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(conn, d.Id(), d.Get("service_name").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return err
		}
	}

	if d.HasChanges("policy", "route_table_ids", "subnet_ids", "security_group_ids", "private_dns_enabled", "ip_address_type") {
		req := &ec2.ModifyVpcEndpointInput{
			VpcEndpointId: aws.String(d.Id()),
		}

		if d.HasChange("ip_address_type") {
			req.IpAddressType = aws.String(d.Get("ip_address_type").(string))
		}

		if d.HasChange("policy") {
			o, n := d.GetChange("policy")

			if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
				policy, err := structure.NormalizeJsonString(d.Get("policy"))
				if err != nil {
					return fmt.Errorf("policy contains an invalid JSON: %s", err)
				}

				if policy == "" {
					req.ResetPolicy = aws.Bool(true)
				} else {
					req.PolicyDocument = aws.String(policy)
				}
			}
		}

		setVPCEndpointUpdateLists(d, "route_table_ids", &req.AddRouteTableIds, &req.RemoveRouteTableIds)
		setVPCEndpointUpdateLists(d, "subnet_ids", &req.AddSubnetIds, &req.RemoveSubnetIds)
		setVPCEndpointUpdateLists(d, "security_group_ids", &req.AddSecurityGroupIds, &req.RemoveSecurityGroupIds)

		if d.HasChange("private_dns_enabled") {
			req.PrivateDnsEnabled = aws.Bool(d.Get("private_dns_enabled").(bool))
		}

		log.Printf("[DEBUG] Updating VPC Endpoint: %#v", req)
		if _, err := conn.ModifyVpcEndpoint(req); err != nil {
			return fmt.Errorf("Error updating VPC Endpoint: %s", err)
		}

		_, err := WaitVPCEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for VPC Endpoint (%s) to become available: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceVPCEndpointRead(d, meta)
}

func resourceVPCEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 VPC Endpoint: %s", d.Id())
	output, err := conn.DeleteVpcEndpoints(&ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 VPC Endpoint (%s): %w", d.Id(), err)
	}

	if _, err = WaitVPCEndpointDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for EC2 VPC Endpoint (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func vpcEndpointAccept(conn *ec2.EC2, vpceID, serviceName string, timeout time.Duration) error {
	serviceConfiguration, err := FindVPCEndpointServiceConfigurationByServiceName(conn, serviceName)

	if err != nil {
		return fmt.Errorf("reading EC2 VPC Endpoint Service Configuration (%s): %w", serviceName, err)
	}

	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      serviceConfiguration.ServiceId,
		VpcEndpointIds: aws.StringSlice([]string{vpceID}),
	}

	log.Printf("[DEBUG] Accepting EC2 VPC Endpoint connection: %s", input)
	_, err = conn.AcceptVpcEndpointConnections(input)

	if err != nil {
		return fmt.Errorf("accepting EC2 VPC Endpoint (%s) connection: %w", vpceID, err)
	}

	if _, err = WaitVPCEndpointAccepted(conn, vpceID, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 VPC Endpoint (%s) acceptance: %w", vpceID, err)
	}

	return nil
}

func setVPCEndpointUpdateLists(d *schema.ResourceData, key string, a, r *[]*string) {
	if d.HasChange(key) {
		o, n := d.GetChange(key)
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		add := flex.ExpandStringSet(ns.Difference(os))
		if len(add) > 0 {
			*a = add
		}

		remove := flex.ExpandStringSet(os.Difference(ns))
		if len(remove) > 0 {
			*r = remove
		}
	}
}

func expandDNSOptionsSpecification(tfMap map[string]interface{}) *ec2.DnsOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.DnsOptionsSpecification{}

	if v, ok := tfMap["dns_record_ip_type"].(string); ok && v != "" {
		apiObject.DnsRecordIpType = aws.String(v)
	}

	return apiObject
}

func flattenDNSEntry(apiObject *ec2.DnsEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsName; v != nil {
		tfMap["dns_name"] = aws.StringValue(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap["hosted_zone_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDNSEntries(apiObjects []*ec2.DnsEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDNSEntry(apiObject))
	}

	return tfList
}

func flattenDNSOptions(apiObject *ec2.DnsOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecordIpType; v != nil {
		tfMap["dns_record_ip_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSecurityGroupIdentifiers(apiObjects []*ec2.SecurityGroupIdentifier) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, aws.StringValue(apiObject.GroupId))
	}

	return tfList
}
