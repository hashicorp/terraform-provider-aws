// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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

	instanceArn := d.Get("instance_arn").(string)
	managedPolicyArn := d.Get("managed_policy_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		ManagedPolicyArn: aws.String(managedPolicyArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := conn.AttachManagedPolicyToPermissionSetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching Managed Policy to SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s", managedPolicyArn, permissionSetArn, instanceArn))

	// Provision ALL accounts after attaching the managed policy
	if err := provisionPermissionSet(ctx, conn, permissionSetArn, instanceArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	return append(diags, resourceManagedPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceManagedPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	managedPolicyArn, permissionSetArn, instanceArn, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSO Managed Policy Attachment ID: %s", err)
	}

	policy, err := FindManagedPolicy(ctx, conn, managedPolicyArn, permissionSetArn, instanceArn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", managedPolicyArn, permissionSetArn)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Managed Policy (%s) for SSO Permission Set (%s): %s", managedPolicyArn, permissionSetArn, err)
	}

	if policy == nil {
		log.Printf("[WARN] Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", managedPolicyArn, permissionSetArn)
		d.SetId("")
		return diags
	}

	d.Set("instance_arn", instanceArn)
	d.Set("managed_policy_arn", policy.Arn)
	d.Set("managed_policy_name", policy.Name)
	d.Set("permission_set_arn", permissionSetArn)

	return diags
}

func resourceManagedPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	managedPolicyArn, permissionSetArn, instanceArn, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSO Managed Policy Attachment ID: %s", err)
	}

	input := &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		ManagedPolicyArn: aws.String(managedPolicyArn),
	}

	_, err = conn.DetachManagedPolicyFromPermissionSetWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "detaching Managed Policy (%s) from SSO Permission Set (%s): %s", managedPolicyArn, permissionSetArn, err)
	}

	// Provision ALL accounts after detaching the managed policy
	if err := provisionPermissionSet(ctx, conn, permissionSetArn, instanceArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", permissionSetArn, err)
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
