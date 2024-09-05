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
	awspolicy "github.com/hashicorp/awspolicyequivalence"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kms/types;awstypes;awstypes.KeyMetadata")
// @Testing(importIgnore="deletion_window_in_days;bypass_policy_lockout_safety_check")
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
			Create: schema.DefaultTimeout(iamPropagationTimeout),
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
			names.AttrDescription: {
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
			names.AttrKeyID: {
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
			names.AttrPolicy: {
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
			"rotation_period_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(90, 2560),
				RequiredWith: []string{"enable_key_rotation"},
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

	if enableKeyRotation, rotationPeriod := d.Get("enable_key_rotation").(bool), d.Get("rotation_period_in_days").(int); enableKeyRotation {
		if err := updateKeyRotationEnabled(ctx, conn, "KMS Key", d.Id(), enableKeyRotation, rotationPeriod); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if enabled := d.Get("is_enabled").(bool); !enabled {
		if err := updateKeyEnabled(ctx, conn, "KMS Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk(names.AttrPolicy); ok {
		if err := waitKeyPolicyPropagated(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) policy update: %s", d.Id(), err)
		}
	}

	if tags := KeyValueTags(ctx, getTagsIn(ctx)); len(tags) > 0 {
		if err := waitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) tag update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	key, err := findKeyInfo(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", d.Id(), err)
	}

	if aws.ToBool(key.metadata.MultiRegion) && key.metadata.MultiRegionConfiguration.MultiRegionKeyType != awstypes.MultiRegionKeyTypePrimary {
		return sdkdiag.AppendErrorf(diags, "KMS Key (%s) is not a multi-Region primary key", d.Id())
	}

	d.Set(names.AttrARN, key.metadata.Arn)
	d.Set("custom_key_store_id", key.metadata.CustomKeyStoreId)
	d.Set("customer_master_key_spec", key.metadata.CustomerMasterKeySpec)
	d.Set(names.AttrDescription, key.metadata.Description)
	d.Set("enable_key_rotation", key.rotation)
	d.Set("is_enabled", key.metadata.Enabled)
	d.Set(names.AttrKeyID, key.metadata.KeyId)
	d.Set("key_usage", key.metadata.KeyUsage)
	d.Set("multi_region", key.metadata.MultiRegion)
	d.Set("rotation_period_in_days", key.rotationPeriodInDays)
	if key.metadata.XksKeyConfiguration != nil {
		d.Set("xks_key_id", key.metadata.XksKeyConfiguration.Id)
	} else {
		d.Set("xks_key_id", nil)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), key.policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	setTagsOut(ctx, key.tags)

	return diags
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && enabled {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, "KMS Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if hasChange, hasChangedRotationPeriod,
		enable, rotationPeriod := d.HasChange("enable_key_rotation"), d.HasChange("rotation_period_in_days"),
		d.Get("enable_key_rotation").(bool), d.Get("rotation_period_in_days").(int); hasChange || (enable && hasChangedRotationPeriod) {
		if err := updateKeyRotationEnabled(ctx, conn, "KMS Key", d.Id(), enable, rotationPeriod); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if hasChange, description := d.HasChange(names.AttrDescription), d.Get(names.AttrDescription).(string); hasChange {
		if err := updateKeyDescription(ctx, conn, "KMS Key", d.Id(), description); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if hasChange, policy, bypass := d.HasChange(names.AttrPolicy), d.Get(names.AttrPolicy).(string), d.Get("bypass_policy_lockout_safety_check").(bool); hasChange {
		if err := updateKeyPolicy(ctx, conn, "KMS Key", d.Id(), policy, bypass); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && !enabled {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, "KMS Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Key: %s", d.Id())
	_, err := conn.ScheduleKeyDeletion(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.KMSInvalidStateException](err, "is pending deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Key (%s): %s", d.Id(), err)
	}

	if _, err := waitKeyDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) delete: %s", d.Id(), err)
	}

	return diags
}

type kmsKeyInfo struct {
	metadata             *awstypes.KeyMetadata
	policy               string
	rotation             *bool
	rotationPeriodInDays *int32
	tags                 []awstypes.Tag
}

func findKeyInfo(ctx context.Context, conn *kms.Client, keyID string, isNewResource bool) (*kmsKeyInfo, error) {
	// Wait for propagation since KMS is eventually consistent.
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		var err error
		var key kmsKeyInfo

		key.metadata, err = findKeyByID(ctx, conn, keyID)

		if err != nil {
			return nil, fmt.Errorf("reading KMS Key (%s): %w", keyID, err)
		}

		policy, err := findKeyPolicyByTwoPartKey(ctx, conn, keyID, policyNameDefault)

		if err != nil {
			return nil, fmt.Errorf("reading KMS Key (%s) policy: %w", keyID, err)
		}

		key.policy, err = structure.NormalizeJsonString(aws.ToString(policy))

		if err != nil {
			return nil, fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		if key.metadata.Origin == awstypes.OriginTypeAwsKms {
			key.rotation, key.rotationPeriodInDays, err = findKeyRotationEnabledByKeyID(ctx, conn, keyID)

			if err != nil {
				return nil, fmt.Errorf("reading KMS Key (%s) rotation enabled: %w", keyID, err)
			}
		}

		tags, err := listTags(ctx, conn, keyID)

		if errs.IsA[*awstypes.NotFoundException](err) {
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

	return outputRaw.(*kmsKeyInfo), nil
}

func findKeyByID(ctx context.Context, conn *kms.Client, keyID string, optFns ...func(*kms.Options)) (*awstypes.KeyMetadata, error) {
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	output, err := findKey(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	// Once the CMK is in the pending (replica) deletion state Terraform considers it logically deleted.
	if state := output.KeyState; state == awstypes.KeyStatePendingDeletion || state == awstypes.KeyStatePendingReplicaDeletion {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findKey(ctx context.Context, conn *kms.Client, input *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*awstypes.KeyMetadata, error) {
	output, err := conn.DescribeKey(ctx, input, optFns...)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyMetadata == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.KeyMetadata, nil
}

func findDefaultKeyARNForService(ctx context.Context, conn *kms.Client, service, region string) (string, error) {
	keyID := fmt.Sprintf("alias/aws/%s", service)
	key, err := findKeyByID(ctx, conn, keyID, func(o *kms.Options) {
		o.Region = region
	})

	if err != nil {
		return "", fmt.Errorf("reading KMS Key (%s): %s", keyID, err)
	}

	return aws.ToString(key.Arn), nil
}

func findKeyPolicyByTwoPartKey(ctx context.Context, conn *kms.Client, keyID, policyName string) (*string, error) {
	input := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyID),
		PolicyName: aws.String(policyName),
	}

	output, err := conn.GetKeyPolicy(ctx, input)

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

	return output.Policy, nil
}

func findKeyRotationEnabledByKeyID(ctx context.Context, conn *kms.Client, keyID string) (*bool, *int32, error) {
	input := &kms.GetKeyRotationStatusInput{
		KeyId: aws.String(keyID),
	}

	output, err := conn.GetKeyRotationStatus(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	if output == nil {
		return nil, nil, tfresource.NewEmptyResultError(input)
	}

	return aws.Bool(output.KeyRotationEnabled), output.RotationPeriodInDays, nil
}

func updateKeyDescription(ctx context.Context, conn *kms.Client, resourceTypeName, keyID, description string) error {
	input := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyID),
	}

	_, err := conn.UpdateKeyDescription(ctx, input)

	if err != nil {
		return fmt.Errorf("updating %s (%s) description: %w", resourceTypeName, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyDescriptionPropagated(ctx, conn, keyID, description); err != nil {
		return fmt.Errorf("waiting for %s (%s) description update: %w", resourceTypeName, keyID, err)
	}

	return nil
}

func updateKeyEnabled(ctx context.Context, conn *kms.Client, resourceTypeName, keyID string, enabled bool) error {
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

	if _, err := tfresource.RetryWhenIsA[*awstypes.NotFoundException](ctx, propagationTimeout, updateFunc); err != nil {
		return fmt.Errorf("%s %s (%s): %w", action, resourceTypeName, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyStatePropagated(ctx, conn, keyID, enabled); err != nil {
		return fmt.Errorf("waiting for %s (%s) update (enabled = %t): %w", resourceTypeName, keyID, enabled, err)
	}

	return nil
}

func updateKeyPolicy(ctx context.Context, conn *kms.Client, resourceTypeName, keyID, policy string, bypassPolicyLockoutSafetyCheck bool) error {
	policy, err := structure.NormalizeJsonString(policy)
	if err != nil {
		return fmt.Errorf("policy contains invalid JSON: %w", err)
	}

	updateFunc := func() (interface{}, error) {
		var err error

		input := &kms.PutKeyPolicyInput{
			BypassPolicyLockoutSafetyCheck: bypassPolicyLockoutSafetyCheck,
			KeyId:                          aws.String(keyID),
			Policy:                         aws.String(policy),
			PolicyName:                     aws.String(policyNameDefault),
		}

		_, err = conn.PutKeyPolicy(ctx, input)

		return nil, err
	}

	if _, err := tfresource.RetryWhenIsOneOf2[*awstypes.NotFoundException, *awstypes.MalformedPolicyDocumentException](ctx, propagationTimeout, updateFunc); err != nil {
		return fmt.Errorf("updating %s (%s) policy: %w", resourceTypeName, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyPolicyPropagated(ctx, conn, keyID, policy); err != nil {
		return fmt.Errorf("waiting for %s (%s) policy update: %w", resourceTypeName, keyID, err)
	}

	return nil
}

func updateKeyRotationEnabled(ctx context.Context, conn *kms.Client, resourceTypeName, keyID string, enabled bool, rotationPeriod int) error {
	var action string

	updateFunc := func() (interface{}, error) {
		var err error

		if enabled {
			action = "enabling"
			input := kms.EnableKeyRotationInput{
				KeyId: aws.String(keyID),
			}
			if rotationPeriod > 0 {
				input.RotationPeriodInDays = aws.Int32(int32(rotationPeriod))
			}
			_, err = conn.EnableKeyRotation(ctx, &input)
		} else {
			action = "disabling"
			_, err = conn.DisableKeyRotation(ctx, &kms.DisableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	if _, err := tfresource.RetryWhenIsOneOf2[*awstypes.NotFoundException, *awstypes.DisabledException](ctx, keyRotationUpdatedTimeout, updateFunc); err != nil {
		return fmt.Errorf("%s %s (%s) rotation: %w", action, resourceTypeName, keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	if err := waitKeyRotationEnabledPropagated(ctx, conn, keyID, enabled, rotationPeriod); err != nil {
		return fmt.Errorf("waiting for %s (%s) rotation update: %w", resourceTypeName, keyID, err)
	}

	return nil
}

func statusKeyState(ctx context.Context, conn *kms.Client, keyID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findKeyByID(ctx, conn, keyID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.KeyState), nil
	}
}

func waitKeyDescriptionPropagated(ctx context.Context, conn *kms.Client, keyID string, description string) error {
	checkFunc := func() (bool, error) {
		output, err := findKeyByID(ctx, conn, keyID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.ToString(output.Description) == description, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}
	const (
		timeout = 10 * time.Minute
	)

	return tfresource.WaitUntil(ctx, timeout, checkFunc, opts)
}

func waitKeyDeleted(ctx context.Context, conn *kms.Client, keyID string) (*awstypes.KeyMetadata, error) { //nolint:unparam
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KeyStateDisabled, awstypes.KeyStateEnabled),
		Target:  []string{},
		Refresh: statusKeyState(ctx, conn, keyID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func waitKeyPolicyPropagated(ctx context.Context, conn *kms.Client, keyID, policy string) error {
	checkFunc := func() (bool, error) {
		output, err := findKeyPolicyByTwoPartKey(ctx, conn, keyID, policyNameDefault)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(aws.ToString(output), policy)

		if err != nil {
			return false, err
		}

		return equivalent, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}
	const (
		timeout = 10 * time.Minute
	)

	return tfresource.WaitUntil(ctx, timeout, checkFunc, opts)
}

func waitKeyRotationEnabledPropagated(ctx context.Context, conn *kms.Client, keyID string, enabled bool, rotationPeriodWant int) error {
	checkFunc := func() (bool, error) {
		rotation, rotationPeriodGot, err := findKeyRotationEnabledByKeyID(ctx, conn, keyID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if rotationPeriodWant != 0 && rotationPeriodGot != nil {
			return aws.ToBool(rotation) == enabled && aws.ToInt32(rotationPeriodGot) == int32(rotationPeriodWant), nil
		}

		return aws.ToBool(rotation) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(ctx, keyRotationUpdatedTimeout, checkFunc, opts)
}

func waitKeyStatePropagated(ctx context.Context, conn *kms.Client, keyID string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := findKeyByID(ctx, conn, keyID)

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
	const (
		timeout = 20 * time.Minute
	)

	return tfresource.WaitUntil(ctx, timeout, checkFunc, opts)
}
