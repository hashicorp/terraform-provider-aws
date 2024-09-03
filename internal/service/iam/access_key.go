// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_access_key", name="Access Key")
func resourceAccessKey() *schema.Resource {
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
				conn := meta.(*conns.AWSClient).IAMClient(ctx)

				input := &iam.GetAccessKeyLastUsedInput{
					AccessKeyId: aws.String(d.Id()),
				}

				output, err := conn.GetAccessKeyLastUsed(ctx, input)

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
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.StatusTypeActive,
				ValidateDiagFunc: enum.Validate[awstypes.StatusType](),
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	username := d.Get("user").(string)

	request := &iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	}

	createResp, err := conn.CreateAccessKey(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Access Key (%s): %s", username, err)
	}

	d.SetId(aws.ToString(createResp.AccessKey.AccessKeyId))

	if createResp.AccessKey == nil || createResp.AccessKey.SecretAccessKey == nil {
		return sdkdiag.AppendErrorf(diags, "CreateAccessKey response did not contain a Secret Access Key as expected")
	}

	sesSMTPPasswordV4, err := sesSMTPPasswordFromSecretKeySigV4(createResp.AccessKey.SecretAccessKey, meta.(*conns.AWSClient).Region)
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

	if v, ok := d.GetOk(names.AttrStatus); ok && v.(string) == string(awstypes.StatusTypeInactive) {
		input := &iam.UpdateAccessKeyInput{
			AccessKeyId: aws.String(d.Id()),
			Status:      awstypes.StatusTypeInactive,
			UserName:    aws.String(d.Get("user").(string)),
		}

		_, err := conn.UpdateAccessKey(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IAM Access Key (%s): %s", d.Id(), err)
		}

		createResp.AccessKey.Status = awstypes.StatusTypeInactive
	}

	resourceAccessKeyReadResult(d, &awstypes.AccessKeyMetadata{
		AccessKeyId: createResp.AccessKey.AccessKeyId,
		CreateDate:  createResp.AccessKey.CreateDate,
		Status:      createResp.AccessKey.Status,
		UserName:    createResp.AccessKey.UserName,
	})

	return diags
}

func resourceAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	username := d.Get("user").(string)
	key, err := findAccessKeyByTwoPartKey(ctx, conn, username, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Key (%s) for User (%s) not found, removing from state", d.Id(), username)
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Key (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.ToTime(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set(names.AttrStatus, key.Status)
	d.Set("user", key.UserName)

	return diags
}

func resourceAccessKeyReadResult(d *schema.ResourceData, key *awstypes.AccessKeyMetadata) {
	d.SetId(aws.ToString(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.ToTime(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set(names.AttrStatus, key.Status)
	d.Set("user", key.UserName)
}

func resourceAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChange(names.AttrStatus) {
		if err := resourceAccessKeyStatusUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Access Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccessKeyRead(ctx, d, meta)...)
}

func resourceAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Deleting IAM Access Key: %s", d.Id())
	_, err := conn.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		UserName:    aws.String(d.Get("user").(string)),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Access Key (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAccessKeyStatusUpdate(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	request := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		Status:      awstypes.StatusType(d.Get(names.AttrStatus).(string)),
		UserName:    aws.String(d.Get("user").(string)),
	}

	_, err := conn.UpdateAccessKey(ctx, request)
	return err
}

func findAccessKeyByTwoPartKey(ctx context.Context, conn *iam.Client, username, id string) (*awstypes.AccessKeyMetadata, error) {
	input := &iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}

	return findAccessKey(ctx, conn, input, func(v awstypes.AccessKeyMetadata) bool {
		return aws.ToString(v.AccessKeyId) == id
	})
}

func findAccessKeysByUser(ctx context.Context, conn *iam.Client, username string) ([]awstypes.AccessKeyMetadata, error) {
	input := &iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}

	return findAccessKeys(ctx, conn, input, tfslices.PredicateTrue[awstypes.AccessKeyMetadata]())
}

func findAccessKey(ctx context.Context, conn *iam.Client, input *iam.ListAccessKeysInput, filter tfslices.Predicate[awstypes.AccessKeyMetadata]) (*awstypes.AccessKeyMetadata, error) {
	output, err := findAccessKeys(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccessKeys(ctx context.Context, conn *iam.Client, input *iam.ListAccessKeysInput, filter tfslices.Predicate[awstypes.AccessKeyMetadata]) ([]awstypes.AccessKeyMetadata, error) {
	var output []awstypes.AccessKeyMetadata

	pages := iam.NewListAccessKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AccessKeyMetadata {
			if !reflect.ValueOf(v).IsZero() && filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func hmacSignature(key []byte, value []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(value); err != nil {
		return []byte(""), err
	}
	return h.Sum(nil), nil
}

func sesSMTPPasswordFromSecretKeySigV4(key *string, region string) (string, error) {
	if key == nil {
		return "", nil
	}
	const version = byte(0x04)
	date := []byte("11111111")
	service := []byte("ses")
	terminal := []byte("aws4_request")
	message := []byte("SendRawEmail")

	rawSig, err := hmacSignature([]byte("AWS4"+aws.ToString(key)), date)
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
	return itypes.Base64Encode(versionedSig), nil
}
