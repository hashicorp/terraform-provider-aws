package kms

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGrant() *schema.Resource {
	return &schema.Resource{
		// There is no API for updating/modifying grants, hence no Update
		// Instead changes to most fields will force a new resource
		CreateWithoutTimeout: resourceGrantCreate,
		ReadWithoutTimeout:   resourceGrantRead,
		DeleteWithoutTimeout: resourceGrantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyId, grantId, err := decodeGrantID(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("key_id", keyId)
				d.Set("grant_id", grantId)
				d.SetId(fmt.Sprintf("%s:%s", keyId, grantId))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validGrantName,
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
				ValidateFunc: verify.ValidARN,
			},
			"operations": {
				Type: schema.TypeSet,
				Set:  schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(kms.GrantOperation_Values(), false),
				},
				Required: true,
				ForceNew: true,
			},
			"constraints": {
				Type:     schema.TypeSet,
				Set:      resourceGrantConstraintsHash,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_context_equals": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// ConflictsWith encryption_context_subset handled in Create, see grantConstraintsIsValid
						},
						"encryption_context_subset": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// ConflictsWith encryption_context_equals handled in Create, see grantConstraintsIsValid
						},
					},
				},
			},
			"retiring_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"grant_creation_tokens": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"retire_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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

func resourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()
	keyId := d.Get("key_id").(string)

	input := kms.CreateGrantInput{
		GranteePrincipal: aws.String(d.Get("grantee_principal").(string)),
		KeyId:            aws.String(keyId),
		Operations:       flex.ExpandStringSet(d.Get("operations").(*schema.Set)),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("constraints"); ok {
		if !grantConstraintsIsValid(v.(*schema.Set)) {
			return sdkdiag.AppendErrorf(diags, "A grant constraint can't have both encryption_context_equals and encryption_context_subset set")
		}
		input.Constraints = expandGrantConstraints(v.(*schema.Set))
	}
	if v, ok := d.GetOk("retiring_principal"); ok {
		input.RetiringPrincipal = aws.String(v.(string))
	}
	if v, ok := d.GetOk("grant_creation_tokens"); ok {
		input.GrantTokens = flex.ExpandStringSet(v.(*schema.Set))
	}

	var out *kms.CreateGrantOutput

	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.CreateGrantWithContext(ctx, &input)

		if err != nil {
			// Error Codes: https://docs.aws.amazon.com/sdk-for-go/api/service/kms/#KMS.CreateGrant
			// Under some circumstances a newly created IAM Role doesn't show up and causes
			// an InvalidArnException to be thrown.
			if tfawserr.ErrCodeEquals(err, kms.ErrCodeDependencyTimeoutException, kms.ErrCodeInternalException, kms.ErrCodeInvalidArnException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.CreateGrantWithContext(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Grant for Key (%s): %s", keyId, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", keyId, aws.StringValue(out.GrantId)))
	d.Set("grant_id", out.GrantId)
	d.Set("grant_token", out.GrantToken)

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	keyId, grantId, err := decodeGrantID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Grant (%s): %s", d.Id(), err)
	}

	grant, err := findGrantByIdRetry(ctx, conn, keyId, grantId)

	if err != nil {
		if tfresource.NotFound(err) {
			log.Printf("[WARN] KMS Grant (%s) not found for Key (%s), removing from state", grantId, keyId)
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading KMS Grant (%s): %s", d.Id(), err)
	}

	if grant == nil {
		log.Printf("[WARN] KMS Grant (%s) not found for Key (%s), removing from state", grantId, keyId)
		d.SetId("")
		return diags
	}

	if grant.GranteePrincipal != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("grantee_principal", grant.GranteePrincipal)
	}
	if grant.RetiringPrincipal != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("retiring_principal", grant.RetiringPrincipal)
	}

	if err := d.Set("operations", aws.StringValueSlice(grant.Operations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting operations: %s", err)
	}
	if aws.StringValue(grant.Name) != "" {
		d.Set("name", grant.Name)
	}
	if grant.Constraints != nil {
		if err := d.Set("constraints", flattenGrantConstraints(grant.Constraints)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting constraints: %s", err)
		}
	}

	return diags
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	keyId, grantId, err := decodeGrantID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Grant (%s): %s", d.Id(), err)
	}
	doRetire := d.Get("retire_on_delete").(bool)

	if doRetire {
		retireInput := kms.RetireGrantInput{
			GrantId: aws.String(grantId),
			KeyId:   aws.String(keyId),
		}

		log.Printf("[DEBUG] Retiring KMS grant: %s", grantId)
		_, err = conn.RetireGrantWithContext(ctx, &retireInput)
	} else {
		revokeInput := kms.RevokeGrantInput{
			GrantId: aws.String(grantId),
			KeyId:   aws.String(keyId),
		}

		log.Printf("[DEBUG] Revoking KMS grant: %s", grantId)
		_, err = conn.RevokeGrantWithContext(ctx, &revokeInput)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting KMS Grant (%s): %s", d.Id(), err)
	}

	if err := WaitForGrantToBeRevoked(ctx, conn, keyId, grantId); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Grant (%s): waiting for completion: %s", d.Id(), err)
	}
	return diags
}

