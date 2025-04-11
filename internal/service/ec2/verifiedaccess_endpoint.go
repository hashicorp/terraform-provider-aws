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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
			"application_domain": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"attachment_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.VerifiedAccessEndpointAttachmentType](),
			},
			"cidr_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"port_range": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"from_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
									"to_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
								},
							},
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(types.VerifiedAccessEndpointProtocolTcp), false),
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
			"device_validation_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"endpoint_domain_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"endpoint_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpointType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.VerifiedAccessEndpointType](),
			},
			"load_balancer_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"load_balancer_arn": {
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
						"port_range": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"from_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
									"to_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
								},
							},
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.VerifiedAccessEndpointProtocol](),
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"network_interface_options": {
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
						"port_range": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"from_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
									"to_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumberOrZero,
									},
								},
							},
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.VerifiedAccessEndpointProtocol](),
						},
					},
				},
			},
			"policy_document": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rds_options": {
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
							ValidateFunc: validation.StringInSlice(enum.Slice(types.VerifiedAccessEndpointProtocolTcp), false),
						},
						"rds_db_cluster_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"rds_db_instance_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"rds_db_proxy_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"rds_endpoint": {
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
			"sse_specification": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customer_managed_key_enabled": {
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

	input := ec2.CreateVerifiedAccessEndpointInput{
		AttachmentType:        types.VerifiedAccessEndpointAttachmentType(d.Get("attachment_type").(string)),
		ClientToken:           aws.String(sdkid.UniqueId()),
		EndpointType:          types.VerifiedAccessEndpointType(d.Get(names.AttrEndpointType).(string)),
		TagSpecifications:     getTagSpecificationsIn(ctx, types.ResourceTypeVerifiedAccessEndpoint),
		VerifiedAccessGroupId: aws.String(d.Get("verified_access_group_id").(string)),
	}

	if v, ok := d.GetOk("application_domain"); ok {
		input.ApplicationDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.CidrOptions = expandCreateVerifiedAccessEndpointCIDROptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_certificate_arn"); ok {
		input.DomainCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("endpoint_domain_prefix"); ok {
		input.EndpointDomainPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("load_balancer_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.LoadBalancerOptions = expandCreateVerifiedAccessEndpointLoadBalancerOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("network_interface_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.NetworkInterfaceOptions = expandCreateVerifiedAccessEndpointENIOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("policy_document"); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rds_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RdsOptions = expandCreateVerifiedAccessEndpointRDSOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sse_specification"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SseSpecification = expandVerifiedAccessSSESpecificationRequest(v.([]any)[0].(map[string]any))
	}

	output, err := conn.CreateVerifiedAccessEndpoint(ctx, &input)

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

	d.Set("application_domain", ep.ApplicationDomain)
	d.Set("attachment_type", ep.AttachmentType)
	if err := d.Set("cidr_options", flattenVerifiedAccessEndpointCIDROptions(ep.CidrOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cidr_options: %s", err)
	}
	d.Set(names.AttrDescription, ep.Description)
	d.Set("device_validation_domain", ep.DeviceValidationDomain)
	d.Set("domain_certificate_arn", ep.DomainCertificateArn)
	d.Set("endpoint_domain", ep.EndpointDomain)
	d.Set(names.AttrEndpointType, ep.EndpointType)
	if err := d.Set("load_balancer_options", flattenVerifiedAccessEndpointLoadBalancerOptions(ep.LoadBalancerOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer_options: %s", err)
	}
	if err := d.Set("network_interface_options", flattenVerifiedAccessEndpointENIOptions(ep.NetworkInterfaceOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interface_options: %s", err)
	}
	if err := d.Set("rds_options", flattenVerifiedAccessEndpointRDSOptions(ep.RdsOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rds_options: %s", err)
	}
	d.Set(names.AttrSecurityGroupIDs, ep.SecurityGroupIds)
	if err := d.Set("sse_specification", flattenVerifiedAccessSSESpecificationResponse(ep.SseSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sse_specification: %s", err)
	}
	d.Set("verified_access_group_id", ep.VerifiedAccessGroupId)
	d.Set("verified_access_instance_id", ep.VerifiedAccessInstanceId)

	output, err := findVerifiedAccessEndpointPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Endpoint (%s) policy: %s", d.Id(), err)
	}

	d.Set("policy_document", output.PolicyDocument)

	return diags
}

func resourceVerifiedAccessEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept("policy_document", names.AttrTags, names.AttrTagsAll) {
		input := ec2.ModifyVerifiedAccessEndpointInput{
			ClientToken:              aws.String(sdkid.UniqueId()),
			VerifiedAccessEndpointId: aws.String(d.Id()),
		}

		if d.HasChanges("cidr_options") {
			if v, ok := d.GetOk("cidr_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.CidrOptions = expandModifyVerifiedAccessEndpointCIDROptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges("load_balancer_options") {
			if v, ok := d.GetOk("load_balancer_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.LoadBalancerOptions = expandModifyVerifiedAccessEndpointLoadBalancerOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("network_interface_options") {
			if v, ok := d.GetOk("network_interface_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.NetworkInterfaceOptions = expandModifyVerifiedAccessEndpointENIOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("rds_options") {
			if v, ok := d.GetOk("rds_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.RdsOptions = expandModifyVerifiedAccessEndpointRDSOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("verified_access_group_id") {
			input.VerifiedAccessGroupId = aws.String(d.Get("verified_access_group_id").(string))
		}

		_, err := conn.ModifyVerifiedAccessEndpoint(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitVerifiedAccessEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Verified Access Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("policy_document") {
		input := ec2.ModifyVerifiedAccessEndpointPolicyInput{
			VerifiedAccessEndpointId: aws.String(d.Id()),
		}

		if v := d.Get("policy_document").(string); v != "" {
			input.PolicyEnabled = aws.Bool(true)
			input.PolicyDocument = aws.String(v)
		} else {
			input.PolicyEnabled = aws.Bool(false)
		}

		_, err := conn.ModifyVerifiedAccessEndpointPolicy(ctx, &input)

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
		ClientToken:              aws.String(sdkid.UniqueId()),
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
		tfMap := map[string]any{}

		if v := apiObject.FromPort; v != nil {
			tfMap["from_port"] = aws.ToInt32(v)
		}

		if v := apiObject.ToPort; v != nil {
			tfMap["to_port"] = aws.ToInt32(v)
		}

		tfList[i] = tfMap
	}

	return tfList
}

func flattenVerifiedAccessEndpointCIDROptions(apiObject *types.VerifiedAccessEndpointCidrOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Cidr; v != nil {
		tfMap["cidr"] = aws.ToString(v)
	}

	if v := apiObject.PortRanges; v != nil {
		tfMap["port_range"] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfMap[names.AttrProtocol] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return []any{tfMap}
}

func flattenVerifiedAccessEndpointLoadBalancerOptions(apiObject *types.VerifiedAccessEndpointLoadBalancerOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.LoadBalancerArn; v != nil {
		tfMap["load_balancer_arn"] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.PortRanges; v != nil {
		tfMap["port_ranges"] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfMap[names.AttrProtocol] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return []any{tfMap}
}

func flattenVerifiedAccessEndpointENIOptions(apiObject *types.VerifiedAccessEndpointEniOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.PortRanges; v != nil {
		tfMap["port_ranges"] = flattenVerifiedAccessEndpointPortRanges(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfMap[names.AttrProtocol] = v
	}

	return []any{tfMap}
}

func flattenVerifiedAccessEndpointRDSOptions(apiObject *types.VerifiedAccessEndpointRdsOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfMap[names.AttrProtocol] = v
	}

	if v := apiObject.RdsDbClusterArn; v != nil {
		tfMap["rds_db_cluster_arn"] = v
	}

	if v := apiObject.RdsDbInstanceArn; v != nil {
		tfMap["rds_db_instance_arn"] = v
	}

	if v := apiObject.RdsDbProxyArn; v != nil {
		tfMap["rds_db_proxy_arn"] = v
	}

	if v := apiObject.RdsEndpoint; v != nil {
		tfMap["rds_endpoint"] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return []any{tfMap}
}

func flattenVerifiedAccessSSESpecificationResponse(apiObject *types.VerifiedAccessSseSpecificationResponse) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CustomerManagedKeyEnabled; v != nil {
		tfMap["customer_managed_key_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.KmsKeyArn; v != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(v)
	}

	return []any{tfMap}
}

func expandCreateVerifiedAccessEndpointCIDROptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointCidrOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessEndpointCidrOptions{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	if v, ok := tfMap["port_range"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.PortRanges = expandCreateVerifiedAccessEndpointPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandCreateVerifiedAccessEndpointRDSOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointRdsOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessEndpointRdsOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap["rds_db_cluster_arn"].(string); ok && v != "" {
		apiObject.RdsDbClusterArn = aws.String(v)
	}

	if v, ok := tfMap["rds_db_instance_arn"].(string); ok && v != "" {
		apiObject.RdsDbInstanceArn = aws.String(v)
	}

	if v, ok := tfMap["rds_db_proxy_arn"].(string); ok && v != "" {
		apiObject.RdsDbProxyArn = aws.String(v)
	}

	if v, ok := tfMap["rds_endpoint"].(string); ok && v != "" {
		apiObject.RdsEndpoint = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandVerifiedAccessEndpointPortRanges(tfList []any) []types.VerifiedAccessEndpointPortRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]types.VerifiedAccessEndpointPortRange, len(tfList))

	for i, tfElem := range tfList {
		tfMap := tfElem.(map[string]any)
		apiObjects[i] = types.VerifiedAccessEndpointPortRange{
			FromPort: aws.Int32(int32(tfMap["from_port"].(int))),
			ToPort:   aws.Int32(int32(tfMap["to_port"].(int))),
		}
	}

	return apiObjects
}

func expandCreateVerifiedAccessEndpointPortRanges(tfList []any) []types.CreateVerifiedAccessEndpointPortRange {
	apiObjects := expandVerifiedAccessEndpointPortRanges(tfList)

	if apiObjects == nil {
		return nil
	}

	return tfslices.ApplyToAll(apiObjects, func(v types.VerifiedAccessEndpointPortRange) types.CreateVerifiedAccessEndpointPortRange {
		return types.CreateVerifiedAccessEndpointPortRange{
			FromPort: v.FromPort,
			ToPort:   v.ToPort,
		}
	})
}

func expandModifyVerifiedAccessEndpointPortRanges(tfList []any) []types.ModifyVerifiedAccessEndpointPortRange {
	apiObjects := expandVerifiedAccessEndpointPortRanges(tfList)

	if apiObjects == nil {
		return nil
	}

	return tfslices.ApplyToAll(apiObjects, func(v types.VerifiedAccessEndpointPortRange) types.ModifyVerifiedAccessEndpointPortRange {
		return types.ModifyVerifiedAccessEndpointPortRange{
			FromPort: v.FromPort,
			ToPort:   v.ToPort,
		}
	})
}

func expandCreateVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap["load_balancer_arn"].(string); ok && v != "" {
		apiObject.LoadBalancerArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["port_range"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.PortRanges = expandCreateVerifiedAccessEndpointPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandCreateVerifiedAccessEndpointENIOptions(tfMap map[string]any) *types.CreateVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap[names.AttrNetworkInterfaceID].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["port_range"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.PortRanges = expandCreateVerifiedAccessEndpointPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}
	return apiObject
}

func expandModifyVerifiedAccessEndpointCIDROptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointCidrOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointCidrOptions{}

	if v, ok := tfMap["port_range"].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessEndpointPortRanges(v.List())
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointRDSOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointRdsOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointRdsOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["rds_endpoint"]; ok {
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

	if v, ok := tfMap["port_range"].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessEndpointPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointENIOptions(tfMap map[string]any) *types.ModifyVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap[names.AttrPort].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["port_range"].(*schema.Set); ok {
		apiObject.PortRanges = expandModifyVerifiedAccessEndpointPortRanges(v.List())
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	return apiObject
}

func expandVerifiedAccessSSESpecificationRequest(tfMap map[string]any) *types.VerifiedAccessSseSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VerifiedAccessSseSpecificationRequest{}

	if v, ok := tfMap["customer_managed_key_enabled"].(bool); ok {
		apiObject.CustomerManagedKeyEnabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok && v != "" {
		apiObject.KmsKeyArn = aws.String(v)
	}

	return apiObject
}
