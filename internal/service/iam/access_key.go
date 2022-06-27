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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAccessKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessKeyCreate,
		Read:   resourceAccessKeyRead,
		Update: resourceAccessKeyUpdate,
		Delete: resourceAccessKeyDelete,

		Importer: &schema.ResourceImporter{
			// ListAccessKeys requires UserName field in certain scenarios:
			//   ValidationError: Must specify userName when calling with non-User credentials
			// To prevent import from requiring this extra information, use GetAccessKeyLastUsed.
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).IAMConn

				input := &iam.GetAccessKeyLastUsedInput{
					AccessKeyId: aws.String(d.Id()),
				}

				output, err := conn.GetAccessKeyLastUsed(input)

				if err != nil {
					return nil, fmt.Errorf("error fetching IAM Access Key (%s) username via GetAccessKeyLastUsed: %w", d.Id(), err)
				}

				if output == nil || output.UserName == nil {
					return nil, fmt.Errorf("error fetching IAM Access Key (%s) username via GetAccessKeyLastUsed: empty response", d.Id())
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

func resourceAccessKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.CreateAccessKeyInput{
		UserName: aws.String(d.Get("user").(string)),
	}

	createResp, err := conn.CreateAccessKey(request)
	if err != nil {
		return fmt.Errorf(
			"Error creating access key for user %s: %s",
			*request.UserName,
			err,
		)
	}

	d.SetId(aws.StringValue(createResp.AccessKey.AccessKeyId))

	if createResp.AccessKey == nil || createResp.AccessKey.SecretAccessKey == nil {
		return fmt.Errorf("CreateAccessKey response did not contain a Secret Access Key as expected")
	}

	sesSMTPPasswordV4, err := SessmTPPasswordFromSecretKeySigV4(createResp.AccessKey.SecretAccessKey, meta.(*conns.AWSClient).Region)
	if err != nil {
		return fmt.Errorf("error getting SES SigV4 SMTP Password from Secret Access Key: %s", err)
	}

	if v, ok := d.GetOk("pgp_key"); ok {
		pgpKey := v.(string)
		encryptionKey, err := retrieveGPGKey(pgpKey)
		if err != nil {
			return err
		}
		fingerprint, encrypted, err := encryptValue(encryptionKey, *createResp.AccessKey.SecretAccessKey, "IAM Access Key Secret")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret", encrypted)

		_, encrypted, err = encryptValue(encryptionKey, sesSMTPPasswordV4, "SES SMTP password")
		if err != nil {
			return err
		}

		d.Set("encrypted_ses_smtp_password_v4", encrypted)
	} else {
		if err := d.Set("secret", createResp.AccessKey.SecretAccessKey); err != nil {
			return err
		}

		if err := d.Set("ses_smtp_password_v4", sesSMTPPasswordV4); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("status"); ok && v.(string) == iam.StatusTypeInactive {
		input := &iam.UpdateAccessKeyInput{
			AccessKeyId: aws.String(d.Id()),
			Status:      aws.String(iam.StatusTypeInactive),
			UserName:    aws.String(d.Get("user").(string)),
		}

		_, err := conn.UpdateAccessKey(input)

		if err != nil {
			return fmt.Errorf("error deactivating IAM Access Key (%s): %w", d.Id(), err)
		}

		createResp.AccessKey.Status = aws.String(iam.StatusTypeInactive)
	}

	return resourceAccessKeyReadResult(d, &iam.AccessKeyMetadata{
		AccessKeyId: createResp.AccessKey.AccessKeyId,
		CreateDate:  createResp.AccessKey.CreateDate,
		Status:      createResp.AccessKey.Status,
		UserName:    createResp.AccessKey.UserName,
	})
}

func resourceAccessKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	username := d.Get("user").(string)

	key, err := FindAccessKey(context.TODO(), conn, username, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Key (%s) for User (%s) not found, removing from state", d.Id(), username)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading IAM access key: %w", err)
	}

	d.SetId(aws.StringValue(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.TimeValue(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set("status", key.Status)
	d.Set("user", key.UserName)

	return nil
}

func resourceAccessKeyReadResult(d *schema.ResourceData, key *iam.AccessKeyMetadata) error {
	d.SetId(aws.StringValue(key.AccessKeyId))

	if key.CreateDate != nil {
		d.Set("create_date", aws.TimeValue(key.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}

	d.Set("status", key.Status)
	d.Set("user", key.UserName)

	return nil
}

func resourceAccessKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	if d.HasChange("status") {
		if err := resourceAccessKeyStatusUpdate(conn, d); err != nil {
			return err
		}
	}

	return resourceAccessKeyRead(d, meta)
}

func resourceAccessKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		UserName:    aws.String(d.Get("user").(string)),
	}

	if _, err := conn.DeleteAccessKey(request); err != nil {
		return fmt.Errorf("Error deleting access key %s: %s", d.Id(), err)
	}
	return nil
}

func resourceAccessKeyStatusUpdate(conn *iam.IAM, d *schema.ResourceData) error {
	request := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		Status:      aws.String(d.Get("status").(string)),
		UserName:    aws.String(d.Get("user").(string)),
	}

	if _, err := conn.UpdateAccessKey(request); err != nil {
		return fmt.Errorf("Error updating access key %s: %s", d.Id(), err)
	}
	return nil
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
