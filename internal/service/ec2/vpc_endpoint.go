// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for VPC Endpoint creation.
	vpcEndpointCreationTimeout = 10 * time.Minute
)

// @SDKResource("aws_vpc_endpoint", name="VPC Endpoint")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRead,
		UpdateWithoutTimeout: resourceVPCEndpointUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
						names.AttrDNSName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DnsRecordIpType](),
						},
						"private_dns_only_for_inbound_resolver_endpoint": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			names.AttrIPAddressType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpAddressType](),
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
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
				Computed: true,
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
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv4": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv6": {
							Type:     schema.TypeString,
							Optional: true,
						},
						// subnet_id in subnet_configuration must have a corresponding subnet in subnet_ids attribute
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_endpoint_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.VpcEndpointTypeGateway,
				ValidateDiagFunc: enum.Validate[awstypes.VpcEndpointType](),
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(vpcEndpointCreationTimeout),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	partition := meta.(*conns.AWSClient).Partition

	serviceName := d.Get(names.AttrServiceName).(string)
	input := &ec2.CreateVpcEndpointInput{
		ClientToken:       aws.String(id.UniqueId()),
		PrivateDnsEnabled: aws.Bool(d.Get("private_dns_enabled").(bool)),
		ServiceName:       aws.String(serviceName),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcEndpoint),
		VpcEndpointType:   awstypes.VpcEndpointType(d.Get("vpc_endpoint_type").(string)),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
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

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		input.IpAddressType = awstypes.IpAddressType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		policy, err := structure.NormalizeJsonString(v)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PolicyDocument = aws.String(policy)
	}

	if v, ok := d.GetOk("route_table_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.RouteTableIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetConfigurations = expandSubnetConfigurations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrSubnetIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateVpcEndpoint(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.TagSpecifications != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.TagSpecifications = nil
		output, err = conn.CreateVpcEndpoint(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC Endpoint (%s): %s", serviceName, err)
	}

	vpce := output.VpcEndpoint
	d.SetId(aws.ToString(vpce.VpcEndpointId))

	if d.Get("auto_accept").(bool) && string(vpce.State) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(ctx, conn, d.Id(), aws.ToString(vpce.ServiceName), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if _, err := waitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) create: %s", serviceName, err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.TagSpecifications == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpce, err := findVPCEndpointByID(ctx, conn, d.Id())

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
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.ToString(vpce.OwnerId),
		Resource:  fmt.Sprintf("vpc-endpoint/%s", d.Id()),
	}.String()
	serviceName := aws.ToString(vpce.ServiceName)
	d.Set(names.AttrARN, arn)
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
	d.Set(names.AttrIPAddressType, vpce.IpAddressType)
	d.Set("network_interface_ids", vpce.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, vpce.OwnerId)
	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	d.Set("requester_managed", vpce.RequesterManaged)
	d.Set("route_table_ids", vpce.RouteTableIds)
	d.Set(names.AttrSecurityGroupIDs, flattenSecurityGroupIdentifiers(vpce.Groups))
	d.Set(names.AttrServiceName, serviceName)
	d.Set(names.AttrState, vpce.State)
	d.Set(names.AttrSubnetIDs, vpce.SubnetIds)
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if v := string(vpce.VpcEndpointType); v == "" {
		d.Set("vpc_endpoint_type", awstypes.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", v)
	}
	d.Set(names.AttrVPCID, vpce.VpcId)

	if pl, err := findPrefixListByName(ctx, conn, serviceName); err != nil {
		if tfresource.NotFound(err) {
			d.Set("cidr_blocks", nil)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Prefix List (%s): %s", serviceName, err)
		}
	} else {
		d.Set("cidr_blocks", pl.Cidrs)
		d.Set("prefix_list_id", pl.PrefixListId)
	}

	subnetConfigurations, err := findSubnetConfigurationsByNetworkInterfaceIDs(ctx, conn, vpce.NetworkInterfaceIds)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint (%s) subnet configurations: %s", d.Id(), err)
	}

	if err := d.Set("subnet_configuration", flattenSubnetConfigurations(subnetConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_configuration: %s", err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(vpce.PolicyDocument))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	setTagsOut(ctx, vpce.Tags)

	return diags
}

func resourceVPCEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("auto_accept") && d.Get("auto_accept").(bool) && d.Get(names.AttrState).(string) == vpcEndpointStatePendingAcceptance {
		if err := vpcEndpointAccept(ctx, conn, d.Id(), d.Get(names.AttrServiceName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("dns_options", names.AttrIPAddressType, names.AttrPolicy, "private_dns_enabled", names.AttrSecurityGroupIDs, "route_table_ids", "subnet_configuration", names.AttrSubnetIDs) {
		input := &ec2.ModifyVpcEndpointInput{
			VpcEndpointId: aws.String(d.Id()),
		}

		if d.HasChange("dns_options") {
			if v, ok := d.GetOk("dns_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				// PrivateDnsOnlyForInboundResolverEndpoint is only supported for services
				// that support both gateway and interface endpoints, i.e. S3.
				if isAmazonS3VPCEndpoint(d.Get(names.AttrServiceName).(string)) {
					input.DnsOptions = expandDNSOptionsSpecificationWithPrivateDNSOnly(tfMap)
				} else {
					input.DnsOptions = expandDNSOptionsSpecification(tfMap)
				}
			}
		}

		if d.HasChange(names.AttrIPAddressType) {
			input.IpAddressType = awstypes.IpAddressType(d.Get(names.AttrIPAddressType).(string))
		}

		privateDNSEnabled := d.Get("private_dns_enabled").(bool)
		if d.HasChange("private_dns_enabled") {
			input.PrivateDnsEnabled = aws.Bool(privateDNSEnabled)
		}

		input.AddRouteTableIds, input.RemoveRouteTableIds = flattenAddAndRemoveStringValueLists(d, "route_table_ids")
		input.AddSecurityGroupIds, input.RemoveSecurityGroupIds = flattenAddAndRemoveStringValueLists(d, names.AttrSecurityGroupIDs)
		input.AddSubnetIds, input.RemoveSubnetIds = flattenAddAndRemoveStringValueLists(d, names.AttrSubnetIDs)

		if d.HasChange(names.AttrPolicy) {
			o, n := d.GetChange(names.AttrPolicy)

			if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
				policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy))

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

		if d.HasChange("subnet_configuration") {
			if v, ok := d.GetOk("subnet_configuration"); ok && v.(*schema.Set).Len() > 0 {
				input.SubnetConfigurations = expandSubnetConfigurations(v.(*schema.Set).List())
			}
		}

		_, err := conn.ModifyVpcEndpoint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 VPC Endpoint: %s", d.Id())
	output, err := conn.DeleteVpcEndpoints(ctx, &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: []string{d.Id()},
	})

	if err == nil && output != nil {
		err = unsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC Endpoint (%s): %s", d.Id(), err)
	}

	if _, err = waitVPCEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func vpcEndpointAccept(ctx context.Context, conn *ec2.Client, vpceID, serviceName string, timeout time.Duration) error {
	serviceConfiguration, err := findVPCEndpointServiceConfigurationByServiceName(ctx, conn, serviceName)

	if err != nil {
		return fmt.Errorf("reading EC2 VPC Endpoint Service Configuration (%s): %w", serviceName, err)
	}

	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      serviceConfiguration.ServiceId,
		VpcEndpointIds: []string{vpceID},
	}

	_, err = conn.AcceptVpcEndpointConnections(ctx, input)

	if err != nil {
		return fmt.Errorf("accepting EC2 VPC Endpoint (%s) connection: %w", vpceID, err)
	}

	if _, err = waitVPCEndpointAccepted(ctx, conn, vpceID, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 VPC Endpoint (%s) acceptance: %w", vpceID, err)
	}

	return nil
}

type subnetConfiguration struct {
	ipv4     *string
	ipv6     *string
	subnetID *string
}

func findSubnetConfigurationsByNetworkInterfaceIDs(ctx context.Context, conn *ec2.Client, networkInterfaceIDs []string) ([]subnetConfiguration, error) {
	var output []subnetConfiguration

	for _, v := range networkInterfaceIDs {
		networkInterface, err := findNetworkInterfaceByID(ctx, conn, v)

		if err != nil {
			return nil, err
		}

		output = append(output, subnetConfiguration{
			ipv4:     networkInterface.PrivateIpAddress,
			ipv6:     networkInterface.Ipv6Address,
			subnetID: networkInterface.SubnetId,
		})
	}

	return output, nil
}

func isAmazonS3VPCEndpoint(serviceName string) bool {
	ok, _ := regexp.MatchString("com\\.amazonaws\\.([a-z]+\\-[a-z]+\\-[0-9])\\.s3", serviceName)
	return ok
}

func expandDNSOptionsSpecification(tfMap map[string]interface{}) *awstypes.DnsOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DnsOptionsSpecification{}

	if v, ok := tfMap["dns_record_ip_type"].(string); ok && v != "" {
		apiObject.DnsRecordIpType = awstypes.DnsRecordIpType(v)
	}

	return apiObject
}

func expandDNSOptionsSpecificationWithPrivateDNSOnly(tfMap map[string]interface{}) *awstypes.DnsOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DnsOptionsSpecification{}

	if v, ok := tfMap["dns_record_ip_type"].(string); ok && v != "" {
		apiObject.DnsRecordIpType = awstypes.DnsRecordIpType(v)
	}

	if v, ok := tfMap["private_dns_only_for_inbound_resolver_endpoint"].(bool); ok {
		apiObject.PrivateDnsOnlyForInboundResolverEndpoint = aws.Bool(v)
	}

	return apiObject
}

