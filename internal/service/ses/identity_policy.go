// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_identity_policy", name="Identity Policy")
func resourceIdentityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdentityPolicyCreate,
		ReadWithoutTimeout:   resourceIdentityPolicyRead,
		UpdateWithoutTimeout: resourceIdentityPolicyUpdate,
		DeleteWithoutTimeout: resourceIdentityPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters, dashes, and underscores"),
				),
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceIdentityPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	identity := d.Get("identity").(string)
	policyName := d.Get(names.AttrName).(string)
	id := identityPolicyCreateResourceID(identity, policyName)
	input := &ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		Policy:     aws.String(policy),
		PolicyName: aws.String(policyName),
	}

	_, err = conn.PutIdentityPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Identity Policy (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceIdentityPolicyRead(ctx, d, meta)...)
}

func resourceIdentityPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity, policyName, err := identityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := findIdentityPolicyByTwoPartKey(ctx, conn, identity, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Identity Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Policy (%s): %s", d.Id(), err)
	}

	d.Set("identity", identity)
	d.Set(names.AttrName, policyName)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceIdentityPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity, policyName, err := identityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		Policy:     aws.String(policy),
		PolicyName: aws.String(policyName),
	}

	_, err = conn.PutIdentityPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Identity Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIdentityPolicyRead(ctx, d, meta)...)
}

func resourceIdentityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	identity, policyName, err := identityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting SES Identity Policy: %s", d.Id())
	_, err = conn.DeleteIdentityPolicy(ctx, &ses.DeleteIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity Policy (%s): %s", d.Id(), err)
	}

	return diags
}

const identityPolicyResourceIDSeparator = "|"

func identityPolicyCreateResourceID(identity, policyName string) string {
	parts := []string{identity, policyName}
	id := strings.Join(parts, identityPolicyResourceIDSeparator)

	return id
}

func identityPolicyParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, identityPolicyResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected IDENTITY%[2]sNAME", id, identityPolicyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findIdentityPolicyByTwoPartKey(ctx context.Context, conn *ses.Client, identity, policyName string) (string, error) {
	input := &ses.GetIdentityPoliciesInput{
		Identity:    aws.String(identity),
		PolicyNames: []string{policyName},
	}
	output, err := findIdentityPolicies(ctx, conn, input)

	if err != nil {
		return "", err
	}

	v, ok := output[policyName]
	if !ok {
		return "", &retry.NotFoundError{}
	}

	return v, nil
}

func findIdentityPolicies(ctx context.Context, conn *ses.Client, input *ses.GetIdentityPoliciesInput) (map[string]string, error) {
	output, err := conn.GetIdentityPolicies(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policies == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policies, nil
}