func getGrantByID(grants []*kms.GrantListEntry, grantIdentifier string) *kms.GrantListEntry {
	for _, grant := range grants {
		if aws.StringValue(grant.GrantId) == grantIdentifier {
			return grant
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
// are handled by the findGrantByID function
func findGrantByIdRetry(ctx context.Context, conn *kms.KMS, keyId string, grantId string) (*kms.GrantListEntry, error) {
	var grant *kms.GrantListEntry
	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		var err error
		grant, err = findGrantByID(ctx, conn, keyId, grantId, nil)

		if tfresource.NotFound(err) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if principal := aws.StringValue(grant.GranteePrincipal); principal != "" {
			if !arn.IsARN(principal) {
				return resource.RetryableError(fmt.Errorf("grantee principal is invalid: %q", principal))
			}
		}

		if principal := aws.StringValue(grant.RetiringPrincipal); principal != "" {
			if !arn.IsARN(principal) {
				return resource.RetryableError(fmt.Errorf("retiring principal is invalid: %q", principal))
			}
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		grant, err = findGrantByID(ctx, conn, keyId, grantId, nil)
	}

	return grant, err
}

// Used by the tests as well
func WaitForGrantToBeRevoked(ctx context.Context, conn *kms.KMS, keyId string, grantId string) error {
	var grant *kms.GrantListEntry
	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		var err error
		grant, err = findGrantByID(ctx, conn, keyId, grantId, nil)

		if tfresource.NotFound(err) {
			return nil
		}

		if grant != nil {
			// Force a retry if the grant still exists
			return resource.RetryableError(
				fmt.Errorf("Grant still exists while expected to be revoked, retyring revocation check: %s", *grant.GrantId))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		grant, err = findGrantByID(ctx, conn, keyId, grantId, nil)
	}

	return err
}

// The ListGrants API defaults to listing only 50 grants
// Use a marker to iterate over all grants in "pages"
// NB: This function only retries on AWS Errors
func findGrantByID(ctx context.Context, conn *kms.KMS, keyId string, grantId string, marker *string) (*kms.GrantListEntry, error) {
	input := kms.ListGrantsInput{
		KeyId:  aws.String(keyId),
		Limit:  aws.Int64(100),
		Marker: marker,
	}

	var out *kms.ListGrantsResponse
	var err error
	var grant *kms.GrantListEntry

	err = resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		out, err = conn.ListGrantsWithContext(ctx, &input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, kms.ErrCodeDependencyTimeoutException) ||
				tfawserr.ErrCodeEquals(err, kms.ErrCodeInternalException) ||
				tfawserr.ErrCodeEquals(err, kms.ErrCodeInvalidArnException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.ListGrantsWithContext(ctx, &input)
	}

	if err != nil {
		return nil, fmt.Errorf("listing KMS Grants: %w", err)
	}

	grant = getGrantByID(out.Grants, grantId)
	if grant != nil {
		return grant, nil
	}
	if aws.BoolValue(out.Truncated) {
		log.Printf("[DEBUG] KMS Grant list truncated, getting next page via marker: %s", aws.StringValue(out.NextMarker))
		return findGrantByID(ctx, conn, keyId, grantId, out.NextMarker)
	}

	return nil, &resource.NotFoundError{
		Message: fmt.Sprintf("grant %s not found for key id: %s", grantId, keyId),
	}
}

// Can't have both constraint options set:
// ValidationException: More than one constraint supplied
// NB: set.List() returns an empty map if the constraint is not set, filter those out
// using len(v) > 0
func grantConstraintsIsValid(constraints *schema.Set) bool {
	constraintCount := 0
	for _, raw := range constraints.List() {
		data := raw.(map[string]interface{})
		if v, ok := data["encryption_context_equals"].(map[string]interface{}); ok {
			if len(v) > 0 {
				constraintCount += 1
			}
		}
		if v, ok := data["encryption_context_subset"].(map[string]interface{}); ok {
			if len(v) > 0 {
				constraintCount += 1
			}
		}
	}

	return constraintCount <= 1
}

func expandGrantConstraints(configured *schema.Set) *kms.GrantConstraints {
	if len(configured.List()) < 1 {
		return nil
	}

	var constraint kms.GrantConstraints

	for _, raw := range configured.List() {
		data := raw.(map[string]interface{})
		if contextEq, ok := data["encryption_context_equals"]; ok {
			constraint.SetEncryptionContextEquals(flex.ExpandStringMap(contextEq.(map[string]interface{})))
		}
		if contextSub, ok := data["encryption_context_subset"]; ok {
			constraint.SetEncryptionContextSubset(flex.ExpandStringMap(contextSub.(map[string]interface{})))
		}
	}

	return &constraint
}

func sortStringMapKeys(m map[string]*string) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// NB: For the constraint hash to be deterministic the order in which
// print the keys and values of the encryption context maps needs to be
// determistic, so sort them.
func sortedConcatStringMap(m map[string]*string) string {
	var strList []string
	mapKeys := sortStringMapKeys(m)
	for _, key := range mapKeys {
		strList = append(strList, key, *m[key])
	}
	return strings.Join(strList, "-")
}

// The hash needs to encapsulate what type of constraint it is
// as well as the keys and values of the constraint.
func resourceGrantConstraintsHash(v interface{}) int {
	var buf bytes.Buffer
	m, castOk := v.(map[string]interface{})
	if !castOk {
		return 0
	}

	if v, ok := m["encryption_context_equals"]; ok {
		if len(v.(map[string]interface{})) > 0 {
			buf.WriteString(fmt.Sprintf("encryption_context_equals-%s-", sortedConcatStringMap(flex.ExpandStringMap(v.(map[string]interface{})))))
		}
	}
	if v, ok := m["encryption_context_subset"]; ok {
		if len(v.(map[string]interface{})) > 0 {
			buf.WriteString(fmt.Sprintf("encryption_context_subset-%s-", sortedConcatStringMap(flex.ExpandStringMap(v.(map[string]interface{})))))
		}
	}

	return create.StringHashcode(buf.String())
}

func flattenGrantConstraints(constraint *kms.GrantConstraints) *schema.Set {
	constraints := schema.NewSet(resourceGrantConstraintsHash, []interface{}{})
	if constraint == nil {
		return constraints
	}

	m := make(map[string]interface{})
	if constraint.EncryptionContextEquals != nil {
		if len(constraint.EncryptionContextEquals) > 0 {
			m["encryption_context_equals"] = flex.PointersMapToStringList(constraint.EncryptionContextEquals)
		}
	}
	if constraint.EncryptionContextSubset != nil {
		if len(constraint.EncryptionContextSubset) > 0 {
			m["encryption_context_subset"] = flex.PointersMapToStringList(constraint.EncryptionContextSubset)
		}
	}
	constraints.Add(m)

	return constraints
}

func decodeGrantID(id string) (string, string, error) {
	if arn.IsARN(id) {
		arnParts := strings.Split(id, "/")
		if len(arnParts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ARN (%q), expected KeyID:GrantID", id)
		}
		arnPrefix := arnParts[0]
		parts := strings.Split(arnParts[1], ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%q), expected KeyID:GrantID", id)
		}
		return fmt.Sprintf("%s/%s", arnPrefix, parts[0]), parts[1], nil
	} else {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%q), expected KeyID:GrantID", id)
		}
		return parts[0], parts[1], nil
	}
}
