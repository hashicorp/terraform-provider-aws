package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsKmsGrant() *schema.Resource {
	return &schema.Resource{
		// There is no API for updating/modifying grants, hence no Update
		// Instead changes to most fields will force a new resource
		Create: resourceAwsKmsGrantCreate,
		Read:   resourceAwsKmsGrantRead,
		Delete: resourceAwsKmsGrantDelete,
		Exists: resourceAwsKmsGrantExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsKmsGrantName,
			},
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"grantee_principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"operations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateAwsKmsGrantOperation,
				},
				Required: true,
				ForceNew: true,
			},
			"constraints": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_context_equals": {
							Type:          schema.TypeMap,
							Optional:      true,
							Elem:          schema.TypeString,
							ConflictsWith: []string{"constraints.0.encryption_context_subset"},
						},
						"encryption_context_subset": {
							Type:          schema.TypeMap,
							Optional:      true,
							Elem:          schema.TypeString,
							ConflictsWith: []string{"constraints.0.encryption_context_equals"},
						},
					},
				},
			},
			"retiring_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"grant_creation_tokens": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"grant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grant_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsKmsGrantCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	input := kms.CreateGrantInput{
		GranteePrincipal: aws.String(d.Get("grantee_principal").(string)),
		KeyId:            aws.String(d.Get("key_id").(string)),
		Operations:       expandStringList(d.Get("operations").([]interface{})),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("constraints"); ok {
		input.Constraints = expandKmsGrantConstraints(v.([]interface{}))
	}
	if v, ok := d.GetOk("retiring_principal"); ok {
		input.RetiringPrincipal = aws.String(v.(string))
	}
	if v, ok := d.GetOk("grant_creation_tokens"); ok {
		input.GrantTokens = expandStringList(v.([]interface{}))
	}

	log.Printf("[DEBUG]: Adding new KMS Grant: %s", input)

	var out *kms.CreateGrantOutput

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error

		out, err = conn.CreateGrant(&input)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// Error Codes: https://docs.aws.amazon.com/sdk-for-go/api/service/kms/#KMS.CreateGrant
				// Under some circumstances a newly created IAM Role doesn't show up and causes
				// an InvalidArnException to be thrown. TODO: Possibly change the aws_iam_role code?
				if awsErr.Code() == "DependencyTimeoutException" ||
					awsErr.Code() == "InternalException" ||
					awsErr.Code() == "InvalidArnException" {
					return resource.RetryableError(
						fmt.Errorf("[WARN] Error adding new KMS Grant for key: %s, retrying %s",
							*input.KeyId, err))
				}
			}
			log.Printf("[ERROR] An error occured creating new AWS KMS Grant: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created new KMS Grant: %s", *out.GrantId)
	d.SetId(*out.GrantId)
	d.Set("grant_id", out.GrantId)
	d.Set("grant_token", out.GrantToken)

	return resourceAwsKmsGrantRead(d, meta)
}

func resourceAwsKmsGrantRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	grantId := d.Id()
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] Looking for grant id: %s", grantId)
	grant, err := findKmsGrantByIdWithRetry(conn, keyId, grantId)

	if err != nil {
		return err
	}
	if grant != nil {
		// The grant sometimes contains principals that identified by their unique id: "AROAJYCVIVUZIMTXXXXX"
		// instead of "arn:aws:...", in this case don't update the state file
		if strings.HasPrefix(*grant.GranteePrincipal, "arn:aws") {
			d.Set("grantee_principal", grant.GranteePrincipal)
		} else {
			log.Printf(
				"[WARN] Unable to update grantee principal state %s for grant id %s for key id %s.",
				*grant.GranteePrincipal, grantId, keyId)
		}

		if grant.RetiringPrincipal != nil {
			if strings.HasPrefix(*grant.RetiringPrincipal, "arn:aws") {
				d.Set("retiring_principal", grant.RetiringPrincipal)
			} else {
				log.Printf(
					"[WARN] Unable to update retiring principal state %s for grant id %s for key id %s",
					*grant.RetiringPrincipal, grantId, keyId)
			}
		}

		d.Set("operations", grant.Operations)
		if *grant.Name != "" {
			d.Set("name", grant.Name)
		}
		if grant.Constraints != nil {
			d.Set("constraints", flattenKmsGrantConstraints(grant.Constraints))
		}
	} else {
		log.Printf("[WARN] %s KMS grant id not found for key id %s, removing from state file", grantId, keyId)
		d.SetId("")
	}

	return nil
}

