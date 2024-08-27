// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	keySigningKeyResourceIDPartCount = 2
)

const (
	keySigningKeyStatusActionNeeded    = "ACTION_NEEDED"
	keySigningKeyStatusActive          = "ACTIVE"
	keySigningKeyStatusDeleting        = "DELETING"
	keySigningKeyStatusInactive        = "INACTIVE"
	keySigningKeyStatusInternalFailure = "INTERNAL_FAILURE"
)

// @SDKResource("aws_route53_key_signing_key", name="Key Signing Key")
func resourceKeySigningKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeySigningKeyCreate,
		ReadWithoutTimeout:   resourceKeySigningKeyRead,
		UpdateWithoutTimeout: resourceKeySigningKeyUpdate,
		DeleteWithoutTimeout: resourceKeySigningKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"digest_algorithm_mnemonic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"digest_algorithm_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"digest_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dnskey_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ds_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flag": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_management_service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"key_tag": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z_.-]"), "must contain only alphanumeric characters, periods, underscores, or hyphens"),
				),
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithm_mnemonic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithm_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  keySigningKeyStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					keySigningKeyStatusActive,
					keySigningKeyStatusInactive,
				}, false),
			},
		},
	}
}

func resourceKeySigningKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	hostedZoneID := d.Get(names.AttrHostedZoneID).(string)
	name := d.Get(names.AttrName).(string)
	status := d.Get(names.AttrStatus).(string)
	id := errs.Must(flex.FlattenResourceId([]string{hostedZoneID, name}, keySigningKeyResourceIDPartCount, false))
	input := &route53.CreateKeySigningKeyInput{
		CallerReference: aws.String(sdkid.UniqueId()),
		HostedZoneId:    aws.String(hostedZoneID),
		Name:            aws.String(name),
		Status:          aws.String(status),
	}

	if v, ok := d.GetOk("key_management_service_arn"); ok {
		input.KeyManagementServiceArn = aws.String(v.(string))
	}

	output, err := conn.CreateKeySigningKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Key Signing Key (%s): %s", id, err)
	}

	d.SetId(id)

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) synchronize: %s", d.Id(), err)
		}
	}

	if _, err := waitKeySigningKeyStatusUpdated(ctx, conn, hostedZoneID, name, status); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status update: %s", d.Id(), err)
	}

	return append(diags, resourceKeySigningKeyRead(ctx, d, meta)...)
}

func resourceKeySigningKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), keySigningKeyResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	hostedZoneID, name := parts[0], parts[1]
	keySigningKey, err := findKeySigningKeyByTwoPartKey(ctx, conn, hostedZoneID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Key Signing Key (%s): %s", d.Id(), err)
	}

	d.Set("digest_algorithm_mnemonic", keySigningKey.DigestAlgorithmMnemonic)
	d.Set("digest_algorithm_type", keySigningKey.DigestAlgorithmType)
	d.Set("digest_value", keySigningKey.DigestValue)
	d.Set("dnskey_record", keySigningKey.DNSKEYRecord)
	d.Set("ds_record", keySigningKey.DSRecord)
	d.Set("flag", keySigningKey.Flag)
	d.Set(names.AttrHostedZoneID, hostedZoneID)
	d.Set("key_management_service_arn", keySigningKey.KmsArn)
	d.Set("key_tag", keySigningKey.KeyTag)
	d.Set(names.AttrName, keySigningKey.Name)
	d.Set(names.AttrPublicKey, keySigningKey.PublicKey)
	d.Set("signing_algorithm_mnemonic", keySigningKey.SigningAlgorithmMnemonic)
	d.Set("signing_algorithm_type", keySigningKey.SigningAlgorithmType)
	d.Set(names.AttrStatus, keySigningKey.Status)

	return diags
}

func resourceKeySigningKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), keySigningKeyResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	hostedZoneID, name := parts[0], parts[1]

	if d.HasChange(names.AttrStatus) {
		var changeInfo *awstypes.ChangeInfo
		status := d.Get(names.AttrStatus).(string)

		if status == keySigningKeyStatusActive {
			input := &route53.ActivateKeySigningKeyInput{
				HostedZoneId: aws.String(hostedZoneID),
				Name:         aws.String(name),
			}

			output, err := conn.ActivateKeySigningKey(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "activating Route 53 Key Signing Key (%s): %s", d.Id(), err)
			}

			changeInfo = output.ChangeInfo
		} else {
			input := &route53.DeactivateKeySigningKeyInput{
				HostedZoneId: aws.String(hostedZoneID),
				Name:         aws.String(name),
			}

			output, err := conn.DeactivateKeySigningKey(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deactivating Route 53 Key Signing Key (%s): %s", d.Id(), err)
			}

			changeInfo = output.ChangeInfo
		}

		if changeInfo != nil {
			if _, err := waitChangeInsync(ctx, conn, aws.ToString(changeInfo.Id)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) synchronize: %s", d.Id(), err)
			}
		}

		if _, err := waitKeySigningKeyStatusUpdated(ctx, conn, hostedZoneID, name, status); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeySigningKeyRead(ctx, d, meta)...)
}

func resourceKeySigningKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), keySigningKeyResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	hostedZoneID, name := parts[0], parts[1]

	if status := d.Get(names.AttrStatus).(string); status == keySigningKeyStatusActive || status == keySigningKeyStatusActionNeeded {
		input := &route53.DeactivateKeySigningKeyInput{
			HostedZoneId: aws.String(hostedZoneID),
			Name:         aws.String(name),
		}

		output, err := conn.DeactivateKeySigningKey(ctx, input)

		if errs.IsA[*awstypes.NoSuchKeySigningKey](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating Route 53 Key Signing Key (%s): %s", d.Id(), err)
		}

		if output.ChangeInfo != nil {
			if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) synchronize: %s", d.Id(), err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting Route 53 Key Signing Key: %s", d.Id())
	output, err := conn.DeleteKeySigningKey(ctx, &route53.DeleteKeySigningKeyInput{
		HostedZoneId: aws.String(hostedZoneID),
		Name:         aws.String(name),
	})

	if errs.IsA[*awstypes.NoSuchHostedZone](err) || errs.IsA[*awstypes.NoSuchKeySigningKey](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Key Signing Key (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) synchronize: %s", d.Id(), err)
		}
	}

	return diags
}

func findKeySigningKeyByTwoPartKey(ctx context.Context, conn *route53.Client, hostedZoneID, name string) (*awstypes.KeySigningKey, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.GetDNSSEC(ctx, input)

	if errs.IsA[*awstypes.NoSuchHostedZone](err) || errs.IsA[*awstypes.NoSuchKeySigningKey](err) {
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

	for _, v := range output.KeySigningKeys {
		if aws.ToString(v.Name) == name {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func statusKeySigningKey(ctx context.Context, conn *route53.Client, hostedZoneID, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findKeySigningKeyByTwoPartKey(ctx, conn, hostedZoneID, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitKeySigningKeyStatusUpdated(ctx context.Context, conn *route53.Client, hostedZoneID, name string, status string) (*awstypes.KeySigningKey, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Target:     []string{status},
		Refresh:    statusKeySigningKey(ctx, conn, hostedZoneID, name),
		MinTimeout: 5 * time.Second,
		Timeout:    timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KeySigningKey); ok {
		if status := aws.ToString(output.Status); status == keySigningKeyStatusInternalFailure {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
