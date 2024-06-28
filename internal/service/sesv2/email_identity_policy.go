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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_email_identity_policy", name="Email Identity Policy")
func ResourceEmailIdentityPolicy() *schema.Resource {
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
				StateFunc: func(v interface{}) string {
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
	ResNameEmailIdentityPolicy = "Email Identity Policy"
)

func resourceEmailIdentityPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	email_identity := d.Get("email_identity").(string)
	policy_name := d.Get("policy_name").(string)
	emailIdentityPolicyID := FormatEmailIdentityPolicyID(email_identity, policy_name)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	in := &sesv2.CreateEmailIdentityPolicyInput{
		EmailIdentity: aws.String(email_identity),
		Policy:        aws.String(policy),
		PolicyName:    aws.String(policy_name),
	}

	out, err := conn.CreateEmailIdentityPolicy(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameEmailIdentityPolicy, emailIdentityPolicyID, err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameEmailIdentityPolicy, emailIdentityPolicyID, errors.New("empty output"))
	}

	d.SetId(emailIdentityPolicyID)

	return append(diags, resourceEmailIdentityPolicyRead(ctx, d, meta)...)
}

func resourceEmailIdentityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	emailIdentity, policyName, err := ParseEmailIdentityPolicyID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	policy, err := FindEmailIdentityPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentityPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	policy, err = verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	policy, err = structure.NormalizeJsonString(policy)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	d.Set("email_identity", emailIdentity)
	d.Set(names.AttrPolicy, policy)
	d.Set("policy_name", policyName)

	return diags
}

func resourceEmailIdentityPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	in := &sesv2.UpdateEmailIdentityPolicyInput{
		EmailIdentity: aws.String(d.Get("email_identity").(string)),
		Policy:        aws.String(policy),
		PolicyName:    aws.String(d.Get("policy_name").(string)),
	}

	log.Printf("[DEBUG] Updating SESV2 EmailIdentityPolicy (%s): %#v", d.Id(), in)
	_, err = conn.UpdateEmailIdentityPolicy(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	return append(diags, resourceEmailIdentityPolicyRead(ctx, d, meta)...)
}

func resourceEmailIdentityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 EmailIdentityPolicy %s", d.Id())

	emailIdentity, policyName, err := ParseEmailIdentityPolicyID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	_, err = conn.DeleteEmailIdentityPolicy(ctx, &sesv2.DeleteEmailIdentityPolicyInput{
		EmailIdentity: aws.String(emailIdentity),
		PolicyName:    aws.String(policyName),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameEmailIdentityPolicy, d.Id(), err)
	}

	return diags
}

func FindEmailIdentityPolicyByID(ctx context.Context, conn *sesv2.Client, id string) (string, error) {
	emailIdentity, policyName, err := ParseEmailIdentityPolicyID(id)
	if err != nil {
		return "", err
	}

	in := &sesv2.GetEmailIdentityPoliciesInput{
		EmailIdentity: aws.String(emailIdentity),
	}

	out, err := conn.GetEmailIdentityPolicies(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return "", &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return "", err
	}

	if out == nil {
		return "", tfresource.NewEmptyResultError(in)
	}

	for name, policy := range out.Policies {
		if policyName == name {
			return policy, nil
		}
	}

	return "", &retry.NotFoundError{}
}

func FormatEmailIdentityPolicyID(emailIdentity, policyName string) string {
	return fmt.Sprintf("%s|%s", emailIdentity, policyName)
}

func ParseEmailIdentityPolicyID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", errors.New("please make sure the ID is in the form EMAIL_IDENTITY|POLICY_NAME")
	}

	return idParts[0], idParts[1], nil
}
