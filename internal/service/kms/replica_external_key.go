package kms

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicaExternalKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicaExternalKeyCreate,
		Read:   resourceReplicaExternalKeyRead,
		Update: resourceReplicaExternalKeyUpdate,
		Delete: resourceReplicaExternalKeyDelete,

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
			"deletion_window_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 8192),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"expiration_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
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
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"primary_key_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"valid_to": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
		},
	}
}

func resourceReplicaExternalKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// e.g. arn:aws:kms:us-east-2:111122223333:key/mrk-1234abcd12ab34cd56ef1234567890ab
	primaryKeyARN, err := arn.Parse(d.Get("primary_key_arn").(string))

	if err != nil {
		return fmt.Errorf("error parsing primary key ARN: %w", err)
	}

	input := &kms.ReplicateKeyInput{
		KeyId:         aws.String(strings.TrimPrefix(primaryKeyARN.Resource, "key/")),
		ReplicaRegion: aws.String(meta.(*conns.AWSClient).Region),
	}

	if v, ok := d.GetOk("bypass_policy_lockout_safety_check"); ok {
		input.BypassPolicyLockoutSafetyCheck = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	// Replication is initiated in the primary key's region.
	session, err := conns.NewSessionForRegion(&conn.Config, primaryKeyARN.Region, meta.(*conns.AWSClient).TerraformVersion)

	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	replicateConn := kms.New(session)

	log.Printf("[DEBUG] Creating KMS Replica External Key: %s", input)
	outputRaw, err := WaitIAMPropagation(func() (interface{}, error) {
		return replicateConn.ReplicateKey(input)
	})

	if err != nil {
		return fmt.Errorf("error creating KMS Replica External Key: %w", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*kms.ReplicateKeyOutput).ReplicaKeyMetadata.KeyId))

	if _, err := WaitReplicaExternalKeyCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for KMS Replica External Key (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("key_material_base64"); ok {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(conn, d.Id(), v.(string), validTo); err != nil {
			return fmt.Errorf("error importing KMS Replica External Key (%s) material: %w", d.Id(), err)
		}

		if _, err := WaitKeyMaterialImported(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for KMS Replica External Key (%s) material import: %w", d.Id(), err)
		}

		if err := WaitKeyValidToPropagated(conn, d.Id(), validTo); err != nil {
			return fmt.Errorf("error waiting for KMS Replica External Key (%s) valid_to propagation: %w", d.Id(), err)
		}

		// The key can only be disabled if key material has been imported, else:
		// "KMSInvalidStateException: arn:aws:kms:us-west-2:123456789012:key/47e3edc1-945f-413b-88b1-e7341c2d89f7 is pending import."
		if enabled := d.Get("enabled").(bool); !enabled {
			if err := updateKeyEnabled(conn, d.Id(), enabled); err != nil {
				return err
			}
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk("policy"); ok {
		if err := WaitKeyPolicyPropagated(conn, d.Id(), v.(string)); err != nil {
			return fmt.Errorf("error waiting for KMS Replica External Key (%s) policy propagation: %w", d.Id(), err)
		}
	}

	if len(tags) > 0 {
		if err := WaitTagsPropagated(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error waiting for KMS Replica External Key (%s) tag propagation: %w", d.Id(), err)
		}
	}

	return resourceReplicaExternalKeyRead(d, meta)
}

func resourceReplicaExternalKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	key, err := findKey(conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS External Replica Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	if keyManager := aws.StringValue(key.metadata.KeyManager); keyManager != kms.KeyManagerTypeCustomer {
		return fmt.Errorf("KMS External Replica Key (%s) has invalid KeyManager: %s", d.Id(), keyManager)
	}

	if origin := aws.StringValue(key.metadata.Origin); origin != kms.OriginTypeExternal {
		return fmt.Errorf("KMS External Replica Key (%s) has invalid Origin: %s", d.Id(), origin)
	}

	if !aws.BoolValue(key.metadata.MultiRegion) ||
		aws.StringValue(key.metadata.MultiRegionConfiguration.MultiRegionKeyType) != kms.MultiRegionKeyTypeReplica {
		return fmt.Errorf("KMS External Replica Key (%s) is not a multi-Region replica key", d.Id())
	}

	d.Set("arn", key.metadata.Arn)
	d.Set("description", key.metadata.Description)
	d.Set("enabled", key.metadata.Enabled)
	d.Set("expiration_model", key.metadata.ExpirationModel)
	d.Set("key_id", key.metadata.KeyId)
	d.Set("key_state", key.metadata.KeyState)
	d.Set("key_usage", key.metadata.KeyUsage)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), key.policy)

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", key.policy, err)
	}

	d.Set("policy", policyToSet)

	d.Set("primary_key_arn", key.metadata.MultiRegionConfiguration.PrimaryKey.Arn)
	if key.metadata.ValidTo != nil {
		d.Set("valid_to", aws.TimeValue(key.metadata.ValidTo).Format(time.RFC3339))
	} else {
		d.Set("valid_to", nil)
	}

	tags := key.tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceReplicaExternalKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	if hasChange, enabled, state := d.HasChange("enabled"), d.Get("enabled").(bool), d.Get("key_state").(string); hasChange && enabled && state != kms.KeyStatePendingImport {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		if err := updateKeyDescription(conn, d.Id(), d.Get("description").(string)); err != nil {
			return err
		}
	}

	if d.HasChange("policy") {
		if err := updateKeyPolicy(conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("valid_to") {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(conn, d.Id(), d.Get("key_material_base64").(string), validTo); err != nil {
			return fmt.Errorf("error importing KMS External Replica Key (%s) material: %s", d.Id(), err)
		}

		if _, err := WaitKeyMaterialImported(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for KMS External Replica Key (%s) material import: %w", d.Id(), err)
		}

		if err := WaitKeyValidToPropagated(conn, d.Id(), validTo); err != nil {
			return fmt.Errorf("error waiting for KMS External Replica Key (%s) valid_to propagation: %w", d.Id(), err)
		}
	}

	if hasChange, enabled, state := d.HasChange("enabled"), d.Get("enabled").(bool), d.Get("key_state").(string); hasChange && !enabled && state != kms.KeyStatePendingImport {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating KMS Replica External Key (%s) tags: %w", d.Id(), err)
		}

		if err := WaitTagsPropagated(conn, d.Id(), tftags.New(n)); err != nil {
			return fmt.Errorf("error waiting for KMS Replica External Key (%s) tag propagation: %w", d.Id(), err)
		}
	}

	return resourceReplicaExternalKeyRead(d, meta)
}

func resourceReplicaExternalKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KMSConn

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Replica External Key: (%s)", d.Id())
	_, err := conn.ScheduleKeyDeletion(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, kms.ErrCodeInvalidStateException, "is pending deletion") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting KMS Replica External Key (%s): %w", d.Id(), err)
	}

	if _, err := WaitKeyDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for KMS Replica External Key (%s) delete: %w", d.Id(), err)
	}

	return nil
}
