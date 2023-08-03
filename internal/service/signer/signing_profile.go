// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"

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
			"platform_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(PlatformID_Values(), false),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.ValidityType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_material": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_arn": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
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

	log.Printf("[DEBUG] Creating Signer signing profile")

	profileName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	profileName = strings.Replace(profileName, "-", "_", -1)

	signingProfileInput := &signer.PutSigningProfileInput{
		ProfileName: aws.String(profileName),
		PlatformId:  aws.String(d.Get("platform_id").(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, exists := d.GetOk("signature_validity_period"); exists {
		signatureValidityPeriod := v.([]interface{})[0].(map[string]interface{})
		signingProfileInput.SignatureValidityPeriod = &types.SignatureValidityPeriod{
			Value: int32(signatureValidityPeriod["value"].(int)),
			Type:  types.ValidityType(signatureValidityPeriod["type"].(string)),
		}
	}

	if v, ok := d.Get("signing_material").([]interface{}); ok && len(v) > 0 {
		signingProfileInput.SigningMaterial = expandSigningMaterial(v)
	}

	_, err := conn.PutSigningProfile(ctx, signingProfileInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Signer signing profile: %s", err)
	}

	d.SetId(profileName)

	return append(diags, resourceSigningProfileRead(ctx, d, meta)...)
}

func resourceSigningProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	signingProfileOutput, err := findSigningProfileByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Signer Signing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer signing profile (%s): %s", d.Id(), err)
	}

	if err := d.Set("platform_id", signingProfileOutput.PlatformId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile platform id: %s", err)
	}
	if signingProfileOutput.SignatureValidityPeriod != nil {
		if err := d.Set("signature_validity_period", []interface{}{
			map[string]interface{}{
				"value": signingProfileOutput.SignatureValidityPeriod.Value,
				"type":  signingProfileOutput.SignatureValidityPeriod.Type,
			},
		}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signer signing profile signature validity period: %s", err)
		}
	}

	if err := d.Set("platform_display_name", signingProfileOutput.PlatformDisplayName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile platform display name: %s", err)
	}

	if err := d.Set("name", signingProfileOutput.ProfileName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile name: %s", err)
	}

	if err := d.Set("arn", signingProfileOutput.Arn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile arn: %s", err)
	}

	if err := d.Set("version", signingProfileOutput.ProfileVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile version: %s", err)
	}

	if err := d.Set("version_arn", signingProfileOutput.ProfileVersionArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile version arn: %s", err)
	}

	if err := d.Set("status", signingProfileOutput.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile status: %s", err)
	}
	if signingProfileOutput.SigningMaterial != nil {
		if err := d.Set("signing_material", flattenSigningMaterial(signingProfileOutput.SigningMaterial)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signer signing profile material: %s", err)
		}
	}

	setTagsOut(ctx, signingProfileOutput.Tags)

	if err := d.Set("revocation_record", flattenSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile revocation record: %s", err)
	}

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

	_, err := conn.CancelSigningProfile(ctx, &signer.CancelSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "canceling Signer signing profile (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Signer signing profile %q canceled", d.Id())
	return diags
}

func expandSigningMaterial(in []interface{}) *types.SigningMaterial {
	if len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]interface{})
	var out types.SigningMaterial

	if v, ok := m["certificate_arn"].(string); ok && v != "" {
		out.CertificateArn = aws.String(v)
	}

	return &out
}

func flattenSigningMaterial(apiObject *types.SigningMaterial) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"certificate_arn": aws.ToString(apiObject.CertificateArn),
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
	in := &signer.GetSigningProfileInput{
		ProfileName: aws.String(name),
	}

	out, err := conn.GetSigningProfile(ctx, in)

	if err != nil {
		return nil, err
	}

	var nfe *types.ResourceNotFoundException
	if errors.As(err, &nfe) {
		return nil, &retry.NotFoundError{
			LastRequest: in,
			LastError:   err,
		}
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
