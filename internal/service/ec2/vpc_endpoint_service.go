// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_service", name="VPC Endpoint Service")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPCEndpointService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointServiceCreate,
		ReadWithoutTimeout:   resourceVPCEndpointServiceRead,
		UpdateWithoutTimeout: resourceVPCEndpointServiceUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"allowed_principals": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"gateway_load_balancer_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network_load_balancer_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"private_dns_name_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supported_ip_address_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.ServiceConnectivityType](),
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCEndpointServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpcEndpointServiceConfigurationInput{
		AcceptanceRequired: aws.Bool(d.Get("acceptance_required").(bool)),
		ClientToken:        aws.String(id.UniqueId()),
		TagSpecifications:  getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcEndpointService),
	}

	if v, ok := d.GetOk("gateway_load_balancer_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.GatewayLoadBalancerArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("network_load_balancer_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.NetworkLoadBalancerArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("private_dns_name"); ok {
		input.PrivateDnsName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_ip_address_types"); ok && v.(*schema.Set).Len() > 0 {
		input.SupportedIpAddressTypes = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 VPC Endpoint Service: %v", input)
	output, err := conn.CreateVpcEndpointServiceConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC Endpoint Service: %s", err)
	}

	d.SetId(aws.ToString(output.ServiceConfiguration.ServiceId))

	if _, err := waitVPCEndpointServiceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint Service (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("allowed_principals"); ok && v.(*schema.Set).Len() > 0 {
		input := &ec2.ModifyVpcEndpointServicePermissionsInput{
			AddAllowedPrincipals: flex.ExpandStringValueSet(v.(*schema.Set)),
			ServiceId:            aws.String(d.Id()),
		}

		if _, err := conn.ModifyVpcEndpointServicePermissions(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 VPC Endpoint Service (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCEndpointServiceRead(ctx, d, meta)...)
}

func resourceVPCEndpointServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	svcCfg, err := findVPCEndpointServiceConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Endpoint Service %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoint Service (%s): %s", d.Id(), err)
	}

	d.Set("acceptance_required", svcCfg.AcceptanceRequired)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-endpoint-service/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrAvailabilityZones, svcCfg.AvailabilityZones)
	d.Set("base_endpoint_dns_names", svcCfg.BaseEndpointDnsNames)
	d.Set("gateway_load_balancer_arns", svcCfg.GatewayLoadBalancerArns)
	d.Set("manages_vpc_endpoints", svcCfg.ManagesVpcEndpoints)
	d.Set("network_load_balancer_arns", svcCfg.NetworkLoadBalancerArns)
	d.Set("private_dns_name", svcCfg.PrivateDnsName)
	// The EC2 API can return a XML structure with no elements.
	if tfMap := flattenPrivateDNSNameConfiguration(svcCfg.PrivateDnsNameConfiguration); len(tfMap) > 0 {
		if err := d.Set("private_dns_name_configuration", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting private_dns_name_configuration: %s", err)
		}
	} else {
		d.Set("private_dns_name_configuration", nil)
	}
	d.Set(names.AttrServiceName, svcCfg.ServiceName)
	if len(svcCfg.ServiceType) > 0 {
		d.Set("service_type", svcCfg.ServiceType[0].ServiceType)
	} else {
		d.Set("service_type", nil)
	}
	d.Set(names.AttrState, svcCfg.ServiceState)
	d.Set("supported_ip_address_types", svcCfg.SupportedIpAddressTypes)

	setTagsOut(ctx, svcCfg.Tags)

	allowedPrincipals, err := findVPCEndpointServicePermissionsByServiceID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoint Service (%s) permissions: %s", d.Id(), err)
	}

	d.Set("allowed_principals", flattenAllowedPrincipals(allowedPrincipals))

	return diags
}

func resourceVPCEndpointServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("acceptance_required", "gateway_load_balancer_arns", "network_load_balancer_arns", "private_dns_name", "supported_ip_address_types") {
		input := &ec2.ModifyVpcEndpointServiceConfigurationInput{
			ServiceId: aws.String(d.Id()),
		}

		if d.HasChange("acceptance_required") {
			input.AcceptanceRequired = aws.Bool(d.Get("acceptance_required").(bool))
		}

		input.AddGatewayLoadBalancerArns, input.RemoveGatewayLoadBalancerArns = flattenAddAndRemoveStringValueLists(d, "gateway_load_balancer_arns")
		input.AddNetworkLoadBalancerArns, input.RemoveNetworkLoadBalancerArns = flattenAddAndRemoveStringValueLists(d, "network_load_balancer_arns")

		if d.HasChange("private_dns_name") {
			input.PrivateDnsName = aws.String(d.Get("private_dns_name").(string))
		}

		input.AddSupportedIpAddressTypes, input.RemoveSupportedIpAddressTypes = flattenAddAndRemoveStringValueLists(d, "supported_ip_address_types")

		log.Printf("[DEBUG] Updating EC2 VPC Endpoint Service: %v", input)
		_, err := conn.ModifyVpcEndpointServiceConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC Endpoint Service (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCEndpointServiceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint Service (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("allowed_principals") {
		input := &ec2.ModifyVpcEndpointServicePermissionsInput{
			ServiceId: aws.String(d.Id()),
		}

		input.AddAllowedPrincipals, input.RemoveAllowedPrincipals = flattenAddAndRemoveStringValueLists(d, "allowed_principals")

		if _, err := conn.ModifyVpcEndpointServicePermissions(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 VPC Endpoint Service (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCEndpointServiceRead(ctx, d, meta)...)
}

func resourceVPCEndpointServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VPC Endpoint Service: %s", d.Id())
	output, err := conn.DeleteVpcEndpointServiceConfigurations(ctx, &ec2.DeleteVpcEndpointServiceConfigurationsInput{
		ServiceIds: []string{d.Id()},
	})

	if err == nil && output != nil {
		err = unsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC Endpoint Service (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCEndpointServiceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Endpoint Service (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func flattenAllowedPrincipals(apiObjects []awstypes.AllowedPrincipal) []*string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Principal)
	}

	return tfList
}

func flattenPrivateDNSNameConfiguration(apiObject *awstypes.PrivateDnsNameConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.State; v != "" {
		tfMap[names.AttrState] = string(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}
