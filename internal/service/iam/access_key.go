// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iam_access_key")
func ResourceAccessKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessKeyCreate,
		ReadWithoutTimeout:   resourceAccessKeyRead,
		UpdateWithoutTimeout: resourceAccessKeyUpdate,
		DeleteWithoutTimeout: resourceAccessKeyDelete,

		Importer: &schema.ResourceImporter{
			// ListAccessKeys requires UserName field in certain scenarios:
			//   ValidationError: Must specify userName when calling with non-User credentials
			// To prevent import from requiring this extra information, use GetAccessKeyLastUsed.
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).IAMConn(ctx)

				input := &iam.GetAccessKeyLastUsedInput{
					AccessKeyId: aws.String(d.Id()),
				}

				output, err := conn.GetAccessKeyLastUsedWithContext(ctx, input)

				if err != nil {
					return nil, fmt.Errorf("fetching IAM Access Key (%s) username via GetAccessKeyLastUsed: %w", d.Id(), err)
				}

				if output == nil || output.UserName == nil {
					return nil, fmt.Errorf("fetching IAM Access Key (%s) username via GetAccessKeyLastUsed: empty response", d.Id())
				}

				d.Set("user", output.UserName)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_ses_smtp_password_v4": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"ses_smtp_password_v4": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iam.StatusTypeActive,
				ValidateFunc: validation.StringInSlice(iam.StatusType_Values(), false),
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	username := d.Get("user").(string)

	request := &iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	}

	createResp, err := conn.CreateAccessKeyWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Access Key (%s): %s", username, err)
	}

	d.SetId(aws.StringValue(createResp.AccessKey.AccessKeyId))

	if createResp.AccessKey == nil || createResp.AccessKey.SecretAccessKey == nil {
		return sdkdiag.AppendErrorf(diags, "CreateAccessKey response did not contain a Secret Access Key as expected")
	}

	sesSMTPPasswordV4, err := SessmTPPasswordFromSecretKeySigV4(createResp.AccessKey.SecretAccessKey, meta.(*conns.AWSClient).Region)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting SES SigV4 SMTP Password from Secret Access Key: %s", err)
	}

	if v, ok := d.GetOk("pgp_key"); ok {
		pgpKey := v.(string)
		encryptionKey, err := retrieveGPGKey(pgpKey)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Access Key (%s): %s", username, err)
		}
		fingerprint, encrypted, err := encryptValue(encryptionKey, *createResp.AccessKey.SecretAccessKey, "IAM Access Key Secret")
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Access Key (%s): %s", username, err)
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret", encrypted)

		_, encrypted, err = encryptValue(encryptionKey, sesSMTPPasswordV4, "SES SMTP password")
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Access Key (%s): %s", username, err)
		}

		d.Set("encrypted_ses_smtp_password_v4", encrypted)
	} else {
		d.Set("secret", createResp.AccessKey.SecretAccessKey)

		d.Set("ses_smtp_password_v4", sesSMTPPasswordV4)
	}

	if v, ok := d.GetOk("status"); ok && v.(string) == iam.StatusTypeInactive {
		input := &iam.UpdateAccessKeyInput{
			AccessKeyId: aws.String(d.Id()),
			Status:      aws.String(iam.StatusTypeInactive),
			UserName:    aws.String(d.Get("user").(string)),
		}

		_, err := conn.UpdateAccessKeyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IAM Access Key (%s): %s", d.Id(), err)
		}

		createResp.AccessKey.Status = aws.String(iam.StatusTypeInactive)
	}

	resourceAccessKeyReadResult(d, &iam.AccessKeyMetadata{
		AccessKeyId: createResp.AccessKey.AccessKeyId,
		CreateDate:  createResp.AccessKey.CreateDate,
		Status:      createResp.AccessKey.Status,
		UserName:    createResp.AccessKey.UserName,
	})

	return diags
}

func resourceAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	username := d.Get("user").(string)

	key, err := FindAccessKey(ctx, conn, username, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Key (%s) for User (%s) not found, removing from state", d.Id(), username)
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Key (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.TimeValue(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set("status", key.Status)
	d.Set("user", key.UserName)

	return diags
}

func resourceAccessKeyReadResult(d *schema.ResourceData, key *iam.AccessKeyMetadata) {
	d.SetId(aws.StringValue(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.TimeValue(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set("status", key.Status)
	d.Set("user", key.UserName)
}

func resourceAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if d.HasChange("status") {
		if err := resourceAccessKeyStatusUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Access Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccessKeyRead(ctx, d, meta)...)
}

func resourceAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	request := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		UserName:    aws.String(d.Get("user").(string)),
	}

	if _, err := conn.DeleteAccessKeyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Access Key (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceAccessKeyStatusUpdate(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
	request := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		Status:      aws.String(d.Get("status").(string)),
		UserName:    aws.String(d.Get("user").(string)),
	}

	_, err := conn.UpdateAccessKeyWithContext(ctx, request)
	return err
}

func hmacSignature(key []byte, value []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(value); err != nil {
		return []byte(""), err
	}
	return h.Sum(nil), nil
}

func SessmTPPasswordFromSecretKeySigV4(key *string, region string) (string, error) {
	if key == nil {
		return "", nil
	}
	const version = byte(0x04)
	date := []byte("11111111")
	service := []byte("ses")
	terminal := []byte("aws4_request")
	message := []byte("SendRawEmail")

	rawSig, err := hmacSignature([]byte("AWS4"+*key), date)
	if err != nil {
		return "", err
	}

	if rawSig, err = hmacSignature(rawSig, []byte(region)); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, service); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, terminal); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, message); err != nil {
		return "", err
	}

	versionedSig := make([]byte, 0, len(rawSig)+1)
	versionedSig = append(versionedSig, version)
	versionedSig = append(versionedSig, rawSig...)
	return base64.StdEncoding.EncodeToString(versionedSig), nil
}
