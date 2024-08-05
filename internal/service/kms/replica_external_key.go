// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_kms_replica_external_key", name="Replica External Key")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kms/types;awstypes;awstypes.KeyMetadata")
// @Testing(importIgnore="deletion_window_in_days;bypass_policy_lockout_safety_check;key_material_base64")
// @Testing(altRegionProvider=true)
func resourceReplicaExternalKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicaExternalKeyCreate,
		ReadWithoutTimeout:   resourceReplicaExternalKeyRead,
		UpdateWithoutTimeout: resourceReplicaExternalKeyUpdate,
		DeleteWithoutTimeout: resourceReplicaExternalKeyDelete,

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
			names.AttrKeyID: {
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
			names.AttrPolicy: {
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

func resourceReplicaExternalKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	// e.g. arn:aws:kms:us-east-2:111122223333:key/mrk-1234abcd12ab34cd56ef1234567890ab
	primaryKeyARN, err := arn.Parse(d.Get("primary_key_arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing primary key ARN: %s", err)
	}

	input := &kms.ReplicateKeyInput{
		KeyId:         aws.String(strings.TrimPrefix(primaryKeyARN.Resource, "key/")),
		ReplicaRegion: aws.String(meta.(*conns.AWSClient).Region),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("bypass_policy_lockout_safety_check"); ok {
		input.BypassPolicyLockoutSafetyCheck = v.(bool)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		input.Policy = aws.String(v.(string))
	}

	output, err := waitIAMPropagation(ctx, iamPropagationTimeout, func() (*kms.ReplicateKeyOutput, error) {
		// Replication is initiated in the primary key's Region.
		return conn.ReplicateKey(ctx, input, func(o *kms.Options) {
			o.Region = primaryKeyARN.Region
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Replica External Key: %s", err)
	}

	d.SetId(aws.ToString(output.ReplicaKeyMetadata.KeyId))

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if _, err := waitReplicaExternalKeyCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("key_material_base64"); ok {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(ctx, conn, "KMS Replica External Key", d.Id(), v.(string), validTo); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitKeyMaterialImported(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) material import: %s", d.Id(), err)
		}

		if err := waitKeyValidToPropagated(ctx, conn, d.Id(), validTo); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) valid_to update: %s", d.Id(), err)
		}

		// The key can only be disabled if key material has been imported, else:
		// "KMSInvalidStateException: arn:aws:kms:us-west-2:123456789012:key/47e3edc1-945f-413b-88b1-e7341c2d89f7 is pending import."
		if enabled := d.Get(names.AttrEnabled).(bool); !enabled {
			if err := updateKeyEnabled(ctx, conn, "KMS Replica External Key", d.Id(), enabled); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	// Wait for propagation since KMS is eventually consistent.
	if v, ok := d.GetOk(names.AttrPolicy); ok {
		if err := waitKeyPolicyPropagated(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) policy update: %s", d.Id(), err)
		}
	}

	if tags := KeyValueTags(ctx, getTagsIn(ctx)); len(tags) > 0 {
		if err := waitTagsPropagated(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) tag update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicaExternalKeyRead(ctx, d, meta)...)
}

func resourceReplicaExternalKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	key, err := findKeyInfo(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS External Replica Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS External Replica Key (%s): %s", d.Id(), err)
	}

	if keyManager := key.metadata.KeyManager; keyManager != awstypes.KeyManagerTypeCustomer {
		return sdkdiag.AppendErrorf(diags, "KMS External Replica Key (%s) has invalid KeyManager: %s", d.Id(), keyManager)
	}

	if origin := key.metadata.Origin; origin != awstypes.OriginTypeExternal {
		return sdkdiag.AppendErrorf(diags, "KMS External Replica Key (%s) has invalid Origin: %s", d.Id(), origin)
	}

	if !aws.ToBool(key.metadata.MultiRegion) || key.metadata.MultiRegionConfiguration.MultiRegionKeyType != awstypes.MultiRegionKeyTypeReplica {
		return sdkdiag.AppendErrorf(diags, "KMS External Replica Key (%s) is not a multi-Region replica key", d.Id())
	}

	d.Set(names.AttrARN, key.metadata.Arn)
	d.Set(names.AttrDescription, key.metadata.Description)
	d.Set(names.AttrEnabled, key.metadata.Enabled)
	d.Set("expiration_model", key.metadata.ExpirationModel)
	d.Set(names.AttrKeyID, key.metadata.KeyId)
	d.Set("key_state", key.metadata.KeyState)
	d.Set("key_usage", key.metadata.KeyUsage)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), key.policy)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	d.Set("primary_key_arn", key.metadata.MultiRegionConfiguration.PrimaryKey.Arn)
	if key.metadata.ValidTo != nil {
		d.Set("valid_to", aws.ToTime(key.metadata.ValidTo).Format(time.RFC3339))
	} else {
		d.Set("valid_to", nil)
	}

	setTagsOut(ctx, key.tags)

	return diags
}

func resourceReplicaExternalKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	if hasChange, enabled, state := d.HasChange(names.AttrEnabled), d.Get(names.AttrEnabled).(bool), awstypes.KeyState(d.Get("key_state").(string)); hasChange && enabled && state != awstypes.KeyStatePendingImport {
		// Enable before any attributes are modified.
		if err := updateKeyEnabled(ctx, conn, "KMS Replica External Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrDescription) {
		if err := updateKeyDescription(ctx, conn, "KMS Replica External Key", d.Id(), d.Get(names.AttrDescription).(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrPolicy) {
		if err := updateKeyPolicy(ctx, conn, "KMS Replica External Key", d.Id(), d.Get(names.AttrPolicy).(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("valid_to") {
		validTo := d.Get("valid_to").(string)

		if err := importExternalKeyMaterial(ctx, conn, "KMS Replica External Key", d.Id(), d.Get("key_material_base64").(string), validTo); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitKeyMaterialImported(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Replica Key (%s) material import: %s", d.Id(), err)
		}

		if err := waitKeyValidToPropagated(ctx, conn, d.Id(), validTo); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for KMS External Replica Key (%s) valid_to update: %s", d.Id(), err)
		}
	}

	if hasChange, enabled, state := d.HasChange(names.AttrEnabled), d.Get(names.AttrEnabled).(bool), awstypes.KeyState(d.Get("key_state").(string)); hasChange && !enabled && state != awstypes.KeyStatePendingImport {
		// Only disable after all attributes have been modified because we cannot modify disabled keys.
		if err := updateKeyEnabled(ctx, conn, "KMS Replica External Key", d.Id(), enabled); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicaExternalKeyRead(ctx, d, meta)...)
}

func resourceReplicaExternalKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	input := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("deletion_window_in_days"); ok {
		input.PendingWindowInDays = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Deleting KMS Replica External Key: %s", d.Id())
	_, err := conn.ScheduleKeyDeletion(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.KMSInvalidStateException](err, "is pending deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Replica External Key (%s): %s", d.Id(), err)
	}

	if _, err := waitKeyDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Replica External Key (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitReplicaExternalKeyCreated(ctx context.Context, conn *kms.Client, id string) (*awstypes.KeyMetadata, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KeyStateCreating),
		Target:  enum.Slice(awstypes.KeyStatePendingImport),
		Refresh: statusKeyState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}
