package ec2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_endpoint", name="Verified Access Endpoint")
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
				ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointAttachmentType_Values(), false),
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
				ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointType_Values(), false),
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
							ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointProtocol_Values(), false),
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
							ValidateFunc: validation.StringInSlice(ec2.VerifiedAccessEndpointProtocol_Values(), false),
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
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.CreateVerifiedAccessEndpointInput{
		ApplicationDomain:     aws.String(d.Get("application_domain").(string)),
		AttachmentType:        aws.String(d.Get("attachment_type").(string)),
		DomainCertificateArn:  aws.String(d.Get("domain_certificate_arn").(string)),
		EndpointDomainPrefix:  aws.String(d.Get("endpoint_domain_prefix").(string)),
		EndpointType:          aws.String(d.Get("endpoint_type").(string)),
		TagSpecifications:     getTagSpecificationsIn(ctx, ec2.ResourceTypeVerifiedAccessEndpoint),
		VerifiedAccessGroupId: aws.String(d.Get("verified_access_group_id").(string)),
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

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		in.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	out, err := conn.CreateVerifiedAccessEndpointWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessEndpoint, "", err)
	}

	if out == nil || out.VerifiedAccessEndpoint == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessEndpoint, "", errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.VerifiedAccessEndpoint.VerifiedAccessEndpointId))

	if _, err := WaitVerifiedAccessEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionWaitingForCreation, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	return resourceVerifiedAccessEndpointRead(ctx, d, meta)
}

func resourceVerifiedAccessEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	out, err := FindVerifiedAccessEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	d.Set("application_domain", out.ApplicationDomain)
	d.Set("attachment_type", out.AttachmentType)
	d.Set("description", out.Description)
	d.Set("device_validation_domain", out.DeviceValidationDomain)
	d.Set("domain_certificate_arn", out.DomainCertificateArn)
	d.Set("endpoint_domain", out.EndpointDomain)
	d.Set("endpoint_type", out.EndpointType)
	d.Set("security_group_ids", aws.StringValueSlice(out.SecurityGroupIds))
	d.Set("status", out.Status)
	d.Set("verified_access_group_id", out.VerifiedAccessGroupId)
	d.Set("verified_access_instance_id", out.VerifiedAccessInstanceId)

	if err := d.Set("load_balancer_options", flattenVerifiedAccessEndpointLoadBalancerOptions(out.LoadBalancerOptions)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	if err := d.Set("network_interface_options", flattenVerifiedAccessEndpointEniOptions(out.NetworkInterfaceOptions)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceVerifiedAccessEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

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
		return nil
	}

	log.Printf("[DEBUG] Updating EC2 VerifiedAccessEndpoint (%s): %#v", d.Id(), in)
	_, err := conn.ModifyVerifiedAccessEndpointWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	if _, err := WaitVerifiedAccessEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionWaitingForUpdate, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	return resourceVerifiedAccessEndpointRead(ctx, d, meta)
}

func resourceVerifiedAccessEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 VerifiedAccessEndpoint %s", d.Id())

	_, err := conn.DeleteVerifiedAccessEndpointWithContext(ctx, &ec2.DeleteVerifiedAccessEndpointInput{
		VerifiedAccessEndpointId: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	if _, err := WaitVerifiedAccessEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.EC2, create.ErrActionWaitingForDeletion, ResNameVerifiedAccessEndpoint, d.Id(), err)
	}

	return nil
}

func flattenVerifiedAccessEndpointLoadBalancerOptions(apiObject *ec2.VerifiedAccessEndpointLoadBalancerOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.LoadBalancerArn; v != nil {
		m["load_balancer_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Port; v != nil {
		m["port"] = aws.Int64Value(v)
	}

	if v := apiObject.Protocol; v != nil {
		m["protocol"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		m["subnet_ids"] = aws.StringValueSlice(v)
	}

	return []interface{}{m}
}

func flattenVerifiedAccessEndpointEniOptions(apiObject *ec2.VerifiedAccessEndpointEniOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		m["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.Port; v != nil {
		m["port"] = aws.Int64Value(v)
	}

	if v := apiObject.Protocol; v != nil {
		m["protocol"] = aws.StringValue(v)
	}

	return []interface{}{m}
}

func expandCreateVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]interface{}) *ec2.CreateVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.CreateVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap["load_balancer_arn"].(string); ok && v != "" {
		a.LoadBalancerArn = aws.String(v)
	}

	if v, ok := tfMap["port"].(int); ok {
		a.Port = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		a.Protocol = aws.String(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		a.SubnetIds = flex.ExpandStringSet(v)
	}

	return a
}

func expandCreateVerifiedAccessEndpointEniOptions(tfMap map[string]interface{}) *ec2.CreateVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.CreateVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		a.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["port"].(int); ok {
		a.Port = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		a.Protocol = aws.String(v)
	}
	return a
}

func expandModifyVerifiedAccessEndpointLoadBalancerOptions(tfMap map[string]interface{}) *ec2.ModifyVerifiedAccessEndpointLoadBalancerOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.ModifyVerifiedAccessEndpointLoadBalancerOptions{}

	if v, ok := tfMap["port"].(int); ok {
		a.Port = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		a.Protocol = aws.String(v)
	}

	if v, ok := tfMap["subnet_ids"]; ok {
		a.SubnetIds = flex.ExpandStringList(v.([]interface{}))
	}

	return a
}

func expandModifyVerifiedAccessEndpointEniOptions(tfMap map[string]interface{}) *ec2.ModifyVerifiedAccessEndpointEniOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.ModifyVerifiedAccessEndpointEniOptions{}

	if v, ok := tfMap["port"].(int); ok {
		a.Port = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		a.Protocol = aws.String(v)
	}
	return a
}
