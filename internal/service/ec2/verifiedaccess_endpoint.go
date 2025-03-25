// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_endpoint", name="Verified Access Endpoint")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVerifiedAccessEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessEndpointCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessEndpointRead,
		UpdateWithoutTimeout: resourceVerifiedAccessEndpointUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			attrVerifiedAccessEndpoint_ApplicationDomain: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			attrVerifiedAccessEndpoint_AttachmentType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(verifiedAccessAttachmentType_Values(), false),
			},
			"cidr_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attrVerifiedAccessEndpoint_CidrOptions_Cidr: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						attrVerifiedAccessEndpoint_PortRange: { // debe estar en singular aquí también
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									attrVerifiedAccessEndpoint_PortRange_FromPort: {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
									attrVerifiedAccessEndpoint_PortRange_ToPort:   {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
								},
							},
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{verifiedAccessEndpointProtocolTCP}, false),
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			attrVerifiedAccessEndpoint_DeviceValidationDomain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			attrVerifiedAccessEndpoint_DomainCertificateArn: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			attrVerifiedAccessEndpoint_EndpointDomainPrefix: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			attrVerifiedAccessEndpoint_EndpointDomain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpointType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(verifiedAccessEndpointType_Values(), false),
			},
			attrVerifiedAccessEndpoint_LoadBalancerOptions: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attrVerifiedAccessEndpoint_LoadBalancerOptions_LoadBalancerArn: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						attrVerifiedAccessEndpoint_PortRange: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									attrVerifiedAccessEndpoint_PortRange_FromPort: {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
									attrVerifiedAccessEndpoint_PortRange_ToPort:   {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
								},
							},
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(verifiedAccessEndpointProtocol_Values(), false),
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			attrVerifiedAccessEndpoint_NetworkInterfaceOptions: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						attrVerifiedAccessEndpoint_PortRange: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									attrVerifiedAccessEndpoint_PortRange_FromPort: {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
									attrVerifiedAccessEndpoint_PortRange_ToPort:   {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IsPortNumberOrZero},
								},
							},
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(verifiedAccessEndpointProtocol_Values(), false),
						},
					},
				},
			},
			attrVerifiedAccessEndpoint_PolicyDocument: {
				Type:     schema.TypeString,
				Optional: true,
			},
			attrVerifiedAccessEndpoint_RdsOptions: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{verifiedAccessEndpointProtocolTCP}, false),
						},
						attrVerifiedAccessEndpoint_RdsOptions_ClusterArn: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						attrVerifiedAccessEndpoint_RdsOptions_InstanceArn: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						attrVerifiedAccessEndpoint_RdsOptions_ProxyArn: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrEndpoint: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			attrVerifiedAccessEndpoint_SseSpecification: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attrVerifiedAccessEndpoint_SseSpecification_CustomManagedKeyEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verified_access_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"verified_access_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVerifiedAccessEndpointCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVerifiedAccessEndpointInput{
		AttachmentType:        types.VerifiedAccessEndpointAttachmentType(d.Get(attrVerifiedAccessEndpoint_AttachmentType).(string)),
		ClientToken:           aws.String(id.UniqueId()),
		EndpointType:          types.VerifiedAccessEndpointType(d.Get(names.AttrEndpointType).(string)),
		TagSpecifications:     getTagSpecificationsIn(ctx, types.ResourceTypeVerifiedAccessEndpoint),
		VerifiedAccessGroupId: aws.String(d.Get("verified_access_group_id").(string)),
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_ApplicationDomain); ok {
		input.ApplicationDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_DomainCertificateArn); ok {
		input.DomainCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_EndpointDomainPrefix); ok {
		input.EndpointDomainPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.CidrOptions = expandCreateVerifiedAccessEndpointCidrOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_LoadBalancerOptions); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.LoadBalancerOptions = expandCreateVerifiedAccessEndpointLoadBalancerOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_NetworkInterfaceOptions); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.NetworkInterfaceOptions = expandCreateVerifiedAccessEndpointEniOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_PolicyDocument); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rds_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RdsOptions = expandCreateVerifiedAccessEndpointRdsOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(attrVerifiedAccessEndpoint_SseSpecification); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SseSpecification = expandCreateVerifiedAccessGenericSseSpecification(v.([]any)[0].(map[string]any))
	}

	output, err := conn.CreateVerifiedAccessEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Endpoint: %s", err)
	}

	d.SetId(aws.ToString(output.VerifiedAccessEndpoint.VerifiedAccessEndpointId))

	if _, err := waitVerifiedAccessEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Verified Access Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVerifiedAccessEndpointRead(ctx, d, meta)...)
}

func resourceVerifiedAccessEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ep, err := findVerifiedAccessEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(attrVerifiedAccessEndpoint_ApplicationDomain, ep.ApplicationDomain)
	d.Set(attrVerifiedAccessEndpoint_AttachmentType, ep.AttachmentType)
	if err := d.Set("cidr_options", flattenVerifiedAccessEndpointCidrOptions(ep.CidrOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cidr_options: %s", err)
	}

	d.Set(names.AttrDescription, ep.Description)
	d.Set(attrVerifiedAccessEndpoint_DeviceValidationDomain, ep.DeviceValidationDomain)
	d.Set(attrVerifiedAccessEndpoint_DomainCertificateArn, ep.DomainCertificateArn)
	d.Set(attrVerifiedAccessEndpoint_EndpointDomain, ep.EndpointDomain)
	d.Set(names.AttrEndpointType, ep.EndpointType)
	if err := d.Set(attrVerifiedAccessEndpoint_LoadBalancerOptions, flattenVerifiedAccessEndpointLoadBalancerOptions(ep.LoadBalancerOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer_options: %s", err)
	}
	if err := d.Set(attrVerifiedAccessEndpoint_NetworkInterfaceOptions, flattenVerifiedAccessEndpointEniOptions(ep.NetworkInterfaceOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interface_options: %s", err)
	}
	if err := d.Set("rds_options", flattenVerifiedAccessEndpointRdsOptions(ep.RdsOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rds_options: %s", err)
	}
	d.Set(names.AttrSecurityGroupIDs, aws.StringSlice(ep.SecurityGroupIds))
	if err := d.Set(attrVerifiedAccessEndpoint_SseSpecification, flattenVerifiedAccessSseSpecificationRequest(ep.SseSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sse_specification: %s", err)
	}
	d.Set("verified_access_group_id", ep.VerifiedAccessGroupId)
	d.Set("verified_access_instance_id", ep.VerifiedAccessInstanceId)

	output, err := findVerifiedAccessEndpointPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Endpoint (%s) policy: %s", d.Id(), err)
	}

	d.Set(attrVerifiedAccessEndpoint_PolicyDocument, output.PolicyDocument)

	return diags
}

func resourceVerifiedAccessEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(attrVerifiedAccessEndpoint_PolicyDocument, names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyVerifiedAccessEndpointInput{
			ClientToken:              aws.String(id.UniqueId()),
			VerifiedAccessEndpointId: aws.String(d.Id()),
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges("cidr_options") {
			if v, ok := d.GetOk("cidr_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.CidrOptions = expandModifyVerifiedAccessEndpointCidrOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges(attrVerifiedAccessEndpoint_LoadBalancerOptions) {
			if v, ok := d.GetOk(attrVerifiedAccessEndpoint_LoadBalancerOptions); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.LoadBalancerOptions = expandModifyVerifiedAccessEndpointLoadBalancerOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges(attrVerifiedAccessEndpoint_NetworkInterfaceOptions) {
			if v, ok := d.GetOk(attrVerifiedAccessEndpoint_NetworkInterfaceOptions); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.NetworkInterfaceOptions = expandModifyVerifiedAccessEndpointEniOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("rds_options") {
			if v, ok := d.GetOk("rds_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.RdsOptions = expandModifyVerifiedAccessEndpointRdsOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("verified_access_group_id") {
			input.VerifiedAccessGroupId = aws.String(d.Get("verified_access_group_id").(string))
		}

		_, err := conn.ModifyVerifiedAccessEndpoint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitVerifiedAccessEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Verified Access Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(attrVerifiedAccessEndpoint_PolicyDocument) {
		input := &ec2.ModifyVerifiedAccessEndpointPolicyInput{
			VerifiedAccessEndpointId: aws.String(d.Id()),
		}

		if v := d.Get(attrVerifiedAccessEndpoint_PolicyDocument).(string); v != "" {
			input.PolicyEnabled = aws.Bool(true)
			input.PolicyDocument = aws.String(v)
		} else {
			input.PolicyEnabled = aws.Bool(false)
		}

		_, err := conn.ModifyVerifiedAccessEndpointPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Endpoint (%s) policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVerifiedAccessEndpointRead(ctx, d, meta)...)
}

func resourceVerifiedAccessEndpointDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Verified Access Endpoint: %s", d.Id())
	input := ec2.DeleteVerifiedAccessEndpointInput{
		ClientToken:              aws.String(id.UniqueId()),
		VerifiedAccessEndpointId: aws.String(d.Id()),
	}
	_, err := conn.DeleteVerifiedAccessEndpoint(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessEndpointIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitVerifiedAccessEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Verified Access Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func flattenVerifiedAccessEndpointPortRanges(apiObjects []types.VerifiedAccessEndpointPortRange) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, len(apiObjects))
	for i, apiObject := range apiObjects {
		tfmap := map[string]any{}

		if v := apiObject.FromPort; v != nil {
			tfmap[attrVerifiedAccessEndpoint_PortRange_FromPort] = aws.ToInt32(v)
		}

		if v := apiObject.ToPort; v != nil {
			tfmap[attrVerifiedAccessEndpoint_PortRange_ToPort] = aws.ToInt32(v)
		}

		tfList[i] = tfmap
	}

	return tfList
}

func flattenVerifiedAccessEndpointCidrOptions(apiObject *types.VerifiedAccessEndpointCidrOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]any{}

	if v := apiObject.Cidr; v != nil {
		tfmap[attrVerifiedAccessEndpoint_CidrOptions_Cidr] = aws.ToString(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap[names.AttrProtocol] = v
	}

	if v := apiObject.PortRanges; v != nil {
		tfmap[attrVerifiedAccessEndpoint_PortRange] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfmap[names.AttrSubnetIDs] = aws.StringSlice(v)
	}

	return []any{tfmap}
}

func flattenVerifiedAccessEndpointLoadBalancerOptions(apiObject *types.VerifiedAccessEndpointLoadBalancerOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]any{}

	if v := apiObject.LoadBalancerArn; v != nil {
		tfmap[attrVerifiedAccessEndpoint_LoadBalancerOptions_LoadBalancerArn] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfmap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.PortRanges; v != nil {
		tfmap["port_ranges"] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap[names.AttrProtocol] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfmap[names.AttrSubnetIDs] = aws.StringSlice(v)
	}

	return []any{tfmap}
}

func flattenVerifiedAccessEndpointEniOptions(apiObject *types.VerifiedAccessEndpointEniOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]any{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfmap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfmap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.PortRanges; v != nil {
		tfmap["port_ranges"] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap[names.AttrProtocol] = v
	}

	return []any{tfmap}
}

func flattenVerifiedAccessEndpointRdsOptions(apiObject *types.VerifiedAccessEndpointRdsOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]any{}

	if v := apiObject.Port; v != nil {
		tfmap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap[names.AttrProtocol] = v
	}

	if v := apiObject.RdsEndpoint; v != nil {
		tfmap[names.AttrEndpoint] = v
	}

	if v := apiObject.RdsDbClusterArn; v != nil {
		tfmap[attrVerifiedAccessEndpoint_RdsOptions_ClusterArn] = v
	}

	if v := apiObject.RdsDbInstanceArn; v != nil {
		tfmap[attrVerifiedAccessEndpoint_RdsOptions_InstanceArn] = v
	}

	if v := apiObject.RdsDbProxyArn; v != nil {
		tfmap[attrVerifiedAccessEndpoint_RdsOptions_ProxyArn] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfmap[names.AttrSubnetIDs] = aws.StringSlice(v)
	}

	return []any{tfmap}
}

func flattenVerifiedAccessSseSpecificationRequest(apiObject *types.VerifiedAccessSseSpecificationResponse) []any {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]any{}

	if v := apiObject.CustomerManagedKeyEnabled; v != nil {
		tfmap[attrVerifiedAccessEndpoint_SseSpecification_CustomManagedKeyEnabled] = aws.ToBool(v)
	}

	if v := apiObject.KmsKeyArn; v != nil {
		tfmap[names.AttrKMSKeyARN] = aws.ToString(v)
	}

	return []any{tfmap}
}

func expandCreateVerifiedAccessEndpointCidrOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointCidrOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointCidrOptions{}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_CidrOptions_Cidr].(string); ok && v != "" {
		apiobject.Cidr = aws.String(v)
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok && v.Len() > 0 {
		apiobject.PortRanges = expandCreateVerifiedAccessPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiobject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiobject
}

func expandCreateVerifiedAccessEndpointRdsOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointRdsOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointRdsOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiobject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_RdsOptions_ClusterArn].(string); ok && v != "" {
		apiobject.RdsDbClusterArn = aws.String(v)
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_RdsOptions_InstanceArn].(string); ok && v != "" {
		apiobject.RdsDbInstanceArn = aws.String(v)
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_RdsOptions_ProxyArn].(string); ok && v != "" {
		apiobject.RdsDbProxyArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEndpoint].(string); ok && v != "" {
		apiobject.RdsEndpoint = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiobject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiobject
}

func expandVerifiedAccessPortRanges(tfList []any) []types.VerifiedAccessEndpointPortRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	portRanges := make([]types.VerifiedAccessEndpointPortRange, len(tfList))
	for i, tfElem := range tfList {
		tfMap := tfElem.(map[string]any)
		portRanges[i] = types.VerifiedAccessEndpointPortRange{
			FromPort: aws.Int32(int32(tfMap[attrVerifiedAccessEndpoint_PortRange_FromPort].(int))),
			ToPort:   aws.Int32(int32(tfMap[attrVerifiedAccessEndpoint_PortRange_ToPort].(int))),
		}
	}
	return portRanges
}

func expandCreateVerifiedAccessPortRanges(tfList []any) []types.CreateVerifiedAccessEndpointPortRange {
	portRanges := expandVerifiedAccessPortRanges(tfList)
	if portRanges == nil {
		return nil
	}

	newPortRanges := make([]types.CreateVerifiedAccessEndpointPortRange, len(portRanges))
	for i, portRange := range portRanges {
		newPortRanges[i] = types.CreateVerifiedAccessEndpointPortRange{
			FromPort: portRange.FromPort,
			ToPort:   portRange.ToPort,
		}
	}
	return newPortRanges
}

func expandModifyVerifiedAccessPortRanges(tfList []any) []types.ModifyVerifiedAccessEndpointPortRange {
	portRanges := expandVerifiedAccessPortRanges(tfList)
	if portRanges == nil {
		return nil
	}

	newPortRanges := make([]types.ModifyVerifiedAccessEndpointPortRange, len(portRanges))
	for i, portRange := range portRanges {
		newPortRanges[i] = types.ModifyVerifiedAccessEndpointPortRange{
			FromPort: portRange.FromPort,
			ToPort:   portRange.ToPort,
		}
	}
	return newPortRanges
}

func expandCreateVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_LoadBalancerOptions_LoadBalancerArn].(string); ok && v != "" {
		apiobject.LoadBalancerArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiobject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok && v.Len() > 0 {
		apiobject.PortRanges = expandCreateVerifiedAccessPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiobject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiobject
}

func expandCreateVerifiedAccessEndpointEniOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap[names.AttrNetworkInterfaceID].(string); ok && v != "" {
		apiobject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiobject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok && v.Len() > 0 {
		apiobject.PortRanges = expandCreateVerifiedAccessPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}
	return apiobject
}

func expandModifyVerifiedAccessEndpointCidrOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointCidrOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointCidrOptions{}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessPortRanges(v.List())
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointRdsOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointRdsOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointRdsOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrEndpoint]; ok {
		apiObject.RdsEndpoint = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointEniOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_PortRange].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}
	return apiObject
}

func expandCreateVerifiedAccessGenericSseSpecification(tfMap map[string]any) *types.VerifiedAccessSseSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VerifiedAccessSseSpecificationRequest{}

	if v, ok := tfMap[attrVerifiedAccessEndpoint_SseSpecification_CustomManagedKeyEnabled].(bool); ok {
		apiObject.CustomerManagedKeyEnabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
		apiObject.KmsKeyArn = aws.String(v)
	}
	return apiObject
}
