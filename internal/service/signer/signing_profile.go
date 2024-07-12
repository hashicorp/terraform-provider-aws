// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_signer_signing_profile", name="Signing Profile")
// @Tags(identifierAttribute="arn")
func ResourceSigningProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningProfileCreate,
		ReadWithoutTimeout:   resourceSigningProfileRead,
		UpdateWithoutTimeout: resourceSigningProfileUpdate,
		DeleteWithoutTimeout: resourceSigningProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(PlatformID_Values(), false),
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revocation_effective_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.ValidityType](),
						},
						names.AttrValue: {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"signing_material": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCertificateARN: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSigningProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("terraform_"),
	).Generate()
	input := &signer.PutSigningProfileInput{
		PlatformId:  aws.String(d.Get("platform_id").(string)),
		ProfileName: aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, exists := d.GetOk("signature_validity_period"); exists {
		signatureValidityPeriod := v.([]interface{})[0].(map[string]interface{})
		input.SignatureValidityPeriod = &types.SignatureValidityPeriod{
			Value: int32(signatureValidityPeriod[names.AttrValue].(int)),
			Type:  types.ValidityType(signatureValidityPeriod[names.AttrType].(string)),
		}
	}

	if v, ok := d.Get("signing_material").([]interface{}); ok && len(v) > 0 {
		input.SigningMaterial = expandSigningMaterial(v)
	}

	_, err := conn.PutSigningProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Signer Signing Profile (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceSigningProfileRead(ctx, d, meta)...)
}

func resourceSigningProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	output, err := findSigningProfileByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Signer Signing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer Signing Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrName, output.ProfileName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.ProfileName)))
	d.Set("platform_display_name", output.PlatformDisplayName)
	d.Set("platform_id", output.PlatformId)
	if err := d.Set("revocation_record", flattenSigningProfileRevocationRecord(output.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_record: %s", err)
	}
	if v := output.SignatureValidityPeriod; v != nil {
		if err := d.Set("signature_validity_period", []interface{}{
			map[string]interface{}{
				names.AttrValue: v.Value,
				names.AttrType:  v.Type,
			},
		}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signature_validity_period: %s", err)
		}
	}
	if output.SigningMaterial != nil {
		if err := d.Set("signing_material", flattenSigningMaterial(output.SigningMaterial)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signing_material: %s", err)
		}
	}
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrVersion, output.ProfileVersion)
	d.Set("version_arn", output.ProfileVersionArn)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceSigningProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceSigningProfileRead(ctx, d, meta)...)
}

func resourceSigningProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	log.Printf("[DEBUG] Deleting Signer Signing Profile: %s", d.Id())
	_, err := conn.CancelSigningProfile(ctx, &signer.CancelSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Signer Signing Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func expandSigningMaterial(in []interface{}) *types.SigningMaterial {
	if len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]interface{})
	var out types.SigningMaterial

	if v, ok := m[names.AttrCertificateARN].(string); ok && v != "" {
		out.CertificateArn = aws.String(v)
	}

	return &out
}

func flattenSigningMaterial(apiObject *types.SigningMaterial) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrCertificateARN: aws.ToString(apiObject.CertificateArn),
	}

	return []interface{}{m}
}

func flattenSigningProfileRevocationRecord(apiObject *types.SigningProfileRevocationRecord) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RevocationEffectiveFrom; v != nil {
		tfMap["revocation_effective_from"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func PlatformID_Values() []string {
	return []string{
		"AWSLambda-SHA384-ECDSA",
		"Notation-OCI-SHA384-ECDSA",
		"AWSIoTDeviceManagement-SHA256-ECDSA",
		"AmazonFreeRTOS-TI-CC3220SF",
		"AmazonFreeRTOS-Default"}
}

func findSigningProfileByName(ctx context.Context, conn *signer.Client, name string) (*signer.GetSigningProfileOutput, error) {
	input := &signer.GetSigningProfileInput{
		ProfileName: aws.String(name),
	}

	output, err := conn.GetSigningProfile(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastRequest: input,
			LastError:   err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Status; status == types.SigningProfileStatusCanceled {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}
