package ec2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_ec2_verified_access_endpoint", name="Verified Access Endpoint")
// Tagging annotations are used for "transparent tagging".
// Change the "identifierAttribute" value to the name of the attribute used in ListTags and UpdateTags calls (e.g. "arn").
// @Tags(identifierAttribute="id")
func ResourceVerifiedAccessEndpoint() *schema.Resource {
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
				Required: true,
				ForceNew: true,
			},
			"attachment_type": {
				Type:         schema.TypeString,
				Required:     true,
				// ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointAttachmentType_Values(), false),
				ForceNew:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"device_validation_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_certificate_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			"endpoint_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_domain_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Required:     true,
				// ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointType_Values(), false),
				ForceNew:     true,
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
							ValidateFunc: verify.ValidARN,
							ForceNew:     true,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							// ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointProtocol_Values(), false),
						},
						"subnet_ids": {
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
						"network_interface_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							// ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointProtocol_Values(), false),
						},
					},
				},
			},
			"sse_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customer_managed_key_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
							ForceNew:     true,
						},
					},
				},
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verified_access_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verified_access_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}


const (
	ResNameVerifiedAccessEndpoint = "Verified Access Endpoint"
)

func resourceVerifiedAccessEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	
	in := &ec2.CreateVerifiedAccessEndpointInput{
		ApplicationDomain: aws.String(d.Get("name").(string)),
		DomainCertificateArn:  aws.String(d.Get("domain_certificate_arn").(string)),
		EndpointDomainPrefix: aws.String(d.Get("endpoint_domain_prefix").(string)),
		VerifiedAccessGroupId: aws.String(d.Get("verified_access_group_id").(string)),
		AttachmentType: types.VerifiedAccessEndpointAttachmentType(d.Get("attachment_type").(string)),
		EndpointType: types.VerifiedAccessEndpointType(d.Get("endpoint_type").(string)),
		TagSpecifications: getTagSpecificationsInV2(ctx, types.ResourceTypeVerifiedAccessEndpoint),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("load_balancer_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.LoadBalancerOptions = expandCreateVerifiedAccessEndpointLoadBalancerOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("network_interface_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.NetworkInterfaceOptions = expandCreateVerifiedAccessEndpointEniOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("sse_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.SseSpecification = expandCreateVerifiedAccessEndpointSseSpecification(v.([]interface{})[0].(map[string]interface{}))
	}
	
	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		in.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	out, err := conn.CreateVerifiedAccessEndpoint(ctx, in)
	if err != nil {

		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessEndpoint, d.Get("name").(string), err)...)
	}

	if out == nil || out.VerifiedAccessEndpoint == nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessEndpoint, d.Get("name").(string), errors.New("empty output"))...)
	}

	if out == nil || out.VerifiedAccessEndpoint == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessEndpoint, "", errors.New("empty output"))
	}
	
	d.SetId(aws.ToString(out.VerifiedAccessEndpoint.VerifiedAccessEndpointId))
	
	if _, err := waitVerifiedAccessEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionWaitingForCreation, ResNameVerifiedAccessEndpoint, d.Id(), err)...)
	}
	
	return append(diags, resourceVerifiedAccessEndpointRead(ctx, d, meta)...)
}

func resourceVerifiedAccessEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	
	out, err := FindVerifiedAccessEndpointByID(ctx, conn, d.Id())
	
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessEndpoint, d.Id(), err)...)
	}
	
	d.Set("application_domain", out.ApplicationDomain)
	d.Set("attachment_type", out.AttachmentType)
	d.Set("description", out.Description)
	d.Set("device_validation_domain", out.DeviceValidationDomain)
	d.Set("domain_certificate_arn", out.DomainCertificateArn)
	d.Set("endpoint_domain", out.EndpointDomain)
	d.Set("endpoint_type", out.EndpointType)
	d.Set("security_group_ids", aws.StringSlice(out.SecurityGroupIds))
	d.Set("status", out.Status)
	d.Set("verified_access_group_id", out.VerifiedAccessGroupId)
	d.Set("verified_access_instance_id", out.VerifiedAccessInstanceId)

	if err := d.Set("load_balancer_options", flattenVerifiedAccessEndpointLoadBalancerOptions(out.LoadBalancerOptions)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	if err := d.Set("network_interface_options", flattenVerifiedAccessEndpointEniOptions(out.NetworkInterfaceOptions)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	if err := d.Set("sse_specification", flattenVerifiedAccessSseSpecificationRequest(out.SseSpecification)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	return diags
}

func resourceVerifiedAccessEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	update := false

	in := &ec2.ModifyVerifiedAccessEndpointInput{
		VerifiedAccessEndpointId: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChanges("load_balancer_options") {
		if v, ok := d.GetOk("load_balancer_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.LoadBalancerOptions = expandModifyVerifiedAccessEndpointLoadBalancerOptions(v.([]interface{})[0].(map[string]interface{}))
			update = true
		}
	}

	if d.HasChanges("network_interface_options") {
		if v, ok := d.GetOk("network_interface_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.NetworkInterfaceOptions = expandModifyVerifiedAccessEndpointEniOptions(v.([]interface{})[0].(map[string]interface{}))
			update = true
		}
	}

	if d.HasChanges("verified_access_group_id") {
		in.VerifiedAccessGroupId = aws.String(d.Get("description").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		return diags
	}
	
	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating EC2 VerifiedAccessEndpoint (%s): %#v", d.Id(), in)
	_, err := conn.ModifyVerifiedAccessEndpoint(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessEndpoint, d.Id(), err)...)
	}
	
	if _, err := waitVerifiedAccessEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionWaitingForUpdate, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}
	
	return append(diags, resourceVerifiedAccessEndpointRead(ctx, d, meta)...)
}

func resourceVerifiedAccessEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	
	log.Printf("[INFO] Deleting EC2 VerifiedAccessEndpoint %s", d.Id())
	
	_, err := conn.DeleteVerifiedAccessEndpoint(ctx, &ec2.DeleteVerifiedAccessEndpointInput{
		VerifiedAccessEndpointId: aws.String(d.Id()),
	})
		
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessEndpoint, d.Id(), err)...)
	}
	
	if _, err := waitVerifiedAccessEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionWaitingForDeletion, ResNameVerifiedAccessEndpoint, d.Id(), err)...)
	}
	
	return diags
}


const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)


func waitVerifiedAccessEndpointCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{types.VerifiedAccessEndpointStatusCode},
		Refresh:                   statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(types.VerifiedAccessEndpoint); ok {
		return out, err
	}

	return nil, err
}

func waitVerifiedAccessEndpointUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.VerifiedAccessEndpoint); ok {
		return out, err
	}

	return nil, err
}

func waitVerifiedAccessEndpointDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.VerifiedAccessEndpoint); ok {
		return out, err
	}

	return nil, err
}

func statusVerifiedAccessEndpoint(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVerifiedAccessEndpointByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		
		return out, aws.ToString(out.Status.Message), nil
	}
}

func flattenVerifiedAccessEndpointLoadBalancerOptions(apiObject *types.VerifiedAccessEndpointLoadBalancerOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]interface{}{}

	if v := apiObject.LoadBalancerArn; v != nil {
		tfmap["load_balancer_arn"] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfmap["port"] = aws.ToInt32(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap["protocol"] = types.VerifiedAccessEndpointProtocol(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfmap["subnet_ids"] = aws.StringSlice(v)
	}

	return []interface{}{tfmap}
}

func flattenVerifiedAccessEndpointEniOptions(apiObject *types.VerifiedAccessEndpointEniOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]interface{}{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfmap["network_interface_id"] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfmap["port"] = aws.ToInt32(v)
	}

	if v := apiObject.Protocol; v != "" {
		tfmap["protocol"] = types.VerifiedAccessEndpointProtocol(v)
	}

	return []interface{}{tfmap}
}

func flattenVerifiedAccessSseSpecificationRequest(apiObject *types.VerifiedAccessSseSpecificationResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfmap := map[string]interface{}{}

	if v := apiObject.CustomerManagedKeyEnabled; v != nil {
		tfmap["customer_managed_key_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.KmsKeyArn; v != nil {
		tfmap["kms_key_arn"] = aws.ToString(v)
	}

	return []interface{}{tfmap}
}

func expandCreateVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]interface{}) *types.CreateVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap["load_balancer_arn"].(string); ok && v != "" {
		apiobject.LoadBalancerArn = aws.String(v)
	}

	if v, ok := tfMap["port"].(int); ok {
		apiobject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiobject.SubnetIds =  flex.ExpandStringValueSet(v)
	}

	return apiobject
}

func expandCreateVerifiedAccessEndpointEniOptions(tfMap map[string]interface{}) *types.CreateVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiobject := &types.CreateVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiobject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["port"].(int); ok {
		apiobject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		apiobject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}
	return apiobject
}

func expandModifyVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]interface{}) *types.ModifyVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap["port"].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}

	if v, ok := tfMap["subnet_ids"]; ok {
		apiObject.SubnetIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	return apiObject
}

func expandModifyVerifiedAccessEndpointEniOptions(tfMap map[string]interface{}) *types.ModifyVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap["port"].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		apiObject.Protocol = types.VerifiedAccessEndpointProtocol(v)
	}
	return apiObject
}

func expandCreateVerifiedAccessEndpointSseSpecification(tfMap map[string]interface{}) *types.VerifiedAccessSseSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VerifiedAccessSseSpecificationRequest{}

	if v, ok := tfMap["customer_managed_key_enabled"].(bool); ok {
		apiObject.CustomerManagedKeyEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["kms_key_arn"].(string); ok && v != "" {
		apiObject.KmsKeyArn = aws.String(v)
	}
	return apiObject
}