func expandSubnetConfiguration(tfMap map[string]interface{}) *awstypes.SubnetConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SubnetConfiguration{}

	if v, ok := tfMap["ipv4"].(string); ok && v != "" {
		apiObject.Ipv4 = aws.String(v)
	}

	if v, ok := tfMap["ipv6"].(string); ok && v != "" {
		apiObject.Ipv6 = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnetID].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	return apiObject
}

func expandSubnetConfigurations(tfList []interface{}) []awstypes.SubnetConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SubnetConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSubnetConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenDNSEntry(apiObject *awstypes.DnsEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsName; v != nil {
		tfMap[names.AttrDNSName] = aws.ToString(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap[names.AttrHostedZoneID] = aws.ToString(v)
	}

	return tfMap
}

func flattenDNSEntries(apiObjects []awstypes.DnsEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDNSEntry(&apiObject))
	}

	return tfList
}

func flattenDNSOptions(apiObject *awstypes.DnsOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"dns_record_ip_type": string(apiObject.DnsRecordIpType),
	}

	if v := apiObject.PrivateDnsOnlyForInboundResolverEndpoint; v != nil {
		tfMap["private_dns_only_for_inbound_resolver_endpoint"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenSecurityGroupIdentifiers(apiObjects []awstypes.SecurityGroupIdentifier) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, aws.ToString(apiObject.GroupId))
	}

	return tfList
}

func flattenAddAndRemoveStringValueLists(d *schema.ResourceData, key string) ([]string, []string) {
	if !d.HasChange(key) {
		return nil, nil
	}

	var add, del []string

	o, n := d.GetChange(key)
	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	if v := flex.ExpandStringValueSet(ns.Difference(os)); len(v) > 0 {
		add = v
	}

	if v := flex.ExpandStringValueSet(os.Difference(ns)); len(v) > 0 {
		del = v
	}

	return add, del
}

func flattenSubnetConfiguration(apiObject *subnetConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ipv4; v != nil {
		tfMap["ipv4"] = aws.ToString(v)
	}

	if v := apiObject.ipv6; v != nil {
		tfMap["ipv6"] = aws.ToString(v)
	}

	if v := apiObject.subnetID; v != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(v)
	}

	return tfMap
}

func flattenSubnetConfigurations(apiObjects []subnetConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenSubnetConfiguration(&apiObject))
	}

	return tfList
}
