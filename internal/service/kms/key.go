package kms

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyCreate,
		ReadWithoutTimeout:   resourceKeyRead,
		UpdateWithoutTimeout: resourceKeyUpdate,
		DeleteWithoutTimeout: resourceKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      kms.CustomerMasterKeySpecSymmetricDefault,
				ValidateFunc: validation.StringInSlice(kms.CustomerMasterKeySpec_Values(), false),
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      kms.KeyUsageTypeEncryptDecrypt,
				ValidateFunc: validation.StringInSlice(kms.KeyUsageType_Values(), false),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &kms.CreateKeyInput{
		BypassPolicyLockoutSafetyCheck: aws.Bool(d.Get("bypass_policy_lockout_safety_check").(bool)),
		CustomerMasterKeySpec:          aws.String(d.Get("customer_master_key_spec").(string)),
		KeyUsage:                       aws.String(d.Get("key_usage").(string)),
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
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", p, err)
		}

		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_key_store_id"); ok {
		input.Origin = aws.String(kms.OriginTypeAwsCloudhsm)
		input.CustomKeyStoreId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	log.Printf("[DEBUG] Creating KMS Key: %s", input)

	outputRaw, err := WaitIAMPropagation(ctx, func() (interface{}, error) {
		return conn.CreateKeyWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Key: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*kms.CreateKeyOutput).KeyMetadata.KeyId))

	if enableKeyRotation := d.Get("enable_key_rotation").(bool); enableKeyRotation {
		if err := updateKeyRotationEnabled(ctx, conn, d.Id(), enableKeyRotation); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating KMS Key (%s): %s", d.Id(), err)
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

	if len(tags) > 0 {
		if err := WaitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) tag propagation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), key.policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", key.policy, err)
	}

	d.Set("policy", policyToSet)

	tags := key.tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && enabled {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, enableKeyRotation := d.HasChange("enable_key_rotation"), d.Get("enable_key_rotation").(bool); hasChange {
		if err := updateKeyRotationEnabled(ctx, conn, d.Id(), enableKeyRotation); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		if err := updateKeyDescription(ctx, conn, d.Id(), d.Get("description").(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		if err := updateKeyPolicy(ctx, conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && !enabled {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Key (%s) tags: %s", d.Id(), err)
		}

		if err := WaitTagsPropagated(ctx, conn, d.Id(), tftags.New(n)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Key (%s) tag propagation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

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
	tags     tftags.KeyValueTags
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
			key.rotation, err = FindKeyRotationEnabledByKeyID(ctx, conn, keyID)

			if err != nil {
				return nil, fmt.Errorf("reading KMS Key (%s) rotation enabled: %w", keyID, err)
			}
		}

		key.tags, err = ListTags(ctx, conn, keyID)

		if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
			return nil, &resource.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, fmt.Errorf("listing tags for KMS Key (%s): %w", keyID, err)
		}

		return &key, nil
	}, isNewResource)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*kmsKey), nil
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

func updateKeyEnabled(ctx context.Context, conn *kms.KMS, keyID string, enabled bool) error {
	var action string

	updateFunc := func() (interface{}, error) {
		var err error

		if enabled {
			log.Printf("[DEBUG] Enabling KMS Key (%s)", keyID)
			action = "enabling"
			_, err = conn.EnableKeyWithContext(ctx, &kms.EnableKeyInput{
				KeyId: aws.String(keyID),
			})
		} else {
			log.Printf("[DEBUG] Disabling KMS Key (%s)", keyID)
			action = "disabling"
			_, err = conn.DisableKeyWithContext(ctx, &kms.DisableKeyInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, PropagationTimeout, updateFunc, kms.ErrCodeNotFoundException)
	if err != nil {
		return fmt.Errorf("%s KMS Key: %w", action, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyStatePropagated(ctx, conn, keyID, enabled)

	if err != nil {
		return fmt.Errorf("%s KMS Key: waiting for completion: %w", action, err)
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

func updateKeyRotationEnabled(ctx context.Context, conn *kms.KMS, keyID string, enabled bool) error {
	var action string

	updateFunc := func() (interface{}, error) {
		var err error

		if enabled {
			log.Printf("[DEBUG] Enabling KMS Key (%s) key rotation", keyID)
			_, err = conn.EnableKeyRotationWithContext(ctx, &kms.EnableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		} else {
			log.Printf("[DEBUG] Disabling KMS Key (%s) key rotation", keyID)
			_, err = conn.DisableKeyRotationWithContext(ctx, &kms.DisableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, KeyRotationUpdatedTimeout, updateFunc, kms.ErrCodeNotFoundException, kms.ErrCodeDisabledException)
	if err != nil {
		return fmt.Errorf("%s key rotation: %w", action, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyRotationEnabledPropagated(ctx, conn, keyID, enabled)

	if err != nil {
		return fmt.Errorf("%s key rotation: waiting for completion: %w", action, err)
	}

	return nil
}
