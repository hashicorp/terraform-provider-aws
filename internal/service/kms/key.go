package kms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyCreate,
		Read:   resourceKeyRead,
		Update: resourceKeyUpdate,
		Delete: resourceKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn
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

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().KmsTags()
	}

	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	log.Printf("[DEBUG] Creating KMS Key: %s", input)

	outputRaw, err := WaitIAMPropagation(func() (interface{}, error) {
		return conn.CreateKey(input)
	})

	if err != nil {
		return fmt.Errorf("error creating KMS Key: %w", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*kms.CreateKeyOutput).KeyMetadata.KeyId))
	d.Set("key_id", d.Id())

	if enableKeyRotation := d.Get("enable_key_rotation").(bool); enableKeyRotation {
		if err := updateKmsKeyRotationEnabled(conn, d.Id(), enableKeyRotation); err != nil {
			return err
		}
	}

	if enabled := d.Get("is_enabled").(bool); !enabled {
		if err := updateKmsKeyEnabled(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk("policy"); ok {
		if err := WaitKeyPolicyPropagated(conn, d.Id(), v.(string)); err != nil {
			return fmt.Errorf("error waiting for KMS Key (%s) policy propagation: %w", d.Id(), err)
		}
	}

	if len(tags) > 0 {
		if err := WaitTagsPropagated(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error waiting for KMS Key (%s) tag propagation: %w", d.Id(), err)
		}
	}

	return resourceKeyRead(d, meta)
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	key, err := findKmsKey(conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("arn", key.metadata.Arn)
	d.Set("customer_master_key_spec", key.metadata.CustomerMasterKeySpec)
	d.Set("description", key.metadata.Description)
	d.Set("enable_key_rotation", key.rotation)
	d.Set("is_enabled", key.metadata.Enabled)
	d.Set("key_id", key.metadata.KeyId)
	d.Set("key_usage", key.metadata.KeyUsage)
	d.Set("policy", key.policy)

	tags := key.tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && enabled {
		// Enable before any attributes are modified.
		if err := updateKmsKeyEnabled(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	if hasChange, enableKeyRotation := d.HasChange("enable_key_rotation"), d.Get("enable_key_rotation").(bool); hasChange {
		if err := updateKmsKeyRotationEnabled(conn, d.Id(), enableKeyRotation); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		if err := updateKmsKeyDescription(conn, d.Id(), d.Get("description").(string)); err != nil {
			return err
		}
	}

	if d.HasChange("policy") {
		if err := updateKmsKeyPolicy(conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return err
		}
	}

	if hasChange, enabled := d.HasChange("is_enabled"), d.Get("is_enabled").(bool); hasChange && !enabled {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKmsKeyEnabled(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.KmsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating KMS Key (%s) tags: %w", d.Id(), err)
		}

		if err := WaitTagsPropagated(conn, d.Id(), tftags.New(n)); err != nil {
			return fmt.Errorf("error waiting for KMS Key (%s) tag propagation: %w", d.Id(), err)
		}
	}

	return resourceKeyRead(d, meta)
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Key: (%s)", d.Id())
	_, err := conn.ScheduleKeyDeletion(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, kms.ErrCodeInvalidStateException, "is pending deletion") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting KMS Key (%s): %w", d.Id(), err)
	}

	if _, err := WaitKeyDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

type kmsKey struct {
	metadata *kms.KeyMetadata
	policy   string
	rotation *bool
	tags     tftags.KeyValueTags
}

func findKmsKey(conn *kms.KMS, keyID string, isNewResource bool) (*kmsKey, error) {
	// Wait for propagation since KMS is eventually consistent.
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		var err error
		var key kmsKey

		key.metadata, err = FindKeyByID(conn, keyID)

		if err != nil {
			return nil, fmt.Errorf("error reading KMS Key (%s): %w", keyID, err)
		}

		policy, err := FindKeyPolicyByKeyIDAndPolicyName(conn, keyID, PolicyNameDefault)

		if err != nil {
			return nil, fmt.Errorf("error reading KMS Key (%s) policy: %w", keyID, err)
		}

		key.policy, err = structure.NormalizeJsonString(aws.StringValue(policy))

		if err != nil {
			return nil, fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		if aws.StringValue(key.metadata.Origin) == kms.OriginTypeAwsKms {
			key.rotation, err = FindKeyRotationEnabledByKeyID(conn, keyID)

			if err != nil {
				return nil, fmt.Errorf("error reading KMS Key (%s) rotation enabled: %w", keyID, err)
			}
		}

		key.tags, err = tftags.KmsListTags(conn, keyID)

		if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
			return nil, &resource.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, fmt.Errorf("error listing tags for KMS Key (%s): %w", keyID, err)
		}

		return &key, nil
	}, isNewResource)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*kmsKey), nil
}

func updateKmsKeyDescription(conn *kms.KMS, keyID string, description string) error {
	input := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyID),
	}

	log.Printf("[DEBUG] Updating KMS Key description: %s", input)
	_, err := conn.UpdateKeyDescription(input)

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) description: %w", keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyDescriptionPropagated(conn, keyID, description)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) description propagation: %w", keyID, err)
	}

	return nil
}

func updateKmsKeyEnabled(conn *kms.KMS, keyID string, enabled bool) error {
	updateFunc := func() (interface{}, error) {
		var err error

		log.Printf("[DEBUG] Updating KMS Key (%s) key enabled: %t", keyID, enabled)
		if enabled {
			_, err = conn.EnableKey(&kms.EnableKeyInput{
				KeyId: aws.String(keyID),
			})
		} else {
			_, err = conn.DisableKey(&kms.DisableKeyInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	_, err := tfresource.RetryWhenAwsErrCodeEquals(PropagationTimeout, updateFunc, kms.ErrCodeNotFoundException)

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) key enabled (%t): %w", keyID, enabled, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyStatePropagated(conn, keyID, enabled)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) key state propagation: %w", keyID, err)
	}

	return nil
}

func updateKmsKeyPolicy(conn *kms.KMS, keyID string, policy string, bypassPolicyLockoutSafetyCheck bool) error {
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

		log.Printf("[DEBUG] Updating KMS Key policy: %s", input)
		_, err = conn.PutKeyPolicy(input)

		return nil, err
	}

	_, err = tfresource.RetryWhenAwsErrCodeEquals(PropagationTimeout, updateFunc, kms.ErrCodeNotFoundException)

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) policy: %w", keyID, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyPolicyPropagated(conn, keyID, policy)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) policy propagation: %w", keyID, err)
	}

	return nil
}

func updateKmsKeyRotationEnabled(conn *kms.KMS, keyID string, enabled bool) error {
	updateFunc := func() (interface{}, error) {
		var err error

		log.Printf("[DEBUG] Updating KMS Key (%s) key rotation enabled: %t", keyID, enabled)
		if enabled {
			_, err = conn.EnableKeyRotation(&kms.EnableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		} else {
			_, err = conn.DisableKeyRotation(&kms.DisableKeyRotationInput{
				KeyId: aws.String(keyID),
			})
		}

		return nil, err
	}

	_, err := tfresource.RetryWhenAwsErrCodeEquals(KeyRotationUpdatedTimeout, updateFunc, kms.ErrCodeNotFoundException, kms.ErrCodeDisabledException)

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) key rotation enabled (%t): %w", keyID, enabled, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	err = WaitKeyRotationEnabledPropagated(conn, keyID, enabled)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) key rotation propagation: %w", keyID, err)
	}

	return nil
}
