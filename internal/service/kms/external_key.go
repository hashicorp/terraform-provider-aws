// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_external_key", name="External Key")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kms/types;awstypes;awstypes.KeyMetadata")
// @Testing(importIgnore="deletion_window_in_days;bypass_policy_lockout_safety_check")
func resourceExternalKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExternalKeyCreate,
		ReadWithoutTimeout:   resourceExternalKeyRead,
		UpdateWithoutTimeout: resourceExternalKeyUpdate,
		DeleteWithoutTimeout: resourceExternalKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deletion_window_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 8192),
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"expiration_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_material_base64": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"key_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_region": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 32768),
					validation.StringIsJSON,
				),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"valid_to": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
		},
	}
}

func resourceExternalKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	input := &kms.CreateKeyInput{
		BypassPolicyLockoutSafetyCheck: d.Get("bypass_policy_lockout_safety_check").(bool),
		KeyUsage:                       awstypes.KeyUsageTypeEncryptDecrypt,
		Origin:                         awstypes.OriginTypeExternal,
		Tags:                           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_region"); ok {
		input.MultiRegion = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		p, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Policy = aws.String(p)
	}

	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// KMS will report this error until it can validate the policy itself.
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	output, err := waitIAMPropagation(ctx, iamPropagationTimeout, func() (*kms.CreateKeyOutput, error) {
		return conn.CreateKey(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS External Key: %s", err)
	}

	d.SetId(aws.ToString(output.KeyMetadata.KeyId))

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if v, ok := d.GetOk("key_material_base64"); ok {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(ctx, conn, "KMS External Key", d.Id(), v.(string), validTo); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitKeyMaterialImported(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) material import: %s", d.Id(), err)
		}

		if err := waitKeyValidToPropagated(ctx, conn, d.Id(), validTo); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) valid_to update: %s", d.Id(), err)
		}

		// The key can only be disabled if key material has been imported, else:
		// "KMSInvalidStateException: arn:aws:kms:us-west-2:123456789012:key/47e3edc1-945f-413b-88b1-e7341c2d89f7 is pending import."
		if enabled := d.Get(names.AttrEnabled).(bool); !enabled {
			if err := updateKeyEnabled(ctx, conn, "KMS External Key", d.Id(), enabled); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk(names.AttrPolicy); ok {
		if err := waitKeyPolicyPropagated(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) policy update: %s", d.Id(), err)
		}
	}

	if tags := KeyValueTags(ctx, getTagsIn(ctx)); len(tags) > 0 {
		if err := waitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) tag update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceExternalKeyRead(ctx, d, meta)...)
}

func resourceExternalKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	key, err := findKeyInfo(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS External Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS External Key (%s): %s", d.Id(), err)
	}

	if keyManager := key.metadata.KeyManager; keyManager != awstypes.KeyManagerTypeCustomer {
		return sdkdiag.AppendErrorf(diags, "KMS External Key (%s) has invalid KeyManager: %s", d.Id(), keyManager)
	}

	if origin := key.metadata.Origin; origin != awstypes.OriginTypeExternal {
		return sdkdiag.AppendErrorf(diags, "KMS External Key (%s) has invalid Origin: %s", d.Id(), origin)
	}

	if aws.ToBool(key.metadata.MultiRegion) && key.metadata.MultiRegionConfiguration.MultiRegionKeyType != awstypes.MultiRegionKeyTypePrimary {
		return sdkdiag.AppendErrorf(diags, "KMS External Key (%s) is not a multi-Region primary key", d.Id())
	}

	d.Set(names.AttrARN, key.metadata.Arn)
	d.Set(names.AttrDescription, key.metadata.Description)
	d.Set(names.AttrEnabled, key.metadata.Enabled)
	d.Set("expiration_model", key.metadata.ExpirationModel)
	d.Set("key_state", key.metadata.KeyState)
	d.Set("key_usage", key.metadata.KeyUsage)
	d.Set("multi_region", key.metadata.MultiRegion)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), key.policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	if key.metadata.ValidTo != nil {
		d.Set("valid_to", aws.ToTime(key.metadata.ValidTo).Format(time.RFC3339))
	} else {
		d.Set("valid_to", nil)
	}

	setTagsOut(ctx, key.tags)

	return diags
}

func resourceExternalKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if hasChange, enabled, state := d.HasChange(names.AttrEnabled), d.Get(names.AttrEnabled).(bool), awstypes.KeyState(d.Get("key_state").(string)); hasChange && enabled && state != awstypes.KeyStatePendingImport {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, "KMS External Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrDescription) {
		if err := updateKeyDescription(ctx, conn, "KMS External Key", d.Id(), d.Get(names.AttrDescription).(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrPolicy) {
		if err := updateKeyPolicy(ctx, conn, "KMS External Key", d.Id(), d.Get(names.AttrPolicy).(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("valid_to") {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(ctx, conn, "KMS External Key", d.Id(), d.Get("key_material_base64").(string), validTo); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitKeyMaterialImported(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) material import: %s", d.Id(), err)
		}

		if err := waitKeyValidToPropagated(ctx, conn, d.Id(), validTo); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) valid_to update: %s", d.Id(), err)
		}
	}

	if hasChange, enabled, state := d.HasChange(names.AttrEnabled), d.Get(names.AttrEnabled).(bool), awstypes.KeyState(d.Get("key_state").(string)); hasChange && !enabled && state != awstypes.KeyStatePendingImport {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, "KMS External Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceExternalKeyRead(ctx, d, meta)...)
}

func resourceExternalKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS External Key: %s", d.Id())
	_, err := conn.ScheduleKeyDeletion(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.KMSInvalidStateException](err, "is pending deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS External Key (%s): %s", d.Id(), err)
	}

	if _, err := waitKeyDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS External Key (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func importExternalKeyMaterial(ctx context.Context, conn *kms.Client, resourceTypeName, keyID, keyMaterialBase64, validTo string) error {
	// Wait for propagation since KMS is eventually consistent.
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.NotFoundException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.GetParametersForImport(ctx, &kms.GetParametersForImportInput{
			KeyId:             aws.String(keyID),
			WrappingAlgorithm: awstypes.AlgorithmSpecRsaesOaepSha256,
			WrappingKeySpec:   awstypes.WrappingKeySpecRsa2048,
		})
	})

	if err != nil {
		return fmt.Errorf("reading %s (%s) parameters for import: %w", resourceTypeName, keyID, err)
	}

	keyMaterial, err := itypes.Base64Decode(keyMaterialBase64)
	if err != nil {
		return err
	}

	output := outputRaw.(*kms.GetParametersForImportOutput)

	publicKey, err := x509.ParsePKIXPublicKey(output.PublicKey)
	if err != nil {
		return fmt.Errorf("parsing %s (%s) public key (PKIX): %w", resourceTypeName, keyID, err)
	}

	encryptedKeyMaterial, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey.(*rsa.PublicKey), keyMaterial, []byte{})
	if err != nil {
		return fmt.Errorf("encrypting %s (%s) key material (RSA-OAEP): %w", resourceTypeName, keyID, err)
	}

	input := &kms.ImportKeyMaterialInput{
		EncryptedKeyMaterial: encryptedKeyMaterial,
		ExpirationModel:      awstypes.ExpirationModelTypeKeyMaterialDoesNotExpire,
		ImportToken:          output.ImportToken,
		KeyId:                aws.String(keyID),
	}

	if validTo != "" {
		t, err := time.Parse(time.RFC3339, validTo)
		if err != nil {
			return err
		}

		input.ExpirationModel = awstypes.ExpirationModelTypeKeyMaterialExpires
		input.ValidTo = aws.Time(t)
	}

	// Wait for propagation since KMS is eventually consistent.
	_, err = tfresource.RetryWhenIsA[*awstypes.NotFoundException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.ImportKeyMaterial(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("importing %s (%s) key material: %w", resourceTypeName, keyID, err)
	}

	return nil
}

func waitKeyMaterialImported(ctx context.Context, conn *kms.Client, id string) (*awstypes.KeyMetadata, error) { //nolint:unparam
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KeyStatePendingImport),
		Target:  enum.Slice(awstypes.KeyStateDisabled, awstypes.KeyStateEnabled),
		Refresh: statusKeyState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func waitKeyValidToPropagated(ctx context.Context, conn *kms.Client, id string, validTo string) error {
	checkFunc := func() (bool, error) {
		output, err := findKeyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if output.ValidTo != nil {
			return aws.ToTime(output.ValidTo).Format(time.RFC3339) == validTo, nil
		}

		return validTo == "", nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}
	const (
		timeout = 5 * time.Minute
	)

	return tfresource.WaitUntil(ctx, timeout, checkFunc, opts)
}
