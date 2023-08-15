// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_kms_grant")
func ResourceGrant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGrantCreate,
		ReadWithoutTimeout:   resourceGrantRead,
		DeleteWithoutTimeout: resourceGrantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyId, grantId, err := GrantParseResourceID(d.Id())
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
			"constraints": {
				Type:     schema.TypeSet,
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
				Set: resourceGrantConstraintsHash,
			},

			"grant_creation_tokens": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"grant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grant_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grantee_principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidARN,
					verify.ValidServicePrincipal,
				),
			},
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validGrantName,
			},
			"operations": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(kms.GrantOperation_Values(), false),
				},
			},
			"retire_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"retiring_principal": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidARN,
					verify.ValidServicePrincipal,
				),
			},
		},
	}
}

func resourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	keyID := d.Get("key_id").(string)
	input := &kms.CreateGrantInput{
		GranteePrincipal: aws.String(d.Get("grantee_principal").(string)),
		KeyId:            aws.String(keyID),
		Operations:       flex.ExpandStringSet(d.Get("operations").(*schema.Set)),
	}

	if v, ok := d.GetOk("constraints"); ok && v.(*schema.Set).Len() > 0 {
		if !grantConstraintsIsValid(v.(*schema.Set)) {
			return sdkdiag.AppendErrorf(diags, "A grant constraint can't have both encryption_context_equals and encryption_context_subset set")
		}

		input.Constraints = expandGrantConstraints(v.(*schema.Set))
	}

	if v, ok := d.GetOk("grant_creation_tokens"); ok && v.(*schema.Set).Len() > 0 {
		input.GrantTokens = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retiring_principal"); ok {
		input.RetiringPrincipal = aws.String(v.(string))
	}

	// Error Codes: https://docs.aws.amazon.com/sdk-for-go/api/service/kms/#KMS.CreateGrant
	// Under some circumstances a newly created IAM Role doesn't show up and causes
	// an InvalidArnException to be thrown.
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 3*time.Minute, func() (interface{}, error) {
		return conn.CreateGrantWithContext(ctx, input)
	}, kms.ErrCodeDependencyTimeoutException, kms.ErrCodeInternalException, kms.ErrCodeInvalidArnException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Grant for Key (%s): %s", keyID, err)
	}

	output := outputRaw.(*kms.CreateGrantOutput)
	grantID := aws.StringValue(output.GrantId)
	GrantCreateResourceID(keyID, grantID)
	d.SetId(GrantCreateResourceID(keyID, grantID))
	d.Set("grant_token", output.GrantToken)

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		timeout = 3 * time.Minute
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	keyID, grantID, err := GrantParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	grant, err := findGrantByTwoPartKeyWithRetry(ctx, conn, keyID, grantID, timeout)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Grant (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading KMS Grant (%s): %s", d.Id(), err)
	}

	if grant.Constraints != nil {
		if err := d.Set("constraints", flattenGrantConstraints(grant.Constraints)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting constraints: %s", err)
		}
	}
	d.Set("grant_id", grant.GrantId)
	if grant.GranteePrincipal != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("grantee_principal", grant.GranteePrincipal)
	}
	d.Set("key_id", keyID)
	if aws.StringValue(grant.Name) != "" {
		d.Set("name", grant.Name)
	}
	d.Set("operations", aws.StringValueSlice(grant.Operations))
	if grant.RetiringPrincipal != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("retiring_principal", grant.RetiringPrincipal)
	}

	return diags
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	keyID, grantID, err := GrantParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	if d.Get("retire_on_delete").(bool) {
		log.Printf("[DEBUG] Retiring KMS Grant: %s", d.Id())
		_, err = conn.RetireGrantWithContext(ctx, &kms.RetireGrantInput{
			GrantId: aws.String(grantID),
			KeyId:   aws.String(keyID),
		})
	} else {
		log.Printf("[DEBUG] Revoking KMS Grant: %s", d.Id())
		_, err = conn.RevokeGrantWithContext(ctx, &kms.RevokeGrantInput{
			GrantId: aws.String(grantID),
			KeyId:   aws.String(keyID),
		})
	}

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Grant (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 3*time.Minute, func() (interface{}, error) {
		return FindGrantByTwoPartKey(ctx, conn, keyID, grantID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Grant (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindGrantByTwoPartKey(ctx context.Context, conn *kms.KMS, keyID, grantID string) (*kms.GrantListEntry, error) {
	input := &kms.ListGrantsInput{
		KeyId: aws.String(keyID),
		Limit: aws.Int64(100),
	}
	var output *kms.GrantListEntry

	err := conn.ListGrantsPagesWithContext(ctx, input, func(page *kms.ListGrantsResponse, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Grants {
			if v == nil {
				continue
			}

			if aws.StringValue(v.GrantId) == grantID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
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

	return output, nil
}

func findGrantByTwoPartKeyWithRetry(ctx context.Context, conn *kms.KMS, keyID, grantID string, timeout time.Duration) (*kms.GrantListEntry, error) {
	var output *kms.GrantListEntry

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		grant, err := FindGrantByTwoPartKey(ctx, conn, keyID, grantID)

		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if principal := aws.StringValue(grant.GranteePrincipal); principal != "" {
			if !arn.IsARN(principal) && !verify.IsServicePrincipal(principal) {
				return retry.RetryableError(fmt.Errorf("grantee principal (%s) is invalid. Perhaps the principal has been deleted or recreated", principal))
			}
		}

		if principal := aws.StringValue(grant.RetiringPrincipal); principal != "" {
			if !arn.IsARN(principal) && !verify.IsServicePrincipal(principal) {
				return retry.RetryableError(fmt.Errorf("retiring principal (%s) is invalid. Perhaps the principal has been deleted or recreated", principal))
			}
		}

		output = grant

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = FindGrantByTwoPartKey(ctx, conn, keyID, grantID)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
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

const grantIDSeparator = ":"

func GrantCreateResourceID(keyID, grantID string) string {
	parts := []string{keyID, grantID}
	id := strings.Join(parts, grantIDSeparator)

	return id
}

func GrantParseResourceID(id string) (string, string, error) {
	if arn.IsARN(id) {
		arnParts := strings.Split(id, "/")
		if len(arnParts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ARN (%q), expected KeyID:GrantID", id)
		}
		arnPrefix := arnParts[0]
		parts := strings.Split(arnParts[1], grantIDSeparator)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%q), expected KeyID:GrantID", id)
		}
		return fmt.Sprintf("%s/%s", arnPrefix, parts[0]), parts[1], nil
	} else {
		parts := strings.Split(id, grantIDSeparator)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%q), expected KeyID:GrantID", id)
		}
		return parts[0], parts[1], nil
	}
}
