// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_identity_policy")
func ResourceIdentityPolicy() *schema.Resource {
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
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceIdentityPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity := d.Get("identity").(string)
	policyName := d.Get(names.AttrName).(string)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	input := &ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(policy),
	}

	_, err = conn.PutIdentityPolicyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Identity (%s) Policy: %s", identity, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, policyName))

	return append(diags, resourceIdentityPolicyRead(ctx, d, meta)...)
}

func resourceIdentityPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Identity Policy (%s): %s", d.Id(), err)
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	req := ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(policy),
	}

	_, err = conn.PutIdentityPolicyWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return append(diags, resourceIdentityPolicyRead(ctx, d, meta)...)
}

func resourceIdentityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Policy (%s): %s", d.Id(), err)
	}

	input := &ses.GetIdentityPoliciesInput{
		Identity:    aws.String(identity),
		PolicyNames: aws.StringSlice([]string{policyName}),
	}

	output, err := conn.GetIdentityPoliciesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting SES Identity (%s) Policy (%s): empty result", identity, policyName)
	}

	if len(output.Policies) == 0 {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return diags
	}

	policy, ok := output.Policies[policyName]
	if !ok {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return diags
	}

	d.Set("identity", identity)
	d.Set(names.AttrName, policyName)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.StringValue(policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Identity Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceIdentityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity Policy (%s): %s", d.Id(), err)
	}

	input := &ses.DeleteIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
	}

	log.Printf("[DEBUG] Deleting SES Identity Policy: %s", input)
	_, err = conn.DeleteIdentityPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return diags
}

func IdentityPolicyParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "|", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected IDENTITY|NAME", id)
	}
	return idParts[0], idParts[1], nil
}
