// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_email_identity_policy", name="Email Identity Policy")
func resourceEmailIdentityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityPolicyCreate,
		ReadWithoutTimeout:   resourceEmailIdentityPolicyRead,
		UpdateWithoutTimeout: resourceEmailIdentityPolicyUpdate,
		DeleteWithoutTimeout: resourceEmailIdentityPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"policy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

const (
	resNameEmailIdentityPolicy = "Email Identity Policy"
)

func resourceEmailIdentityPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	email_identity := d.Get("email_identity").(string)
	policy_name := d.Get("policy_name").(string)
	emailIdentityPolicyID := emailIdentityPolicyCreateResourceID(email_identity, policy_name)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := &sesv2.CreateEmailIdentityPolicyInput{
		EmailIdentity: aws.String(email_identity),
		Policy:        aws.String(policy),
		PolicyName:    aws.String(policy_name),
	}

	out, err := conn.CreateEmailIdentityPolicy(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameEmailIdentityPolicy, emailIdentityPolicyID, err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameEmailIdentityPolicy, emailIdentityPolicyID, errors.New("empty output"))
	}

	d.SetId(emailIdentityPolicyID)

	return append(diags, resourceEmailIdentityPolicyRead(ctx, d, meta)...)
}

func resourceEmailIdentityPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	emailIdentity, policyName, err := emailIdentityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findEmailIdentityPolicyByTwoPartKey(ctx, conn, emailIdentity, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentityPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameEmailIdentityPolicy, d.Id(), err)
	}

	policy, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(out))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err = structure.NormalizeJsonString(policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("email_identity", emailIdentity)
	d.Set(names.AttrPolicy, policy)
	d.Set("policy_name", policyName)

	return diags
}

func resourceEmailIdentityPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	emailIdentity, policyName, err := emailIdentityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := &sesv2.UpdateEmailIdentityPolicyInput{
		EmailIdentity: aws.String(emailIdentity),
		Policy:        aws.String(policy),
		PolicyName:    aws.String(policyName),
	}

	_, err = conn.UpdateEmailIdentityPolicy(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameEmailIdentityPolicy, d.Id(), err)
	}

	return append(diags, resourceEmailIdentityPolicyRead(ctx, d, meta)...)
}

func resourceEmailIdentityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	emailIdentity, policyName, err := emailIdentityPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting SESV2 EmailIdentityPolicy: %s", d.Id())
	_, err = conn.DeleteEmailIdentityPolicy(ctx, &sesv2.DeleteEmailIdentityPolicyInput{
		EmailIdentity: aws.String(emailIdentity),
		PolicyName:    aws.String(policyName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameEmailIdentityPolicy, d.Id(), err)
	}

	return diags
}

const emailIdentityPolicyResourceIDSeparator = "|"

func emailIdentityPolicyCreateResourceID(emailIdentity, policyName string) string {
	parts := []string{emailIdentity, policyName}
	id := strings.Join(parts, emailIdentityPolicyResourceIDSeparator)

	return id
}

func emailIdentityPolicyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, emailIdentityPolicyResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected EMAIL_IDENTITY%[2]sPOLICY_NAME", id, emailIdentityPolicyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findEmailIdentityPolicyByTwoPartKey(ctx context.Context, conn *sesv2.Client, emailIdentity, policyName string) (*string, error) {
	input := &sesv2.GetEmailIdentityPoliciesInput{
		EmailIdentity: aws.String(emailIdentity),
	}
	output, err := findEmailIdentityPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output, ok := output[policyName]; ok {
		return aws.String(output), nil
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findEmailIdentityPolicies(ctx context.Context, conn *sesv2.Client, input *sesv2.GetEmailIdentityPoliciesInput) (map[string]string, error) {
	output, err := conn.GetEmailIdentityPolicies(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policies == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policies, nil
}
