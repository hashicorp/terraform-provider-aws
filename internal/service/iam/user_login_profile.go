package iam

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserLoginProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserLoginProfileCreate,
		Read:   resourceUserLoginProfileRead,
		Delete: resourceUserLoginProfileDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
			"password": {
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

func resourceUserLoginProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	username := d.Get("user").(string)

	passwordLength := d.Get("password_length").(int)
	initialPassword, err := GeneratePassword(passwordLength)
	if err != nil {
		return err
	}

	request := &iam.CreateLoginProfileInput{
		UserName:              aws.String(username),
		Password:              aws.String(initialPassword),
		PasswordResetRequired: aws.Bool(d.Get("password_reset_required").(bool)),
	}

	log.Println("[DEBUG] Create IAM User Login Profile request:", request)
	createResp, err := conn.CreateLoginProfile(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM User Login Profile for %q: %w", username, err)
	}

	d.SetId(aws.StringValue(createResp.LoginProfile.UserName))

	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := retrieveGPGKey(v.(string))
		if err != nil {
			return fmt.Errorf("error retrieving GPG Key during IAM User Login Profile (%s) creation: %w", username, err)
		}

		fingerprint, encrypted, err := encryptValue(encryptionKey, initialPassword, "Password")
		if err != nil {
			return fmt.Errorf("error encrypting password during IAM User Login Profile (%s) creation: %w", username, err)
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_password", encrypted)
	} else {
		d.Set("password", initialPassword)
	}

	return nil
}

func resourceUserLoginProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.GetLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	var output *iam.GetLoginProfileOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.GetLoginProfile(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetLoginProfile(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Login Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM User Login Profile (%s): %w", d.Id(), err)
	}

	if output == nil || output.LoginProfile == nil {
		return fmt.Errorf("error reading IAM User Login Profile (%s): empty response", d.Id())
	}

	loginProfile := output.LoginProfile

	d.Set("user", loginProfile.UserName)
	d.Set("password_reset_required", loginProfile.PasswordResetRequired)

	return nil
}

func resourceUserLoginProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting IAM User Login Profile (%s): %s", d.Id(), input)
	// Handle IAM eventual consistency
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteLoginProfile(input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}

		// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeEntityTemporarilyUnmodifiableException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Handle AWS Go SDK automatic retries
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteLoginProfile(input)
	}

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IAM User Login Profile (%s): %w", d.Id(), err)
	}

	return nil
}
