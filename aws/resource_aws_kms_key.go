package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		if err := updateKmsKeyRotationStatus(conn, d); err != nil {
			return err
		}
	}

	if enabled := d.Get("is_enabled").(bool); !enabled {
		if err := updateKmsKeyStatus(conn, d.Id(), enabled); err != nil {
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

	if d.HasChange("is_enabled") && d.Get("is_enabled").(bool) {
		// Enable before any attributes will be modified
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_key_rotation") {
		if err := updateKmsKeyRotationStatus(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		if err := resourceAwsKmsKeyDescriptionUpdate(conn, d); err != nil {
			return err
		}
	}
	if d.HasChange("policy") {
		if err := resourceAwsKmsKeyPolicyUpdate(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("is_enabled") && !d.Get("is_enabled").(bool) {
		// Only disable when all attributes are modified
		// because we cannot modify disabled keys
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
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

func resourceAwsKmsKeyDescriptionUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	description := d.Get("description").(string)
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] KMS key: %s, update description: %s", keyId, description)

	req := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyId),
	}
	_, err := conn.UpdateKeyDescription(req)

	return err
}

func resourceAwsKmsKeyPolicyUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %w", err)
	}
	keyId := d.Get("key_id").(string)
	bypassPolicyLockoutCheck := d.Get("bypass_policy_lockout_check").(bool)

	log.Printf("[DEBUG] KMS key: %s, bypass policy lockout check: %t, update policy: %s", keyId, bypassPolicyLockoutCheck, policy)

	req := &kms.PutKeyPolicyInput{
		KeyId:                          aws.String(keyId),
		Policy:                         aws.String(policy),
		PolicyName:                     aws.String("default"),
		BypassPolicyLockoutSafetyCheck: aws.Bool(bypassPolicyLockoutCheck),
	}
	_, err = conn.PutKeyPolicy(req)

	return err
}

func updateKmsKeyStatus(conn *kms.KMS, id string, shouldBeEnabled bool) error {
	var err error

	if shouldBeEnabled {
		log.Printf("[DEBUG] Enabling KMS key %q", id)
		_, err = conn.EnableKey(&kms.EnableKeyInput{
			KeyId: aws.String(id),
		})
	} else {
		log.Printf("[DEBUG] Disabling KMS key %q", id)
		_, err = conn.DisableKey(&kms.DisableKeyInput{
			KeyId: aws.String(id),
		})
	}

	if err != nil {
		return fmt.Errorf("Failed to set KMS key %q status to %t: %q",
			id, shouldBeEnabled, err.Error())
	}

	// Wait for propagation since KMS is eventually consistent
	wait := resource.StateChangeConf{
		Pending:                   []string{fmt.Sprintf("%t", !shouldBeEnabled)},
		Target:                    []string{fmt.Sprintf("%t", shouldBeEnabled)},
		Timeout:                   20 * time.Minute,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 15,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if KMS key %s enabled status is %t",
				id, shouldBeEnabled)
			resp, err := conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
					return nil, fmt.Sprintf("%t", !shouldBeEnabled), nil
				}
				return resp, "FAILED", err
			}
			status := fmt.Sprintf("%t", *resp.KeyMetadata.Enabled)
			log.Printf("[DEBUG] KMS key %s status received: %s, retrying", id, status)

			return resp, status, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed setting KMS key status to %t: %w", shouldBeEnabled, err)
	}

	return nil
}

func updateKmsKeyRotationStatus(conn *kms.KMS, d *schema.ResourceData) error {
	shouldEnableRotation := d.Get("enable_key_rotation").(bool)

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		err := handleKeyRotation(conn, shouldEnableRotation, aws.String(d.Id()))

		if err != nil {
			if isAWSErr(err, kms.ErrCodeDisabledException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		err = handleKeyRotation(conn, shouldEnableRotation, aws.String(d.Id()))
	}

	if err != nil {
		return fmt.Errorf("Failed to set key rotation for %q to %t: %q",
			d.Id(), shouldEnableRotation, err.Error())
	}

	// Wait for propagation since KMS is eventually consistent
	wait := resource.StateChangeConf{
		Pending:                   []string{fmt.Sprintf("%t", !shouldEnableRotation)},
		Target:                    []string{fmt.Sprintf("%t", shouldEnableRotation)},
		Timeout:                   5 * time.Minute,
		MinTimeout:                1 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if KMS key %s rotation status is %t",
				d.Id(), shouldEnableRotation)

			out, err := retryOnAwsCode(kms.ErrCodeNotFoundException, func() (interface{}, error) {
				return conn.GetKeyRotationStatus(&kms.GetKeyRotationStatusInput{
					KeyId: aws.String(d.Id()),
				})
			})
			if err != nil {
				return 42, "", err
			}
			resp, _ := out.(*kms.GetKeyRotationStatusOutput)

			status := fmt.Sprintf("%t", *resp.KeyRotationEnabled)
			log.Printf("[DEBUG] KMS key %s rotation status received: %s, retrying", d.Id(), status)

			return resp, status, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed setting KMS key rotation status to %t: %s", shouldEnableRotation, err)
	}

	return nil
}

func handleKeyRotation(conn *kms.KMS, shouldEnableRotation bool, keyId *string) error {
	var err error
	if shouldEnableRotation {
		log.Printf("[DEBUG] Enabling key rotation for KMS key %q", *keyId)
		_, err = conn.EnableKeyRotation(&kms.EnableKeyRotationInput{
			KeyId: keyId,
		})
	} else {
		log.Printf("[DEBUG] Disabling key rotation for KMS key %q", *keyId)
		_, err = conn.DisableKeyRotation(&kms.DisableKeyRotationInput{
			KeyId: keyId,
		})
	}
	return err
}

func resourceAwsKmsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	keyId := d.Get("key_id").(string)

	req := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(keyId),
	}
	if v, exists := d.GetOk("deletion_window_in_days"); exists {
		req.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}
	_, err := conn.ScheduleKeyDeletion(req)

	if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error scheduling deletion for KMS Key (%s): %w", d.Id(), err)
	}

	//  Error: error scheduling deletion for KMS Key (a93d0bf5-ff4c-4323-9f9a-2151ac639254): KMSInvalidStateException: arn:aws:kms:us-west-2:346386234494:key/a93d0bf5-ff4c-4323-9f9a-2151ac639254 is pending deletion.

	_, err = waiter.KeyStatePendingDeletion(conn, d.Id())

	if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) to schedule deletion: %w", d.Id(), err)
	}

	return nil
}
