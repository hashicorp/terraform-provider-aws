package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	rolePolicyNameMaxLen       = 128
	rolePolicyNamePrefixMaxLen = rolePolicyNameMaxLen - resource.UniqueIDSuffixLength
)

func ResourceRolePolicy() *schema.Resource {
	return &schema.Resource{
		// PutRolePolicy API is idempotent, so these can be the same.
		CreateWithoutTimeout: resourceRolePolicyPut,
		UpdateWithoutTimeout: resourceRolePolicyPut,

		ReadWithoutTimeout:   resourceRolePolicyRead,
		DeleteWithoutTimeout: resourceRolePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := verify.LegacyPolicyNormalize(v)
					return json
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validRolePolicyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(rolePolicyNamePrefixMaxLen),
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validRolePolicyRole,
			},
		},
	}
}

func resourceRolePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	policy, err := verify.LegacyPolicyNormalize(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	request := &iam.PutRolePolicyInput{
		RoleName:       aws.String(d.Get("role").(string)),
		PolicyDocument: aws.String(policy),
	}

	var policyName string
	if v, ok := d.GetOk("name"); ok {
		policyName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		policyName = resource.PrefixedUniqueId(v.(string))
	} else {
		policyName = resource.UniqueId()
	}
	request.PolicyName = aws.String(policyName)

	if _, err := conn.PutRolePolicyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM role policy %s: %s", *request.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.RoleName, *request.PolicyName))
	return append(diags, resourceRolePolicyRead(ctx, d, meta)...)
}

func resourceRolePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	request := &iam.GetRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	var getResp *iam.GetRolePolicyOutput

	err = resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = conn.GetRolePolicyWithContext(ctx, request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetRolePolicyWithContext(ctx, request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Role Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get("policy").(string), policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	d.Set("name", name)
	d.Set("role", role)

	return diags
}

func resourceRolePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM role policy (%s): %s", d.Id(), err)
	}

	request := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	if _, err := conn.DeleteRolePolicyWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM role policy (%s): %s", d.Id(), err)
	}
	return diags
}

func RolePolicyParseID(id string) (roleName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("role_policy id must be of the form <role name>:<policy name>")
		return
	}

	roleName = parts[0]
	policyName = parts[1]
	return
}
