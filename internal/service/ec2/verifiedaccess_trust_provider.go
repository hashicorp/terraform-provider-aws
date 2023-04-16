package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_trust_provider", name="Verified Access Trust Provider")
// @Tags(identifierAttribute="id")
func ResourceVerifiedAccessTrustProvider() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceVerifiedAccessTrustProviderCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessTrustProviderRead,
		UpdateWithoutTimeout: resourceVerifiedAccessTrustProviderUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessTrustProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"device_options": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tenant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"device_trust_provider_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.DeviceTrustProviderType_Values(), false),
			},
			"oidc_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"client_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"client_secret": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"issuer": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"scope": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"token_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"user_info_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
					},
				},
			},
			"policy_reference_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), ""),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"trust_provider_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrustProviderType_Values(), false),
			},
			"user_trust_provider_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.UserTrustProviderType_Values(), false),
			},
		},
		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			CustomizeDiffValidateOIDCOptions,
			CustomizeDiffValidateTrustProviderType,
		),
	}
}

const (
	ResNameVerifiedAccessTrustProvider = "Verified Access Trust Provider"
)

func resourceVerifiedAccessTrustProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.CreateVerifiedAccessTrustProviderInput{
		PolicyReferenceName: aws.String(d.Get("policy_reference_name").(string)),
		TagSpecifications:   getTagSpecificationsIn(ctx, ec2.ResourceTypeVerifiedAccessTrustProvider),
		TrustProviderType:   aws.String(d.Get("trust_provider_type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("device_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.DeviceOptions = expandCreateVerifiedAccessTrustProviderDeviceOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("device_trust_provider_type"); ok {
		in.DeviceTrustProviderType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.OidcOptions = expandCreateVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("user_trust_provider_type"); ok {
		in.UserTrustProviderType = aws.String(v.(string))
	}

	out, err := conn.CreateVerifiedAccessTrustProviderWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProvider, "", err)
	}

	if out == nil || out.VerifiedAccessTrustProvider == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProvider, "", errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.VerifiedAccessTrustProvider.VerifiedAccessTrustProviderId))

	return resourceVerifiedAccessTrustProviderRead(ctx, d, meta)
}

func resourceVerifiedAccessTrustProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).EC2Conn()
	out, err := FindVerifiedAccessTrustProviderByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessTrustProvider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessTrustProvider, d.Id(), err)
	}

	d.Set("description", out.Description)

	if v := out.DeviceOptions; v != nil {
		if err := d.Set("device_options", flattenDeviceOptions(v)); err != nil {
			return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessTrustProvider, d.Id(), err)
		}
	}

	d.Set("device_trust_provider_type", out.DeviceTrustProviderType)

	if v := out.OidcOptions; v != nil {
		if err := d.Set("oidc_options", flattenOIDCOptions(v)); err != nil {
			return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessTrustProvider, d.Id(), err)
		}
	}

	d.Set("policy_reference_name", out.PolicyReferenceName)
	d.Set("trust_provider_type", out.TrustProviderType)
	d.Set("user_trust_provider_type", out.UserTrustProviderType)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceVerifiedAccessTrustProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).EC2Conn()

	update := false

	in := &ec2.ModifyVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String(v.(string))
			update = true
		}
	}

	if d.HasChanges("oidc_options") {
		if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.OidcOptions = expandModifyVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
			update = true
		}
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating EC2 VerifiedAccessTrustProvider (%s): %#v", d.Id(), in)
	_, err := conn.ModifyVerifiedAccessTrustProviderWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessTrustProvider, d.Id(), err)
	}

	return resourceVerifiedAccessTrustProviderRead(ctx, d, meta)
}

func resourceVerifiedAccessTrustProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 VerifiedAccessTrustProvider %s", d.Id())

	_, err := conn.DeleteVerifiedAccessTrustProviderWithContext(ctx, &ec2.DeleteVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessTrustProvider, d.Id(), err)
	}

	return nil
}

func flattenDeviceOptions(apiObject *ec2.DeviceOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.TenantId; v != nil {
		m["tenant_id"] = aws.StringValue(v)
	}

	return []interface{}{m}
}

func flattenOIDCOptions(apiObject *ec2.OidcOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AuthorizationEndpoint; v != nil {
		m["authorization_endpoint"] = aws.StringValue(v)
	}
	if v := apiObject.ClientId; v != nil {
		m["client_id"] = aws.StringValue(v)
	}
	if v := apiObject.ClientSecret; v != nil {
		m["client_secret"] = aws.StringValue(v)
	}
	if v := apiObject.Issuer; v != nil {
		m["issuer"] = aws.StringValue(v)
	}
	if v := apiObject.Scope; v != nil {
		m["scope"] = aws.StringValue(v)
	}
	if v := apiObject.TokenEndpoint; v != nil {
		m["token_endpoint"] = aws.StringValue(v)
	}
	if v := apiObject.UserInfoEndpoint; v != nil {
		m["user_info_endpoint"] = aws.StringValue(v)
	}

	return []interface{}{m}
}

