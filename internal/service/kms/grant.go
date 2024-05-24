// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_grant", name="Grant")
func resourceGrant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGrantCreate,
		ReadWithoutTimeout:   resourceGrantRead,
		DeleteWithoutTimeout: resourceGrantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyID, grantID, err := grantParseResourceID(d.Id())
				if err != nil {
					return nil, err
				}

				d.SetId(grantCreateResourceID(keyID, grantID))
				d.Set(names.AttrKeyID, keyID)
				d.Set("grant_id", grantID)

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
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
			names.AttrKeyID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.GrantOperation](),
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
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID := d.Get(names.AttrKeyID).(string)
	input := &kms.CreateGrantInput{
		GranteePrincipal: aws.String(d.Get("grantee_principal").(string)),
		KeyId:            aws.String(keyID),
		Operations:       flex.ExpandStringyValueSet[awstypes.GrantOperation](d.Get("operations").(*schema.Set)),
	}

	if v, ok := d.GetOk("constraints"); ok && v.(*schema.Set).Len() > 0 {
		if !grantConstraintsIsValid(v.(*schema.Set)) {
			return sdkdiag.AppendErrorf(diags, "A grant constraint can't have both encryption_context_equals and encryption_context_subset set")
		}

		input.Constraints = expandGrantConstraints(v.(*schema.Set))
	}

	if v, ok := d.GetOk("grant_creation_tokens"); ok && v.(*schema.Set).Len() > 0 {
		input.GrantTokens = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retiring_principal"); ok {
		input.RetiringPrincipal = aws.String(v.(string))
	}

	// Error Codes: https://docs.aws.amazon.com/sdk-for-go/api/service/kms/#KMS.CreateGrant
	// Under some circumstances a newly created IAM Role doesn't show up and causes
	// an InvalidArnException to be thrown.
	outputRaw, err := tfresource.RetryWhenIsOneOf3[*awstypes.DependencyTimeoutException, *awstypes.KMSInternalException, *awstypes.InvalidArnException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateGrant(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Grant for Key (%s): %s", keyID, err)
	}

	output := outputRaw.(*kms.CreateGrantOutput)
	grantID := aws.ToString(output.GrantId)
	d.SetId(grantCreateResourceID(keyID, grantID))
	d.Set("grant_token", output.GrantToken)

	return append(diags, resourceGrantRead(ctx, d, meta)...)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID, grantID, err := grantParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	grant, err := findGrantByTwoPartKeyWithRetry(ctx, conn, keyID, grantID, propagationTimeout)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Grant (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Grant (%s): %s", d.Id(), err)
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
	d.Set(names.AttrKeyID, keyID)
	if aws.ToString(grant.Name) != "" {
		d.Set(names.AttrName, grant.Name)
	}
	d.Set("operations", grant.Operations)
	if grant.RetiringPrincipal != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("retiring_principal", grant.RetiringPrincipal)
	}

	return diags
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID, grantID, err := grantParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.Get("retire_on_delete").(bool) {
		log.Printf("[DEBUG] Retiring KMS Grant: %s", d.Id())
		_, err = conn.RetireGrant(ctx, &kms.RetireGrantInput{
			GrantId: aws.String(grantID),
			KeyId:   aws.String(keyID),
		})
	} else {
		log.Printf("[DEBUG] Revoking KMS Grant: %s", d.Id())
		_, err = conn.RevokeGrant(ctx, &kms.RevokeGrantInput{
			GrantId: aws.String(grantID),
			KeyId:   aws.String(keyID),
		})
	}

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Grant (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findGrantByTwoPartKey(ctx, conn, keyID, grantID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for KMS Grant (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGrant(ctx context.Context, conn *kms.Client, input *kms.ListGrantsInput, filter tfslices.Predicate[*awstypes.GrantListEntry]) (*awstypes.GrantListEntry, error) {
	output, err := findGrants(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGrants(ctx context.Context, conn *kms.Client, input *kms.ListGrantsInput, filter tfslices.Predicate[*awstypes.GrantListEntry]) ([]awstypes.GrantListEntry, error) {
	var output []awstypes.GrantListEntry

	pages := kms.NewListGrantsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Grants {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findGrantByTwoPartKey(ctx context.Context, conn *kms.Client, keyID, grantID string) (*awstypes.GrantListEntry, error) {
	input := &kms.ListGrantsInput{
		KeyId: aws.String(keyID),
		Limit: aws.Int32(100),
	}

	return findGrant(ctx, conn, input, func(v *awstypes.GrantListEntry) bool {
		return aws.ToString(v.GrantId) == grantID
	})
}

func findGrantByTwoPartKeyWithRetry(ctx context.Context, conn *kms.Client, keyID, grantID string, timeout time.Duration) (*awstypes.GrantListEntry, error) {
	var output *awstypes.GrantListEntry

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		grant, err := findGrantByTwoPartKey(ctx, conn, keyID, grantID)

		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if principal := aws.ToString(grant.GranteePrincipal); principal != "" {
			if !arn.IsARN(principal) && !verify.IsServicePrincipal(principal) {
				return retry.RetryableError(fmt.Errorf("grantee principal (%s) is invalid. Perhaps the principal has been deleted or recreated", principal))
			}
		}

		if principal := aws.ToString(grant.RetiringPrincipal); principal != "" {
			if !arn.IsARN(principal) && !verify.IsServicePrincipal(principal) {
				return retry.RetryableError(fmt.Errorf("retiring principal (%s) is invalid. Perhaps the principal has been deleted or recreated", principal))
			}
		}

		output = grant

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = findGrantByTwoPartKey(ctx, conn, keyID, grantID)
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

func expandGrantConstraints(tfSet *schema.Set) *awstypes.GrantConstraints {
	if len(tfSet.List()) < 1 {
		return nil
	}

	apiObject := &awstypes.GrantConstraints{}

	for _, tfMapRaw := range tfSet.List() {
		tfMap := tfMapRaw.(map[string]interface{})

		if v, ok := tfMap["encryption_context_equals"]; ok {
			apiObject.EncryptionContextEquals = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}

		if v, ok := tfMap["encryption_context_subset"]; ok {
			apiObject.EncryptionContextSubset = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}
	}

	return apiObject
}

func flattenGrantConstraints(apiObject *awstypes.GrantConstraints) *schema.Set {
	tfSet := schema.NewSet(resourceGrantConstraintsHash, []interface{}{})

	if apiObject == nil {
		return tfSet
	}

	tfMap := make(map[string]interface{})

	if len(apiObject.EncryptionContextEquals) > 0 {
		tfMap["encryption_context_equals"] = flex.FlattenStringValueMap(apiObject.EncryptionContextEquals)
	}

	if len(apiObject.EncryptionContextSubset) > 0 {
		tfMap["encryption_context_subset"] = flex.FlattenStringValueMap(apiObject.EncryptionContextSubset)
	}

	tfSet.Add(tfMap)

	return tfSet
}

// NB: For the constraint hash to be deterministic the order in which
// print the keys and values of the encryption context maps needs to be
// determistic, so sort them.
func sortedConcatStringMap(m map[string]string) string {
	keys := tfmaps.Keys(m)
	slices.Sort(keys)

	var elems []string
	for _, key := range keys {
		elems = append(elems, key, m[key])
	}

	return strings.Join(elems, "-")
}

// The hash needs to encapsulate what type of constraint it is
// as well as the keys and values of the constraint.
func resourceGrantConstraintsHash(v interface{}) int {
	var buf bytes.Buffer

	tfMap, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	if v, ok := tfMap["encryption_context_equals"]; ok {
		if len(v.(map[string]interface{})) > 0 {
			buf.WriteString(fmt.Sprintf("encryption_context_equals-%s-", sortedConcatStringMap(flex.ExpandStringValueMap(v.(map[string]interface{})))))
		}
	}
	if v, ok := tfMap["encryption_context_subset"]; ok {
		if len(v.(map[string]interface{})) > 0 {
			buf.WriteString(fmt.Sprintf("encryption_context_subset-%s-", sortedConcatStringMap(flex.ExpandStringValueMap(v.(map[string]interface{})))))
		}
	}

	return create.StringHashcode(buf.String())
}

const grantResourceIDSeparator = ":"

func grantCreateResourceID(keyID, grantID string) string {
	parts := []string{keyID, grantID}
	id := strings.Join(parts, grantResourceIDSeparator)

	return id
}

func grantParseResourceID(id string) (string, string, error) {
	if arn.IsARN(id) {
		arnParts := strings.Split(id, "/")
		if len(arnParts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ARN (%[1]s), expected KeyID%[2]sGrantID", id, grantResourceIDSeparator)
		}
		arnPrefix := arnParts[0]
		parts := strings.Split(arnParts[1], grantResourceIDSeparator)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected KeyID%[2]sGrantID", id, grantResourceIDSeparator)
		}
		return fmt.Sprintf("%s/%s", arnPrefix, parts[0]), parts[1], nil
	} else {
		parts := strings.Split(id, grantResourceIDSeparator)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected KeyID%[2]sGrantID", id, grantResourceIDSeparator)
		}
		return parts[0], parts[1], nil
	}
}