// Retiring grants requires special permissions (i.e. the
// caller to be root, retiree principal, or grantee principal with retire grant
// privileges). So just revoke grants.
func resourceAwsKmsGrantDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	grantId := d.Get("grant_id").(string)
	keyId := d.Get("key_id").(string)
	input := kms.RevokeGrantInput{
		GrantId: aws.String(grantId),
		KeyId:   aws.String(keyId),
	}

	log.Printf("[DEBUG] Revoking KMS grant: %s", grantId)
	_, err := conn.RevokeGrant(&input)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Checking if grant is revoked: %s", grantId)
	err = waitForKmsGrantToBeRevoked(conn, keyId, grantId)

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceAwsKmsGrantExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*AWSClient).kmsconn

	grantId := d.Id()
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] Looking for Grant: %s", grantId)
	grant, err := findKmsGrantByIdWithRetry(conn, keyId, grantId)

	if err != nil {
		if grant != nil {
			return true, err
		}
		return false, err
	}
	if grant != nil {
		return true, err
	}

	return false, err
}

func getKmsGrantById(grants []*kms.GrantListEntry, grantIdentifier string) *kms.GrantListEntry {
	for idx := range grants {
		if *grants[idx].GrantId == grantIdentifier {
			return grants[idx]
		}
	}

	return nil
}

/*
In the functions below it is not possible to use retryOnAwsCodes function, as there
is no describe grants call, so an error has to be created if the grant is or isn't returned
by the list grants call when expected.
*/

// NB: This function only retries the grant not being returned and some edge cases, while AWS Errors
// are handled by the findKmsGrantById function
func findKmsGrantByIdWithRetry(conn *kms.KMS, keyId string, grantId string) (*kms.GrantListEntry, error) {
	var grant *kms.GrantListEntry
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		grant, err = findKmsGrantById(conn, keyId, grantId, nil)

		if err != nil {
			if serr, ok := err.(KmsGrantMissingError); ok {
				// Force a retry if the grant should exist
				return resource.RetryableError(serr)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	return grant, err
}

// Used by the tests as well
func waitForKmsGrantToBeRevoked(conn *kms.KMS, keyId string, grantId string) error {
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		grant, err := findKmsGrantById(conn, keyId, grantId, nil)
		if err != nil {
			if _, ok := err.(KmsGrantMissingError); ok {
				if grant == nil {
					return nil
				}
			}
		}

		if grant != nil {
			// Force a retry if the grant still exists
			return resource.RetryableError(
				fmt.Errorf("[DEBUG] Grant still exists while expected to be revoked, retyring revocation check: %s", *grant.GrantId))
		}

		return resource.NonRetryableError(err)
	})

	return err
}

// The ListGrants API defaults to listing only 50 grants
// Use a marker to iterate over all grants in "pages"
// NB: This function only retries on AWS Errors
func findKmsGrantById(conn *kms.KMS, keyId string, grantId string, marker *string) (*kms.GrantListEntry, error) {

	input := kms.ListGrantsInput{
		KeyId:  aws.String(keyId),
		Limit:  aws.Int64(int64(100)),
		Marker: marker,
	}

	var out *kms.ListGrantsResponse
	var err error
	var grant *kms.GrantListEntry

	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		out, err = conn.ListGrants(&input)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NotFoundException" ||
					awsErr.Code() == "DependencyTimeoutException" ||
					awsErr.Code() == "InternalException" {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	grant = getKmsGrantById(out.Grants, grantId)
	if grant != nil {
		return grant, nil
	}
	if *out.Truncated {
		log.Printf("[DEBUG] KMS Grant list truncated, getting next page via marker: %s", *out.NextMarker)
		return findKmsGrantById(conn, keyId, grantId, out.NextMarker)
	}

	return nil, NewKmsGrantMissingError(fmt.Sprintf("[DEBUG] Grant %s not found for key id: %s", grantId, keyId))
}

func expandKmsGrantConstraints(configured []interface{}) *kms.GrantConstraints {
	if len(configured) < 1 {
		return nil
	}

	var constraint kms.GrantConstraints

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		if contextEq, ok := data["encryption_context_equals"]; ok {
			constraint.EncryptionContextEquals = stringMapToPointers(contextEq.(map[string]interface{}))
		}
		if contextSub, ok := data["encryption_context_subset"]; ok {
			constraint.SetEncryptionContextSubset(stringMapToPointers(contextSub.(map[string]interface{})))
		}
	}

	return &constraint
}

func flattenKmsGrantConstraints(constraint *kms.GrantConstraints) []interface{} {
	if constraint == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{}, 0)

	if constraint.EncryptionContextEquals != nil {
		m["encryption_context_equals"] = pointersMapToStringList(constraint.EncryptionContextEquals)
	}
	if constraint.EncryptionContextSubset != nil {
		m["encryption_context_subset"] = pointersMapToStringList(constraint.EncryptionContextSubset)
	}

	return []interface{}{m}
}

// Custom error, so we don't have to rely on
// the content of an error message
type KmsGrantMissingError struct {
	msg string
}

func (e KmsGrantMissingError) Error() string {
	return e.msg
}

func NewKmsGrantMissingError(msg string) KmsGrantMissingError {
	return KmsGrantMissingError{
		msg: msg,
	}
}
