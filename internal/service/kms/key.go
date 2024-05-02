// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_key", name="Key")
// @Tags(identifierAttribute="id")
func resourceKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyCreate,
		ReadWithoutTimeout:   resourceKeyRead,
		UpdateWithoutTimeout: resourceKeyUpdate,
		DeleteWithoutTimeout: resourceKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"custom_key_store_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 22),
			},
			"customer_master_key_spec": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.CustomerMasterKeySpecSymmetricDefault,
				ValidateDiagFunc: enum.Validate[awstypes.CustomerMasterKeySpec](),
			},
			"deletion_window_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 8192),
			},
			"enable_key_rotation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_usage": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.KeyUsageTypeEncryptDecrypt,
				ValidateDiagFunc: enum.Validate[awstypes.KeyUsageType](),
			},
			"multi_region": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"xks_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"custom_key_store_id"},
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	input := &kms.CreateKeyInput{
		BypassPolicyLockoutSafetyCheck: d.Get("bypass_policy_lockout_safety_check").(bool),
		CustomerMasterKeySpec:          awstypes.CustomerMasterKeySpec(d.Get("customer_master_key_spec").(string)),
		KeyUsage:                       awstypes.KeyUsageType(d.Get("key_usage").(string)),
		Tags:                           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_region"); ok {
		input.MultiRegion = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("policy"); ok {
		p, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Policy = aws.String(p)
	}

	if v, ok := d.GetOk("custom_key_store_id"); ok {
		input.Origin = awstypes.OriginTypeAwsCloudhsm
		input.CustomKeyStoreId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_key_id"); ok {
		input.Origin = awstypes.OriginTypeExternalKeyStore
		input.XksKeyId = aws.String(v.(string))
	}

	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	output, err := waitIAMPropagation(ctx, d.Timeout(schema.TimeoutCreate), func() (*kms.CreateKeyOutput, error) {
		return conn.CreateKey(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Key: %s", err)
	}

	d.SetId(aws.ToString(output.KeyMetadata.KeyId))

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if enableKeyRotation := d.Get("enable_key_rotation").(bool); enableKeyRotation {
		if err := updateKeyRotationEnabled(ctx, conn, d.Id(), enableKeyRotation); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if enabled := d.Get("is_enabled").(bool); !enabled {
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating KMS Key (%s): %s", d.Id(), err)
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk("policy"); ok {
		if err := WaitKeyPolicyPropagated(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) policy propagation: %s", d.Id(), err)
		}
	}

	if tags := KeyValueTags(ctx, getTagsIn(ctx)); len(tags) > 0 {
		if err := waitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) tag propagation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	key, err := findKey(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", d.Id(), err)
	}

	if aws.BoolValue(key.metadata.MultiRegion) &&
		aws.StringValue(key.metadata.MultiRegionConfiguration.MultiRegionKeyType) != kms.MultiRegionKeyTypePrimary {
		return sdkdiag.AppendErrorf(diags, "KMS Key (%s) is not a multi-Region primary key", d.Id())
	}

	d.Set("arn", key.metadata.Arn)
	d.Set("custom_key_store_id", key.metadata.CustomKeyStoreId)
	d.Set("customer_master_key_spec", key.metadata.CustomerMasterKeySpec)
	d.Set("description", key.metadata.Description)
	d.Set("enable_key_rotation", key.rotation)
	d.Set("is_enabled", key.metadata.Enabled)
	d.Set("key_id", key.metadata.KeyId)
	d.Set("key_usage", key.metadata.KeyUsage)
	d.Set("multi_region", key.metadata.MultiRegion)

	if key.metadata.XksKeyConfiguration != nil {
		d.Set("xks_key_id", key.metadata.XksKeyConfiguration.Id)
	} else {
		d.Set("xks_key_id", nil)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), key.policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", key.policy, err)
	}

	d.Set("policy", policyToSet)

	setTagsOut(ctx, key.tags)

	return diags
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && enabled {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, enable := d.HasChange("enable_key_rotation"), d.Get("enable_key_rotation").(bool); hasChange {
		if err := updateKeyRotationEnabled(ctx, conn, d.Id(), enable); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if hasChange, description := d.HasChange("description"), d.Get("description").(string); hasChange {
		if err := updateKeyDescription(ctx, conn, d.Id(), description); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, policy, bypass := d.HasChange("policy"), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool); hasChange {
		if err := updateKeyPolicy(ctx, conn, d.Id(), policy, bypass); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && !enabled {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Key: (%s)", d.Id())
	_, err := conn.ScheduleKeyDeletionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return diags
	}

	if tfawserr.ErrMessageContains(err, kms.ErrCodeInvalidStateException, "is pending deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Key (%s): %s", d.Id(), err)
	}

	if _, err := WaitKeyDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

type kmsKey struct {
	metadata *kms.KeyMetadata
	policy   string
	rotation *bool
	tags     []*kms.Tag
}

func findKey(ctx context.Context, conn *kms.KMS, keyID string, isNewResource bool) (*kmsKey, error) {
	// Wait for propagation since KMS is eventually consistent.
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		var err error
		var key kmsKey

		key.metadata, err = FindKeyByID(ctx, conn, keyID)

		if err != nil {
			return nil, fmt.Errorf("reading KMS Key (%s): %w", keyID, err)
		}

		policy, err := FindKeyPolicyByKeyIDAndPolicyName(ctx, conn, keyID, PolicyNameDefault)

		if err != nil {
			return nil, fmt.Errorf("reading KMS Key (%s) policy: %w", keyID, err)
		}

		key.policy, err = structure.NormalizeJsonString(aws.StringValue(policy))

		if err != nil {
			return nil, fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		if aws.StringValue(key.metadata.Origin) == kms.OriginTypeAwsKms {
			key.rotation, err = findKeyRotationEnabledByKeyID(ctx, conn, keyID)

			if err != nil {
				return nil, fmt.Errorf("reading KMS Key (%s) rotation enabled: %w", keyID, err)
			}
		}

		tags, err := listTags(ctx, conn, keyID)

		if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
			return nil, &retry.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, fmt.Errorf("listing tags for KMS Key (%s): %w", keyID, err)
		}

		key.tags = Tags(tags)

		return &key, nil
	}, isNewResource)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*kmsKey), nil
}

func findKeyRotationEnabledByKeyID(ctx context.Context, conn *kms.Client, keyID string) (*bool, error) {
	input := &kms.GetKeyRotationStatusInput{
		KeyId: aws.String(keyID),
	}

	output, err := conn.GetKeyRotationStatus(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return aws.Bool(output.KeyRotationEnabled), nil
}

func updateKeyDescription(ctx context.Context, conn *kms.KMS, keyID string, description string) error {
	input := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyID),
	}

	_, err := conn.UpdateKeyDescriptionWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("updating description: %w", err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyDescriptionPropagated(ctx, conn, keyID, description)
	if err != nil {
		return fmt.Errorf("updating description: waiting for completion: %w", err)
	}

	return nil
}

func updateKeyEnabled(ctx context.Context, conn *kms.Client, keyID string, enabled bool) error {
	var action string

	updateFunc := func() (interface{}, error) {
		var err error

		if enabled {
			action = "enabling"
			_, err = conn.EnableKey(ctx, &kms.EnableKeyInput{
				KeyId: aws.String(keyID),
			})
		} else {
			action = "disabling"
			_, err = conn.DisableKey(ctx, &kms.DisableKeyInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	if _, err := tfresource.RetryWhenIsA[*awstypes.NotFoundException](ctx, PropagationTimeout, updateFunc); err != nil {
		return fmt.Errorf("%s KMS Key (%s): %w", action, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyStatePropagated(ctx, conn, keyID, enabled); err != nil {
		return fmt.Errorf("waiting for KMS Key (%s) enabled = %t: %w", keyID, enabled, err)
	}

	return nil
}

func updateKeyPolicy(ctx context.Context, conn *kms.KMS, keyID string, policy string, bypassPolicyLockoutSafetyCheck bool) error {
	policy, err := structure.NormalizeJsonString(policy)
	if err != nil {
		return fmt.Errorf("policy contains invalid JSON: %w", err)
	}

	updateFunc := func() (interface{}, error) {
		var err error

		input := &kms.PutKeyPolicyInput{
			BypassPolicyLockoutSafetyCheck: aws.Bool(bypassPolicyLockoutSafetyCheck),
			KeyId:                          aws.String(keyID),
			Policy:                         aws.String(policy),
			PolicyName:                     aws.String(PolicyNameDefault),
		}

		_, err = conn.PutKeyPolicyWithContext(ctx, input)

		return nil, err
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, PropagationTimeout, updateFunc, kms.ErrCodeNotFoundException, kms.ErrCodeMalformedPolicyDocumentException)
	if err != nil {
		return fmt.Errorf("updating policy: %w", err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyPolicyPropagated(ctx, conn, keyID, policy)
	if err != nil {
		return fmt.Errorf("updating policy: waiting for completion: %w", err)
	}

	return nil
}

func updateKeyRotationEnabled(ctx context.Context, conn *kms.Client, keyID string, enabled bool) error {
	var action string

	updateFunc := func() (interface{}, error) {
		var err error

		if enabled {
			action = "enabling"
			_, err = conn.EnableKeyRotation(ctx, &kms.EnableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		} else {
			action = "disabling"
			_, err = conn.DisableKeyRotation(ctx, &kms.DisableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	if _, err := tfresource.RetryWhenIsOneOf[*awstypes.NotFoundException, *awstypes.DisabledException](ctx, KeyRotationUpdatedTimeout, updateFunc); err != nil {
		return fmt.Errorf("%s KMS Key (%s) rotation: %w", action, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyRotationEnabledPropagated(ctx, conn, keyID, enabled); err != nil {
		return fmt.Errorf("waiting for KMS Key (%s) rotation: %w", keyID, err)
	}

	return nil
}

func waitKeyStatePropagated(ctx context.Context, conn *kms.Client, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return output.Enabled == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 15,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyStatePropagationTimeout, checkFunc, opts)
}

func waitKeyRotationEnabledPropagated(ctx context.Context, conn *kms.Client, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := findKeyRotationEnabledByKeyID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.ToBool(output) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyRotationUpdatedTimeout, checkFunc, opts)
}
