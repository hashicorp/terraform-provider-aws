// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for VPC Endpoint creation
	VPCEndpointCreationTimeout = 10 * time.Minute
)

// @SDKResource("aws_vpc_endpoint", name="VPC Endpoint")
// @Tags(identifierAttribute="id")
func ResourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRead,
		UpdateWithoutTimeout: resourceVPCEndpointUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
						"private_dns_only_for_inbound_resolver_endpoint": {
							Type:     schema.TypeBool,
							Optional: true,
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
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

func resourceVPCEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	serviceName := d.Get("service_name").(string)
	input := &ec2.CreateVpcEndpointInput{
		ClientToken:       aws.String(id.UniqueId()),
		PrivateDnsEnabled: aws.Bool(d.Get("private_dns_enabled").(bool)),
		ServiceName:       aws.String(serviceName),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeVpcEndpoint),
		VpcEndpointType:   aws.String(d.Get("vpc_endpoint_type").(string)),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("dns_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		// PrivateDnsOnlyForInboundResolverEndpoint is only supported for services
		// that support both gateway and interface endpoints, i.e. S3.
		if isAmazonS3VPCEndpoint(serviceName) {
			input.DnsOptions = expandDNSOptionsSpecificationWithPrivateDNSOnly(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.DnsOptions = expandDNSOptionsSpecification(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PolicyDocument = aws.String(policy)
	}

	if v, ok := d.GetOk("route_table_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.RouteTableIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateVpcEndpointWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.TagSpecifications != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.TagSpecifications = nil
		output, err = conn.CreateVpcEndpointWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC Endpoint (%s): %s", serviceName, err)
	}

	vpce := output.VpcEndpoint
	d.SetId(aws.StringValue(vpce.VpcEndpointId))

	if d.Get("auto_accept").(bool) && aws.StringValue(vpce.State) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(ctx, conn, d.Id(), aws.StringValue(vpce.ServiceName), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, err = WaitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) create: %s", serviceName, err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.TagSpecifications == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EC2 VPC Endpoint (%s) tags: %s", serviceName, err)
		}
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	vpce, err := FindVPCEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint (%s): %s", d.Id(), err)
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
		return sdkdiag.AppendErrorf(diags, "setting dns_entry: %s", err)
	}
	if vpce.DnsOptions != nil {
		if err := d.Set("dns_options", []interface{}{flattenDNSOptions(vpce.DnsOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_options: %s", err)
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

	if pl, err := FindPrefixListByName(ctx, conn, serviceName); err != nil {
		if tfresource.NotFound(err) {
			d.Set("cidr_blocks", nil)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Prefix List (%s): %s", serviceName, err)
		}
	} else {
		d.Set("cidr_blocks", aws.StringValueSlice(pl.Cidrs))
		d.Set("prefix_list_id", pl.PrefixListId)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(vpce.PolicyDocument))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("policy", policyToSet)

	setTagsOut(ctx, vpce.Tags)

	return diags
}

func resourceVPCEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChange("auto_accept") && d.Get("auto_accept").(bool) && d.Get("state").(string) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(ctx, conn, d.Id(), d.Get("service_name").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("dns_options", "ip_address_type", "policy", "private_dns_enabled", "security_group_ids", "route_table_ids", "subnet_ids") {
		privateDNSEnabled := d.Get("private_dns_enabled").(bool)
		input := &ec2.ModifyVpcEndpointInput{
			VpcEndpointId: aws.String(d.Id()),
		}

		if d.HasChange("dns_options") {
			if v, ok := d.GetOk("dns_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				// PrivateDnsOnlyForInboundResolverEndpoint is only supported for services
				// that support both gateway and interface endpoints, i.e. S3.
				if isAmazonS3VPCEndpoint(d.Get("service_name").(string)) {
					input.DnsOptions = expandDNSOptionsSpecificationWithPrivateDNSOnly(tfMap)
				} else {
					input.DnsOptions = expandDNSOptionsSpecification(tfMap)
				}
			}
		}

		if d.HasChange("ip_address_type") {
			input.IpAddressType = aws.String(d.Get("ip_address_type").(string))
		}

		if d.HasChange("private_dns_enabled") {
			input.PrivateDnsEnabled = aws.Bool(privateDNSEnabled)
		}

		input.AddRouteTableIds, input.RemoveRouteTableIds = flattenAddAndRemoveStringLists(d, "route_table_ids")
		input.AddSecurityGroupIds, input.RemoveSecurityGroupIds = flattenAddAndRemoveStringLists(d, "security_group_ids")
		input.AddSubnetIds, input.RemoveSubnetIds = flattenAddAndRemoveStringLists(d, "subnet_ids")

		if d.HasChange("policy") {
			o, n := d.GetChange("policy")

			if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
				policy, err := structure.NormalizeJsonString(d.Get("policy"))

				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				if policy == "" {
					input.ResetPolicy = aws.Bool(true)
				} else {
					input.PolicyDocument = aws.String(policy)
				}
			}
		}

		_, err := conn.ModifyVpcEndpointWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := WaitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 VPC Endpoint: %s", d.Id())
	output, err := conn.DeleteVpcEndpointsWithContext(ctx, &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC Endpoint (%s): %s", d.Id(), err)
	}

	if _, err = WaitVPCEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func vpcEndpointAccept(ctx context.Context, conn *ec2.EC2, vpceID, serviceName string, timeout time.Duration) error {
	serviceConfiguration, err := FindVPCEndpointServiceConfigurationByServiceName(ctx, conn, serviceName)

	if err != nil {
		return fmt.Errorf("reading EC2 VPC Endpoint Service Configuration (%s): %w", serviceName, err)
	}

	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      serviceConfiguration.ServiceId,
		VpcEndpointIds: aws.StringSlice([]string{vpceID}),
	}

	_, err = conn.AcceptVpcEndpointConnectionsWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("accepting EC2 VPC Endpoint (%s) connection: %w", vpceID, err)
	}

	if _, err = WaitVPCEndpointAccepted(ctx, conn, vpceID, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 VPC Endpoint (%s) acceptance: %w", vpceID, err)
	}

	return nil
}

func isAmazonS3VPCEndpoint(serviceName string) bool {
	ok, _ := regexp.MatchString("com\\.amazonaws\\.([a-z]+\\-[a-z]+\\-[0-9])\\.s3", serviceName)
	return ok
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

func expandDNSOptionsSpecificationWithPrivateDNSOnly(tfMap map[string]interface{}) *ec2.DnsOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.DnsOptionsSpecification{}

	if v, ok := tfMap["dns_record_ip_type"].(string); ok && v != "" {
		apiObject.DnsRecordIpType = aws.String(v)
	}

	if v, ok := tfMap["private_dns_only_for_inbound_resolver_endpoint"].(bool); ok {
		apiObject.PrivateDnsOnlyForInboundResolverEndpoint = aws.Bool(v)
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

	if v := apiObject.PrivateDnsOnlyForInboundResolverEndpoint; v != nil {
		tfMap["private_dns_only_for_inbound_resolver_endpoint"] = aws.BoolValue(v)
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

func flattenAddAndRemoveStringLists(d *schema.ResourceData, key string) ([]*string, []*string) {
	if !d.HasChange(key) {
		return nil, nil
	}

	var add, del []*string

	o, n := d.GetChange(key)
	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	if v := flex.ExpandStringSet(ns.Difference(os)); len(v) > 0 {
		add = v
	}

	if v := flex.ExpandStringSet(os.Difference(ns)); len(v) > 0 {
		del = v
	}

	return add, del
}
