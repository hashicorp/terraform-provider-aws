// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ssoadmin_managed_policy_attachment")
func ResourceManagedPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceManagedPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceManagedPolicyAttachmentRead,
		DeleteWithoutTimeout: resourceManagedPolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"managed_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"managed_policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceManagedPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	instanceARN := d.Get("instance_arn").(string)
	managedPolicyARN := d.Get("managed_policy_arn").(string)
	permissionSetARN := d.Get("permission_set_arn").(string)

	// Check for duplicates.
	_, err := FindManagedPolicy(ctx, conn, managedPolicyARN, permissionSetARN, instanceARN)

	if err == nil {
		return sdkdiag.AppendErrorf(diags, "attaching Managed Policy (%s) to SSO Permission Set (%s): already attached", managedPolicyARN, permissionSetARN)
	} else if !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading SSO Managed Policy (%s) Attachment (%s): %s", managedPolicyARN, permissionSetARN, err)
	}

	input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		ManagedPolicyArn: aws.String(managedPolicyARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	_, err = conn.AttachManagedPolicyToPermissionSetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching Managed Policy (%s) to SSO Permission Set (%s): %s", managedPolicyARN, permissionSetARN, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s", managedPolicyARN, permissionSetARN, instanceARN))

	// Provision ALL accounts after attaching the managed policy.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceManagedPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceManagedPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	managedPolicyARN, permissionSetARN, instanceARN, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := FindManagedPolicy(ctx, conn, managedPolicyARN, permissionSetARN, instanceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Managed Policy Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Managed Policy Attachment (%s): %s", d.Id(), err)
	}

	d.Set("instance_arn", instanceARN)
	d.Set("managed_policy_arn", policy.Arn)
	d.Set("managed_policy_name", policy.Name)
	d.Set("permission_set_arn", permissionSetARN)

	return diags
}

func resourceManagedPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	managedPolicyARN, permissionSetARN, instanceARN, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		ManagedPolicyArn: aws.String(managedPolicyARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	_, err = conn.DetachManagedPolicyFromPermissionSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "detaching Managed Policy (%s) from SSO Permission Set (%s): %s", managedPolicyARN, permissionSetARN, err)
	}

	// Provision ALL accounts after detaching the managed policy.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func ParseManagedPolicyAttachmentID(id string) (string, string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("parsing ID: expected MANAGED_POLICY_ARN,PERMISSION_SET_ARN,INSTANCE_ARN")
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func FindManagedPolicy(ctx context.Context, conn *ssoadmin.SSOAdmin, managedPolicyARN, permissionSetARN, instanceARN string) (*ssoadmin.AttachedManagedPolicy, error) {
	input := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}
	filter := func(a *ssoadmin.AttachedManagedPolicy) bool {
		return aws.StringValue(a.Arn) == managedPolicyARN
	}

	return findAttachedManagedPolicy(ctx, conn, input, filter)
}

func findAttachedManagedPolicy(ctx context.Context, conn *ssoadmin.SSOAdmin, input *ssoadmin.ListManagedPoliciesInPermissionSetInput, filter tfslices.Predicate[*ssoadmin.AttachedManagedPolicy]) (*ssoadmin.AttachedManagedPolicy, error) {
	output, err := findAttachedManagedPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAttachedManagedPolicies(ctx context.Context, conn *ssoadmin.SSOAdmin, input *ssoadmin.ListManagedPoliciesInPermissionSetInput, filter tfslices.Predicate[*ssoadmin.AttachedManagedPolicy]) ([]*ssoadmin.AttachedManagedPolicy, error) {
	var output []*ssoadmin.AttachedManagedPolicy

	err := conn.ListManagedPoliciesInPermissionSetPagesWithContext(ctx, input, func(page *ssoadmin.ListManagedPoliciesInPermissionSetOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AttachedManagedPolicies {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