func expandCreateVerifiedAccessTrustProviderDeviceOptions(tfMap map[string]interface{}) *ec2.CreateVerifiedAccessTrustProviderDeviceOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.CreateVerifiedAccessTrustProviderDeviceOptions{}

	if v, ok := tfMap["tenant_id"].(string); ok && v != "" {
		a.TenantId = aws.String(v)
	}

	return a
}

func expandCreateVerifiedAccessTrustProviderOIDCOptions(tfMap map[string]interface{}) *ec2.CreateVerifiedAccessTrustProviderOidcOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.CreateVerifiedAccessTrustProviderOidcOptions{}

	if v, ok := tfMap["authorization_endpoint"].(string); ok && v != "" {
		a.AuthorizationEndpoint = aws.String(v)
	}
	if v, ok := tfMap["client_id"].(string); ok && v != "" {
		a.ClientId = aws.String(v)
	}
	if v, ok := tfMap["client_secret"].(string); ok && v != "" {
		a.ClientSecret = aws.String(v)
	}
	if v, ok := tfMap["issuer"].(string); ok && v != "" {
		a.Issuer = aws.String(v)
	}
	if v, ok := tfMap["scope"].(string); ok && v != "" {
		a.Scope = aws.String(v)
	}
	if v, ok := tfMap["token_endpoint"].(string); ok && v != "" {
		a.TokenEndpoint = aws.String(v)
	}
	if v, ok := tfMap["user_info_endpoint"].(string); ok && v != "" {
		a.UserInfoEndpoint = aws.String(v)
	}

	return a
}

func expandModifyVerifiedAccessTrustProviderOIDCOptions(tfMap map[string]interface{}) *ec2.ModifyVerifiedAccessTrustProviderOidcOptions {
	if tfMap == nil {
		return nil
	}

	a := &ec2.ModifyVerifiedAccessTrustProviderOidcOptions{}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		a.Scope = aws.String(v)
	}

	return a
}

func CustomizeDiffValidateOIDCOptions(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// oidc_options were not specified, ignore logic
	if _, ok := diff.GetOk("oidc_options"); !ok {
		return nil
	}

	if _, ok := diff.GetOk("oidc_options"); ok {
		if v, ok := diff.GetOk("user_trust_provider_type"); ok {
			if v == ec2.UserTrustProviderTypeOidc {
				return nil
			}
		}
	}

	if _, ok := diff.GetOk("oidc_options"); ok {
		if v, ok := diff.GetOk("trust_provider_type"); ok {
			if v == ec2.TrustProviderTypeDevice {
				return fmt.Errorf("argument 'trust_provider_type' must be set to %q when specifying 'oidc_options'", ec2.TrustProviderTypeUser)
			}
		}
	}

	return fmt.Errorf("argument 'user_trust_provider_type' must be set to %q when specifying 'oidc_options'", ec2.UserTrustProviderTypeOidc)
}

func CustomizeDiffValidateTrustProviderType(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("trust_provider_type"); ok {
		if v == "device" {
			if _, ok := diff.GetOk("device_trust_provider_type"); !ok {
				return fmt.Errorf("argument 'device_trust_provider_type' must be set when 'trust_provider_type' is %q", ec2.TrustProviderTypeDevice)
			}
			if _, ok := diff.GetOk("user_trust_provider_type"); ok {
				return fmt.Errorf("argument 'user_trust_provider_type' must NOT be set when 'trust_provider_type' is %q", ec2.TrustProviderTypeDevice)
			}
		}

		if v == "user" {
			if v, ok := diff.GetOk("device_options"); ok && v != "" {
				return fmt.Errorf("argument 'device_options' must NOT be set when 'trust_provider_type' is %q", ec2.TrustProviderTypeUser)
			}
			if v, ok := diff.GetOk("device_trust_provider_type"); ok && v != "" {
				return fmt.Errorf("argument 'device_trust_provider_type' must NOT be set when 'trust_provider_type' is %q", ec2.TrustProviderTypeUser)
			}
			if _, ok := diff.GetOk("user_trust_provider_type"); !ok {
				return fmt.Errorf("argument 'user_trust_provider_type' must be set when 'trust_provider_type' is %q", ec2.TrustProviderTypeUser)
			}
		}
	}
	return nil
}
