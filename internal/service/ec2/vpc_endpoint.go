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
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
				Set:      schema.HashString,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
				Set:      schema.HashString,
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

	req := &ec2.CreateVpcEndpointInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		VpcEndpointType:   aws.String(d.Get("vpc_endpoint_type").(string)),
		ServiceName:       aws.String(d.Get("service_name").(string)),
		PrivateDnsEnabled: aws.Bool(d.Get("private_dns_enabled").(bool)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, "vpc-endpoint"),
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		req.PolicyDocument = aws.String(policy)
	}

	setVPCEndpointCreateList(d, "route_table_ids", &req.RouteTableIds)
	setVPCEndpointCreateList(d, "subnet_ids", &req.SubnetIds)
	setVPCEndpointCreateList(d, "security_group_ids", &req.SecurityGroupIds)

	log.Printf("[DEBUG] Creating VPC Endpoint: %#v", req)
	resp, err := conn.CreateVpcEndpoint(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC Endpoint: %s", err)
	}

	vpce := resp.VpcEndpoint
	d.SetId(aws.StringValue(vpce.VpcEndpointId))

	if d.Get("auto_accept").(bool) && aws.StringValue(vpce.State) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(conn, d.Id(), aws.StringValue(vpce.ServiceName), d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}

	_, err = WaitVPCEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to become available: %w", d.Id(), err)
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
	d.Set("arn", arn)

	serviceName := aws.StringValue(vpce.ServiceName)
	d.Set("service_name", serviceName)
	d.Set("state", vpce.State)
	d.Set("vpc_id", vpce.VpcId)

	respPl, err := conn.DescribePrefixLists(&ec2.DescribePrefixListsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-name": serviceName,
		}),
	})
	if err != nil {
		return fmt.Errorf("error reading Prefix List (%s): %s", serviceName, err)
	}
	if respPl == nil || len(respPl.PrefixLists) == 0 {
		d.Set("cidr_blocks", []interface{}{})
	} else if len(respPl.PrefixLists) > 1 {
		return fmt.Errorf("multiple prefix lists associated with the service name '%s'. Unexpected", serviceName)
	} else {
		pl := respPl.PrefixLists[0]

		d.Set("prefix_list_id", pl.PrefixListId)
		err = d.Set("cidr_blocks", flex.FlattenStringList(pl.Cidrs))
		if err != nil {
			return fmt.Errorf("error setting cidr_blocks: %s", err)
		}
	}

	err = d.Set("dns_entry", flattenVPCEndpointDNSEntries(vpce.DnsEntries))
	if err != nil {
		return fmt.Errorf("error setting dns_entry: %s", err)
	}
	err = d.Set("network_interface_ids", flex.FlattenStringSet(vpce.NetworkInterfaceIds))
	if err != nil {
		return fmt.Errorf("error setting network_interface_ids: %s", err)
	}
	d.Set("owner_id", vpce.OwnerId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(vpce.PolicyDocument))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	err = d.Set("route_table_ids", flex.FlattenStringSet(vpce.RouteTableIds))
	if err != nil {
		return fmt.Errorf("error setting route_table_ids: %s", err)
	}
	d.Set("requester_managed", vpce.RequesterManaged)
	err = d.Set("security_group_ids", flattenVPCEndpointSecurityGroupIds(vpce.Groups))
	if err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	err = d.Set("subnet_ids", flex.FlattenStringSet(vpce.SubnetIds))
	if err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if vpceType := aws.StringValue(vpce.VpcEndpointType); vpceType == "" {
		d.Set("vpc_endpoint_type", ec2.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", vpceType)
	}

	tags := KeyValueTags(vpce.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
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

	if d.HasChanges("policy", "route_table_ids", "subnet_ids", "security_group_ids", "private_dns_enabled") {
		req := &ec2.ModifyVpcEndpointInput{
			VpcEndpointId: aws.String(d.Id()),
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

	input := &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{d.Id()}),
	}

	output, err := conn.DeleteVpcEndpoints(input)

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPC Endpoint (%s): %w", d.Id(), err)
	}

	_, err = WaitVPCEndpointDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC Endpoint (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func vpcEndpointAccept(conn *ec2.EC2, vpceId, svcName string, timeout time.Duration) error {
	describeSvcReq := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}
	describeSvcReq.Filters = BuildAttributeFilterList(
		map[string]string{
			"service-name": svcName,
		},
	)

	describeSvcResp, err := conn.DescribeVpcEndpointServiceConfigurations(describeSvcReq)
	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Service (%s): %s", svcName, err)
	}
	if describeSvcResp == nil || len(describeSvcResp.ServiceConfigurations) == 0 {
		return fmt.Errorf("No matching VPC Endpoint Service found")
	}

	acceptEpReq := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      describeSvcResp.ServiceConfigurations[0].ServiceId,
		VpcEndpointIds: aws.StringSlice([]string{vpceId}),
	}

	log.Printf("[DEBUG] Accepting VPC Endpoint connection: %#v", acceptEpReq)
	_, err = conn.AcceptVpcEndpointConnections(acceptEpReq)
	if err != nil {
		return fmt.Errorf("error accepting VPC Endpoint (%s) connection: %s", vpceId, err)
	}

	_, err = WaitVPCEndpointAccepted(conn, vpceId, timeout)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to be accepted: %w", vpceId, err)
	}

	return nil
}

func setVPCEndpointCreateList(d *schema.ResourceData, key string, c *[]*string) {
	if v, ok := d.GetOk(key); ok {
		list := v.(*schema.Set)
		if list.Len() > 0 {
			*c = flex.ExpandStringSet(list)
		}
	}
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

func flattenVPCEndpointDNSEntries(dnsEntries []*ec2.DnsEntry) []interface{} {
	vDnsEntries := []interface{}{}

	for _, dnsEntry := range dnsEntries {
		vDnsEntries = append(vDnsEntries, map[string]interface{}{
			"dns_name":       aws.StringValue(dnsEntry.DnsName),
			"hosted_zone_id": aws.StringValue(dnsEntry.HostedZoneId),
		})
	}

	return vDnsEntries
}

func flattenVPCEndpointSecurityGroupIds(groups []*ec2.SecurityGroupIdentifier) *schema.Set {
	vSecurityGroupIds := []interface{}{}

	for _, group := range groups {
		vSecurityGroupIds = append(vSecurityGroupIds, aws.StringValue(group.GroupId))
	}

	return schema.NewSet(schema.HashString, vSecurityGroupIds)
}
