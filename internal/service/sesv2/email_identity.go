// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_email_identity", name="Email Identity")
// @Tags(identifierAttribute="arn")
func ResourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityCreate,
		ReadWithoutTimeout:   resourceEmailIdentityRead,
		UpdateWithoutTimeout: resourceEmailIdentityUpdate,
		DeleteWithoutTimeout: resourceEmailIdentityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"dkim_signing_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_signing_private_key": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"dkim_signing_attributes.0.domain_signing_selector"},
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 20480),
								verify.ValidBase64String,
							),
						},
						"domain_signing_selector": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"dkim_signing_attributes.0.domain_signing_private_key"},
							ValidateFunc: validation.StringLenBetween(1, 63),
						},
						"last_key_generation_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_signing_key_length": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ConflictsWith:    []string{"dkim_signing_attributes.0.domain_signing_private_key", "dkim_signing_attributes.0.domain_signing_selector"},
							ValidateDiagFunc: enum.Validate[types.DkimSigningKeyLength](),
						},
						"signing_attributes_origin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tokens": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verified_for_sending_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameEmailIdentity = "Email Identity"
)

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(d.Get("email_identity").(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("configuration_set_name"); ok {
		in.ConfigurationSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dkim_signing_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.DkimSigningAttributes = expandDKIMSigningAttributes(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateEmailIdentity(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameEmailIdentity, d.Get("email_identity").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameEmailIdentity, d.Get("email_identity").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("email_identity").(string))

	return append(diags, resourceEmailIdentityRead(ctx, d, meta)...)
}

func emailIdentityNameToARN(meta interface{}, emailIdentityName string) string {
	return arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", emailIdentityName),
	}.String()
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindEmailIdentityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameEmailIdentity, d.Id(), err)
	}

	arn := emailIdentityNameToARN(meta, d.Id())

	d.Set(names.AttrARN, arn)
	d.Set("configuration_set_name", out.ConfigurationSetName)
	d.Set("email_identity", d.Id())

	if out.DkimAttributes != nil {
		tfMap := flattenDKIMAttributes(out.DkimAttributes)
		tfMap["domain_signing_private_key"] = d.Get("dkim_signing_attributes.0.domain_signing_private_key").(string)
		tfMap["domain_signing_selector"] = d.Get("dkim_signing_attributes.0.domain_signing_selector").(string)

		if err := d.Set("dkim_signing_attributes", []interface{}{tfMap}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, d.Id(), err)
		}
	} else {
		d.Set("dkim_signing_attributes", nil)
	}

	d.Set("identity_type", string(out.IdentityType))
	d.Set("verified_for_sending_status", out.VerifiedForSendingStatus)

	return diags
}

func resourceEmailIdentityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	if d.HasChanges("configuration_set_name") {
		in := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{
			EmailIdentity: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("configuration_set_name"); ok {
			in.ConfigurationSetName = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating SESV2 EmailIdentity ConfigurationSetAttributes (%s): %#v", d.Id(), in)
		_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameEmailIdentity, d.Id(), err)
		}
	}

	if d.HasChanges("dkim_signing_attributes") {
		in := &sesv2.PutEmailIdentityDkimSigningAttributesInput{
			EmailIdentity:           aws.String(d.Id()),
			SigningAttributesOrigin: types.DkimSigningAttributesOriginAwsSes,
		}

		if v, ok := d.GetOk("dkim_signing_attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.SigningAttributes = expandDKIMSigningAttributes(v.([]interface{})[0].(map[string]interface{}))
			in.SigningAttributesOrigin = getSigningAttributesOrigin(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating SESV2 EmailIdentity DkimSigningAttributes (%s): %#v", d.Id(), in)
		_, err := conn.PutEmailIdentityDkimSigningAttributes(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameEmailIdentity, d.Id(), err)
		}
	}

	return append(diags, resourceEmailIdentityRead(ctx, d, meta)...)
}

func resourceEmailIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 EmailIdentity %s", d.Id())

	_, err := conn.DeleteEmailIdentity(ctx, &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameEmailIdentity, d.Id(), err)
	}

	return diags
}

func FindEmailIdentityByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetEmailIdentityOutput, error) {
	in := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(id),
	}
	out, err := conn.GetEmailIdentity(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandDKIMSigningAttributes(tfMap map[string]interface{}) *types.DkimSigningAttributes {
	if tfMap == nil {
		return nil
	}

	a := &types.DkimSigningAttributes{}

	if v, ok := tfMap["domain_signing_private_key"].(string); ok && v != "" {
		a.DomainSigningPrivateKey = aws.String(v)
	}

	if v, ok := tfMap["domain_signing_selector"].(string); ok && v != "" {
		a.DomainSigningSelector = aws.String(v)
	}

	if v, ok := tfMap["next_signing_key_length"].(string); ok && v != "" {
		a.NextSigningKeyLength = types.DkimSigningKeyLength(v)
	}

	return a
}

func getSigningAttributesOrigin(tfMap map[string]interface{}) types.DkimSigningAttributesOrigin {
	if tfMap == nil {
		return types.DkimSigningAttributesOriginAwsSes
	}

	if v, ok := tfMap["next_signing_key_length"].(string); ok && v != "" {
		return types.DkimSigningAttributesOriginAwsSes
	}

	if v, ok := tfMap["domain_signing_private_key"].(string); ok && v != "" {
		return types.DkimSigningAttributesOriginExternal
	}

	if v, ok := tfMap["domain_signing_selector"].(string); ok && v != "" {
		return types.DkimSigningAttributesOriginExternal
	}

	return types.DkimSigningAttributesOriginAwsSes
}

func flattenDKIMAttributes(apiObject *types.DkimAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"current_signing_key_length": string(apiObject.CurrentSigningKeyLength),
		"next_signing_key_length":    string(apiObject.NextSigningKeyLength),
		"signing_attributes_origin":  string(apiObject.SigningAttributesOrigin),
		names.AttrStatus:             string(apiObject.Status),
	}

	if v := apiObject.LastKeyGenerationTimestamp; v != nil {
		m["last_key_generation_timestamp"] = v.Format(time.RFC3339)
	}

	if v := apiObject.Tokens; v != nil {
		m["tokens"] = apiObject.Tokens
	}

	return m
}
