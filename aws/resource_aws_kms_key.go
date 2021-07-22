package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfkms "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsKmsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKmsKeyCreate,
		Read:   resourceAwsKmsKeyRead,
		Update: resourceAwsKmsKeyUpdate,
		Delete: resourceAwsKmsKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"bypass_policy_lockout_check": {
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
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsKmsKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &kms.CreateKeyInput{
		BypassPolicyLockoutSafetyCheck: aws.Bool(d.Get("bypass_policy_lockout_check").(bool)),
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

	outputRaw, err := waiter.IAMPropagation(func() (interface{}, error) {
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

	return resourceAwsKmsKeyRead(d, meta)
}

func resourceAwsKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	type Key struct {
		metadata *kms.KeyMetadata
		policy   string
		rotation *bool
		tags     keyvaluetags.KeyValueTags
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(waiter.PropagationTimeout, func() (interface{}, error) {
		var err error
		var key Key

		key.metadata, err = finder.KeyByID(conn, d.Id())

		if err != nil {
			return nil, fmt.Errorf("error reading KMS Key (%s): %w", d.Id(), err)
		}

		policy, err := finder.KeyPolicyByKeyIDAndPolicyName(conn, d.Id(), tfkms.PolicyNameDefault)

		if err != nil {
			return nil, fmt.Errorf("error reading KMS Key (%s) policy: %w", d.Id(), err)
		}

		key.policy, err = structure.NormalizeJsonString(aws.StringValue(policy))

		if err != nil {
			return nil, fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		key.rotation, err = finder.KeyRotationEnabledByKeyID(conn, d.Id())

		if err != nil {
			return nil, fmt.Errorf("error reading KMS Key (%s) rotation enabled: %w", d.Id(), err)
		}

		key.tags, err = keyvaluetags.KmsListTags(conn, d.Id())

		if err != nil {
			return nil, fmt.Errorf("error listing tags for KMS Key (%s): %w", d.Id(), err)
		}

		return &key, nil
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	key := outputRaw.(*Key)

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

func resourceAwsKmsKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

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
		input := &kms.UpdateKeyDescriptionInput{
			Description: aws.String(d.Get("description").(string)),
			KeyId:       aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating KMS Key description: %s", input)
		_, err := conn.UpdateKeyDescription(input)

		if err != nil {
			return fmt.Errorf("error updating KMS Key (%s) description: %w", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

		if err != nil {
			return fmt.Errorf("policy contains invalid JSON: %w", err)
		}

		input := &kms.PutKeyPolicyInput{
			BypassPolicyLockoutSafetyCheck: aws.Bool(d.Get("bypass_policy_lockout_check").(bool)),
			KeyId:                          aws.String(d.Id()),
			Policy:                         aws.String(policy),
			PolicyName:                     aws.String(tfkms.PolicyNameDefault),
		}

		log.Printf("[DEBUG] Updating KMS Key policy: %s", input)
		_, err = conn.PutKeyPolicy(input)

		if err != nil {
			return fmt.Errorf("error updating KMS Key (%s) policy: %w", d.Id(), err)
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

		if err := keyvaluetags.KmsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating KMS Key (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsKmsKeyRead(d, meta)
}

func resourceAwsKmsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

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

	if _, err := waiter.KeyDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func updateKmsKeyEnabled(conn *kms.KMS, keyID string, enabled bool) error {
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

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) key enabled (%t): %w", keyID, enabled, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	checkFunc := func() (bool, error) {
		output, err := finder.KeyByID(conn, keyID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output.Enabled) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 15,
		MinTimeout:                2 * time.Second,
	}

	err = tfresource.WaitUntil(waiter.KeyStatePropagationTimeout, checkFunc, opts)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) key state propagation: %w", keyID, err)
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

	_, err := tfresource.RetryWhenAwsErrCodeEquals(waiter.KeyRotationUpdatedTimeout, updateFunc, kms.ErrCodeNotFoundException, kms.ErrCodeDisabledException)

	if err != nil {
		return fmt.Errorf("error updating KMS Key (%s) key rotation enabled (%t): %w", keyID, enabled, err)
	}

	// Wait for propagation since KMS is eventually consistent.
	checkFunc := func() (bool, error) {
		output, err := finder.KeyRotationEnabledByKeyID(conn, keyID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	err = tfresource.WaitUntil(waiter.PropagationTimeout, checkFunc, opts)

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) key rotation propagation: %w", keyID, err)
	}

	return nil
}
