// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"log"
	"math/big"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_user_login_profile", name="User Login Profile")
func resourceUserLoginProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserLoginProfileCreate,
		ReadWithoutTimeout:   resourceUserLoginProfileRead,
		DeleteWithoutTimeout: resourceUserLoginProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("encrypted_password", "")
				d.Set("key_fingerprint", "")
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"password_reset_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"password_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(5, 128),
			},

			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_password": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPassword: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	charLower   = "abcdefghijklmnopqrstuvwxyz"
	charUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charNumbers = "0123456789"
	charSymbols = "!@#$%^&*()_+-=[]{}|'"
)

// GeneratePassword generates a random password of a given length, matching the
// most restrictive iam password policy.
func GeneratePassword(length int) (string, error) {
	const charset = charLower + charUpper + charNumbers + charSymbols

	result := make([]byte, length)
	charsetSize := big.NewInt(int64(len(charset)))

	// rather than trying to artificially add specific characters from each
	// class to the password to match the policy, we generate passwords
	// randomly and reject those that don't match.
	//
	// Even in the worst case, this tends to take less than 10 tries to find a
	// matching password. Any sufficiently long password is likely to succeed
	// on the first try
	for n := 0; n < 100000; n++ {
		for i := range result {
			r, err := rand.Int(rand.Reader, charsetSize)
			if err != nil {
				return "", err
			}
			if !r.IsInt64() {
				return "", errors.New("rand.Int() not representable as an Int64")
			}

			result[i] = charset[r.Int64()]
		}

		if !CheckPwdPolicy(result) {
			continue
		}

		return string(result), nil
	}

	return "", errors.New("failed to generate acceptable password")
}

// Check the generated password contains all character classes listed in the
// IAM password policy.
func CheckPwdPolicy(pass []byte) bool {
	return (bytes.ContainsAny(pass, charLower) &&
		bytes.ContainsAny(pass, charNumbers) &&
		bytes.ContainsAny(pass, charSymbols) &&
		bytes.ContainsAny(pass, charUpper))
}

func resourceUserLoginProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)
	username := d.Get("user").(string)

	passwordLength := d.Get("password_length").(int)
	initialPassword, err := GeneratePassword(passwordLength)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM User Login Profile for %q: %s", username, err)
	}

	request := &iam.CreateLoginProfileInput{
		UserName:              aws.String(username),
		Password:              aws.String(initialPassword),
		PasswordResetRequired: d.Get("password_reset_required").(bool),
	}

	createResp, err := conn.CreateLoginProfile(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM User Login Profile for %q: %s", username, err)
	}

	d.SetId(aws.ToString(createResp.LoginProfile.UserName))

	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := retrieveGPGKey(v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM User Login Profile for %q: %s", username, err)
		}

		fingerprint, encrypted, err := encryptValue(encryptionKey, initialPassword, "Password")
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM User Login Profile for %q: %s", username, err)
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_password", encrypted)
	} else {
		d.Set(names.AttrPassword, initialPassword)
	}

	return append(diags, resourceUserLoginProfileRead(ctx, d, meta)...)
}

func resourceUserLoginProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.GetLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	var output *iam.GetLoginProfileOutput

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		output, err = conn.GetLoginProfile(ctx, input)

		if d.IsNewResource() && errs.IsA[*awstypes.NoSuchEntityException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetLoginProfile(ctx, input)
	}

	if !d.IsNewResource() && errs.IsA[*awstypes.NoSuchEntityException](err) {
		log.Printf("[WARN] IAM User Login Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Login Profile (%s): %s", d.Id(), err)
	}

	if output == nil || output.LoginProfile == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Login Profile (%s): empty response", d.Id())
	}

	d.Set("user", output.LoginProfile.UserName)
	return diags
}

func resourceUserLoginProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting IAM User Login Profile (%s): %v", d.Id(), input)
	// Handle IAM eventual consistency
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.DeleteLoginProfile(ctx, input)

		var nse *awstypes.NoSuchEntityException
		if tfawserr.ErrCodeEquals(err, nse.ErrorCode()) {
			return nil
		}

		// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
		var etu *awstypes.EntityTemporarilyUnmodifiableException
		if tfawserr.ErrCodeEquals(err, etu.ErrorCode()) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	// Handle AWS Go SDK automatic retries
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteLoginProfile(ctx, input)
	}

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Login Profile (%s): %s", d.Id(), err)
	}

	return diags
}
