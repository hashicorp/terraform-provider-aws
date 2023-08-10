// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_route53_key_signing_key")
func ResourceKeySigningKey() *schema.Resource {
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
			"hosted_zone_id": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9._-]"), "must contain only alphanumeric characters, periods, underscores, or hyphens"),
				),
			},
			"public_key": {
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
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  KeySigningKeyStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					KeySigningKeyStatusActive,
					KeySigningKeyStatusInactive,
				}, false),
			},
		},
	}
}

func resourceKeySigningKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	hostedZoneID := d.Get("hosted_zone_id").(string)
	name := d.Get("name").(string)
	status := d.Get("status").(string)

	input := &route53.CreateKeySigningKeyInput{
		CallerReference: aws.String(id.UniqueId()),
		HostedZoneId:    aws.String(hostedZoneID),
		Name:            aws.String(name),
		Status:          aws.String(status),
	}

	if v, ok := d.GetOk("key_management_service_arn"); ok {
		input.KeyManagementServiceArn = aws.String(v.(string))
	}

	output, err := conn.CreateKeySigningKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Key Signing Key: %s", err)
	}

	d.SetId(KeySigningKeyCreateResourceID(hostedZoneID, name))

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) creation: %s", d.Id(), err)
		}
	}

	if _, err := waitKeySigningKeyStatusUpdated(ctx, conn, hostedZoneID, name, status); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status (%s): %s", d.Id(), status, err)
	}

	return append(diags, resourceKeySigningKeyRead(ctx, d, meta)...)
}

func resourceKeySigningKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	hostedZoneID, name, err := KeySigningKeyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Route 53 Key Signing Key (%s) identifier: %s", d.Id(), err)
	}

	keySigningKey, err := FindKeySigningKey(ctx, conn, hostedZoneID, name)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Key Signing Key (%s): %s", d.Id(), err)
	}

	if keySigningKey == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Route 53 Key Signing Key (%s): not found", d.Id())
		}

		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("digest_algorithm_mnemonic", keySigningKey.DigestAlgorithmMnemonic)
	d.Set("digest_algorithm_type", keySigningKey.DigestAlgorithmType)
	d.Set("digest_value", keySigningKey.DigestValue)
	d.Set("dnskey_record", keySigningKey.DNSKEYRecord)
	d.Set("ds_record", keySigningKey.DSRecord)
	d.Set("flag", keySigningKey.Flag)
	d.Set("hosted_zone_id", hostedZoneID)
	d.Set("key_management_service_arn", keySigningKey.KmsArn)
	d.Set("key_tag", keySigningKey.KeyTag)
	d.Set("name", keySigningKey.Name)
	d.Set("public_key", keySigningKey.PublicKey)
	d.Set("signing_algorithm_mnemonic", keySigningKey.SigningAlgorithmMnemonic)
	d.Set("signing_algorithm_type", keySigningKey.SigningAlgorithmType)
	d.Set("status", keySigningKey.Status)

	return diags
}

func resourceKeySigningKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	if d.HasChange("status") {
		status := d.Get("status").(string)

		switch status {
		default:
			return sdkdiag.AppendErrorf(diags, "updating Route 53 Key Signing Key (%s) status: unknown status (%s)", d.Id(), status)
		case KeySigningKeyStatusActive:
			input := &route53.ActivateKeySigningKeyInput{
				HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
				Name:         aws.String(d.Get("name").(string)),
			}

			output, err := conn.ActivateKeySigningKeyWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Route 53 Key Signing Key (%s) status (%s): %s", d.Id(), status, err)
			}

			if output != nil && output.ChangeInfo != nil {
				if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status (%s) update: %s", d.Id(), status, err)
				}
			}
		case KeySigningKeyStatusInactive:
			input := &route53.DeactivateKeySigningKeyInput{
				HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
				Name:         aws.String(d.Get("name").(string)),
			}

			output, err := conn.DeactivateKeySigningKeyWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Route 53 Key Signing Key (%s) status (%s): %s", d.Id(), status, err)
			}

			if output != nil && output.ChangeInfo != nil {
				if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status (%s) update: %s", d.Id(), status, err)
				}
			}
		}

		if _, err := waitKeySigningKeyStatusUpdated(ctx, conn, d.Get("hosted_zone_id").(string), d.Get("name").(string), status); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status (%s): %s", d.Id(), status, err)
		}
	}

	return append(diags, resourceKeySigningKeyRead(ctx, d, meta)...)
}

func resourceKeySigningKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	status := d.Get("status").(string)

	if status == KeySigningKeyStatusActive || status == KeySigningKeyStatusActionNeeded {
		input := &route53.DeactivateKeySigningKeyInput{
			HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
			Name:         aws.String(d.Get("name").(string)),
		}

		output, err := conn.DeactivateKeySigningKeyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route 53 Key Signing Key (%s) status (%s): %s", d.Id(), status, err)
		}

		if output != nil && output.ChangeInfo != nil {
			if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) status (%s) update: %s", d.Id(), status, err)
			}
		}
	}

	input := &route53.DeleteKeySigningKeyInput{
		HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	output, err := conn.DeleteKeySigningKeyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Key Signing Key (%s), status (%s): %s", d.Id(), d.Get("status").(string), err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Key Signing Key (%s) deletion: %s", d.Id(), err)
		}
	}

	return diags
}
