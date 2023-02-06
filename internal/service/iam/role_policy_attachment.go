package iam

import (
	"context"
	"fmt"
	"log"
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
)

func ResourceRolePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceRolePolicyAttachmentRead,
		DeleteWithoutTimeout: resourceRolePolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRolePolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRolePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	role := d.Get("role").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToRole(ctx, conn, role, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching policy %s to IAM Role %s: %v", arn, role, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", role)))

	return append(diags, resourceRolePolicyAttachmentRead(ctx, d, meta)...)
}

func resourceRolePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	role := d.Get("role").(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s:%s", role, policyARN)

	var hasPolicyAttachment bool

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		hasPolicyAttachment, err = RoleHasPolicyARNAttachment(ctx, conn, role, policyARN)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && !hasPolicyAttachment {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("IAM Role Managed Policy Attachment (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		hasPolicyAttachment, err = RoleHasPolicyARNAttachment(ctx, conn, role, policyARN)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Role Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Managed Policy Attachment (%s): %s", id, err)
	}

	if !d.IsNewResource() && !hasPolicyAttachment {
		log.Printf("[WARN] IAM Role Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	d.Set("policy_arn", policyARN)
	d.Set("role", role)

	return diags
}

func resourceRolePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	role := d.Get("role").(string)
	arn := d.Get("policy_arn").(string)

	err := DetachPolicyFromRole(ctx, conn, role, arn)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing policy %s from IAM Role %s: %v", arn, role, err)
	}
	return diags
}

func resourceRolePolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <role-name>/<policy_arn>", d.Id())
	}

	roleName := idParts[0]
	policyARN := idParts[1]

	d.Set("role", roleName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", roleName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToRole(ctx context.Context, conn *iam.IAM, role string, arn string) error {
	_, err := conn.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(role),
		PolicyArn: aws.String(arn),
	})
	return err
}

func DetachPolicyFromRole(ctx context.Context, conn *iam.IAM, role string, arn string) error {
	_, err := conn.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(role),
		PolicyArn: aws.String(arn),
	})
	return err
}

func RoleHasPolicyARNAttachment(ctx context.Context, conn *iam.IAM, role string, policyARN string) (bool, error) {
	hasPolicyAttachment := false
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(role),
	}

	err := conn.ListAttachedRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		for _, p := range page.AttachedPolicies {
			if aws.StringValue(p.PolicyArn) == policyARN {
				hasPolicyAttachment = true
				return false
			}
		}

		return !lastPage
	})

	return hasPolicyAttachment, err
}
