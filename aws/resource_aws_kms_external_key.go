package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsKmsExternalKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKmsExternalKeyCreate,
		Read:   resourceAwsKmsExternalKeyRead,
		Update: resourceAwsKmsExternalKeyUpdate,
		Delete: resourceAwsKmsExternalKeyDelete,
		Exists: resourceAwsKmsExternalKeyExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_usage": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !(value == "ENCRYPT_DECRYPT" || value == "") {
						es = append(es, fmt.Errorf(
							"%q must be ENCRYPT_DECRYPT or not specified", k))
					}
					return
				},
			},
			"policy": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validateJsonString,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"is_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				// Cannot enable PendingImport keys
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("key_state") == "PendingImport" {
						return true
					}
					return false
				},
			},
			"deletion_window_in_days": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(int)
					if value > 30 || value < 7 {
						es = append(es, fmt.Errorf(
							"%q must be between 7 and 30 days inclusive", k))
					}
					return
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsKmsExternalKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	// Allow aws to chose default values if we don't pass them
	var req kms.CreateKeyInput
	if v, exists := d.GetOk("description"); exists {
		req.Description = aws.String(v.(string))
	}
	if v, exists := d.GetOk("key_usage"); exists {
		req.KeyUsage = aws.String(v.(string))
	}
	if v, exists := d.GetOk("policy"); exists {
		req.Policy = aws.String(v.(string))
	}
	if v, exists := d.GetOk("tags"); exists {
		req.Tags = tagsFromMapKMS(v.(map[string]interface{}))
	}
	req.Origin = aws.String("EXTERNAL")

	var resp *kms.CreateKeyOutput
	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateKey(&req)
		if isAWSErr(err, "MalformedPolicyDocumentException", "") {
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if err != nil {
		return err
	}

	d.SetId(*resp.KeyMetadata.KeyId)
	d.Set("key_id", resp.KeyMetadata.KeyId)
	d.Set("key_state", resp.KeyMetadata.KeyState)

	return resourceAwsKmsExternalKeyUpdate(d, meta)
}

func resourceAwsKmsExternalKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	req := &kms.DescribeKeyInput{
		KeyId: aws.String(d.Id()),
	}

	var resp *kms.DescribeKeyOutput
	var err error
	if d.IsNewResource() {
		var out interface{}
		out, err = retryOnAwsCode("NotFoundException", func() (interface{}, error) {
			return conn.DescribeKey(req)
		})
		resp, _ = out.(*kms.DescribeKeyOutput)
	} else {
		resp, err = conn.DescribeKey(req)
	}
	if err != nil {
		return err
	}
	metadata := resp.KeyMetadata

	if *metadata.KeyState == "PendingDeletion" {
		log.Printf("[WARN] Removing KMS key %s because it's already gone", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(*metadata.KeyId)

	d.Set("arn", metadata.Arn)
	d.Set("key_id", metadata.KeyId)
	d.Set("key_state", metadata.KeyState)
	d.Set("description", metadata.Description)
	d.Set("key_usage", metadata.KeyUsage)
	d.Set("is_enabled", metadata.Enabled)

	p, err := conn.GetKeyPolicy(&kms.GetKeyPolicyInput{
		KeyId:      metadata.KeyId,
		PolicyName: aws.String("default"),
	})
	if err != nil {
		return err
	}

	policy, err := normalizeJsonString(*p.Policy)
	if err != nil {
		return errwrap.Wrapf("policy contains an invalid JSON: {{err}}", err)
	}
	d.Set("policy", policy)

	tagList, err := conn.ListResourceTags(&kms.ListResourceTagsInput{
		KeyId: metadata.KeyId,
	})
	if err != nil {
		return fmt.Errorf("Failed to get KMS key tags (key: %s): %s", d.Get("key_id").(string), err)
	}
	d.Set("tags", tagsToMapKMS(tagList.Tags))

	return nil
}

func resourceAwsKmsExternalKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	if d.HasChange("is_enabled") && d.Get("is_enabled").(bool) && d.Get("key_state") != "PendingImport" {
		// Enable before any attributes will be modified
		if err := updateKmsExternalKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		if err := resourceAwsKmsExternalKeyDescriptionUpdate(conn, d); err != nil {
			return err
		}
	}
	if d.HasChange("policy") {
		if err := resourceAwsKmsExternalKeyPolicyUpdate(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("is_enabled") && !d.Get("is_enabled").(bool) && d.Get("key_state") != "PendingImport" {
		// Only disable when all attributes are modified
		// because we cannot modify disabled keys
		if err := updateKmsExternalKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
			return err
		}
	}

	if err := setTagsKMS(conn, d, d.Id()); err != nil {
		return err
	}

	return resourceAwsKmsExternalKeyRead(d, meta)
}

func resourceAwsKmsExternalKeyDescriptionUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	description := d.Get("description").(string)
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] KMS key: %s, update description: %s", keyId, description)

	req := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyId),
	}
	_, err := retryOnAwsCode("NotFoundException", func() (interface{}, error) {
		return conn.UpdateKeyDescription(req)
	})
	return err
}

func resourceAwsKmsExternalKeyPolicyUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	policy, err := normalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return errwrap.Wrapf("policy contains an invalid JSON: {{err}}", err)
	}
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] KMS key: %s, update policy: %s", keyId, policy)

	req := &kms.PutKeyPolicyInput{
		KeyId:      aws.String(keyId),
		Policy:     aws.String(policy),
		PolicyName: aws.String("default"),
	}
	_, err = retryOnAwsCode("NotFoundException", func() (interface{}, error) {
		return conn.PutKeyPolicy(req)
	})
	return err
}

func updateKmsExternalKeyStatus(conn *kms.KMS, id string, shouldBeEnabled bool) error {
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
				awsErr, ok := err.(awserr.Error)
				if ok && awsErr.Code() == "NotFoundException" {
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
		return fmt.Errorf("Failed setting KMS key status to %t: %s", shouldBeEnabled, err)
	}

	return nil
}

func resourceAwsKmsExternalKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*AWSClient).kmsconn

	req := &kms.DescribeKeyInput{
		KeyId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeKey(req)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return false, nil
			}
		}
		return false, err
	}
	metadata := resp.KeyMetadata

	if *metadata.KeyState == "PendingDeletion" {
		return false, nil
	}

	return true, nil
}

func resourceAwsKmsExternalKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	keyId := d.Get("key_id").(string)

	req := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(keyId),
	}
	if v, exists := d.GetOk("deletion_window_in_days"); exists {
		req.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}
	_, err := conn.ScheduleKeyDeletion(req)
	if err != nil {
		return err
	}

	// Wait for propagation since KMS is eventually consistent
	wait := resource.StateChangeConf{
		Pending:                   []string{"Enabled", "Disabled"},
		Target:                    []string{"PendingDeletion"},
		Timeout:                   20 * time.Minute,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 10,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if KMS key %s state is PendingDeletion", keyId)
			resp, err := conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: aws.String(keyId),
			})
			if err != nil {
				return resp, "Failed", err
			}

			metadata := *resp.KeyMetadata
			log.Printf("[DEBUG] KMS key %s state is %s, retrying", keyId, *metadata.KeyState)

			return resp, *metadata.KeyState, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed deactivating KMS key %s: %s", keyId, err)
	}

	log.Printf("[DEBUG] KMS Key %s deactivated.", keyId)
	d.SetId("")
	return nil
}
