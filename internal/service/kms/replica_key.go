package kms

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicaKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicaKeyCreate,
		ReadWithoutTimeout:   resourceReplicaKeyRead,
		UpdateWithoutTimeout: resourceReplicaKeyUpdate,
		DeleteWithoutTimeout: resourceReplicaKeyDelete,

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
				Default:  true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"key_spec": {
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
		},
	}
}

func resourceReplicaKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// e.g. arn:aws:kms:us-east-2:111122223333:key/mrk-1234abcd12ab34cd56ef1234567890ab
	primaryKeyARN, err := arn.Parse(d.Get("primary_key_arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing primary key ARN: %s", err)
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
		return sdkdiag.AppendErrorf(diags, "creating AWS session: %s", err)
	}

	replicateConn := kms.New(session)

	log.Printf("[DEBUG] Creating KMS Replica Key: %s", input)
	outputRaw, err := WaitIAMPropagation(ctx, func() (interface{}, error) {
		return replicateConn.ReplicateKeyWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Replica Key: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*kms.ReplicateKeyOutput).ReplicaKeyMetadata.KeyId))

	if _, err := WaitReplicaKeyCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica Key (%s) create: %s", d.Id(), err)
	}

	d.Set("key_id", d.Id())

	if enabled := d.Get("enabled").(bool); !enabled {
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating KMS Replica Key (%s): %s", d.Id(), err)
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk("policy"); ok {
		if err := WaitKeyPolicyPropagated(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica Key (%s) policy propagation: %s", d.Id(), err)
		}
	}

	if len(tags) > 0 {
		if err := WaitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica Key (%s) tag propagation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicaKeyRead(ctx, d, meta)...)
}

func resourceReplicaKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	key, err := findKey(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Replica Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Replica Key (%s): %s", d.Id(), err)
	}

	if keyManager := aws.StringValue(key.metadata.KeyManager); keyManager != kms.KeyManagerTypeCustomer {
		return sdkdiag.AppendErrorf(diags, "KMS Replica Key (%s) has invalid KeyManager: %s", d.Id(), keyManager)
	}

	if origin := aws.StringValue(key.metadata.Origin); origin != kms.OriginTypeAwsKms {
		return sdkdiag.AppendErrorf(diags, "KMS Replica Key (%s) has invalid Origin: %s", d.Id(), origin)
	}

	if !aws.BoolValue(key.metadata.MultiRegion) ||
		aws.StringValue(key.metadata.MultiRegionConfiguration.MultiRegionKeyType) != kms.MultiRegionKeyTypeReplica {
		return sdkdiag.AppendErrorf(diags, "KMS Replica Key (%s) is not a multi-Region replica key", d.Id())
	}

	d.Set("arn", key.metadata.Arn)
	d.Set("description", key.metadata.Description)
	d.Set("enabled", key.metadata.Enabled)
	d.Set("key_id", key.metadata.KeyId)
	d.Set("key_rotation_enabled", key.rotation)
	d.Set("key_spec", key.metadata.KeySpec)
	d.Set("key_usage", key.metadata.KeyUsage)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), key.policy)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", key.policy, err)
	}

	d.Set("policy", policyToSet)

	d.Set("primary_key_arn", key.metadata.MultiRegionConfiguration.PrimaryKey.Arn)

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

func resourceReplicaKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	if hasChange, enabled := d.HasChange("enabled"), d.Get("enabled").(bool); hasChange && enabled {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Replica Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		if err := updateKeyDescription(ctx, conn, d.Id(), d.Get("description").(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Replica Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		if err := updateKeyPolicy(ctx, conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Replica Key (%s): %s", d.Id(), err)
		}
	}

	if hasChange, enabled := d.HasChange("enabled"), d.Get("enabled").(bool); hasChange && !enabled {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, d.Id(), enabled); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Replica Key (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Replica Key (%s) tags: %s", d.Id(), err)
		}

		if err := WaitTagsPropagated(ctx, conn, d.Id(), tftags.New(n)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica Key (%s) tag propagation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicaKeyRead(ctx, d, meta)...)
}

func resourceReplicaKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Replica Key: (%s)", d.Id())
	_, err := conn.ScheduleKeyDeletionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return diags
	}

	if tfawserr.ErrMessageContains(err, kms.ErrCodeInvalidStateException, "is pending deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Replica Key (%s): %s", d.Id(), err)
	}

	if _, err := WaitKeyDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica Key (%s) delete: %s", d.Id(), err)
	}

	return diags
}
