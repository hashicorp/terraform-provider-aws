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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	permissionsBoundaryAttachmentTimeout = 5 * time.Minute
)

// @SDKResource("aws_ssoadmin_permissions_boundary_attachment")
func ResourcePermissionsBoundaryAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionsBoundaryAttachmentCreate,
		ReadWithoutTimeout:   resourcePermissionsBoundaryAttachmentRead,
		DeleteWithoutTimeout: resourcePermissionsBoundaryAttachmentDelete,

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
			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"permissions_boundary": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customer_managed_policy_reference": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"path": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "/",
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 512),
									},
								},
							},
						},
						"managed_policy_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 2048),
						},
					},
				},
			},
		},
	}
}

func resourcePermissionsBoundaryAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	tfMap := d.Get("permissions_boundary").([]interface{})[0].(map[string]interface{})
	instanceARN := d.Get("instance_arn").(string)
	permissionSetARN := d.Get("permission_set_arn").(string)
	id := PermissionsBoundaryAttachmentCreateResourceID(permissionSetARN, instanceARN)
	input := &ssoadmin.PutPermissionsBoundaryToPermissionSetInput{
		PermissionsBoundary: expandPermissionsBoundary(tfMap),
		InstanceArn:         aws.String(instanceARN),
		PermissionSetArn:    aws.String(permissionSetARN),
	}

	log.Printf("[INFO] Attaching permissions boundary to permission set: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, permissionsBoundaryAttachmentTimeout, func() (interface{}, error) {
		return conn.PutPermissionsBoundaryToPermissionSetWithContext(ctx, input)
	}, ssoadmin.ErrCodeConflictException, ssoadmin.ErrCodeThrottlingException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Permissions Boundary Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	// After the policy has been attached to the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", permissionSetARN, err)
	}

	return append(diags, resourcePermissionsBoundaryAttachmentRead(ctx, d, meta)...)
}

func resourcePermissionsBoundaryAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permissions Boundary Attachment (%s): %s", d.Id(), err)
	}

	policy, err := FindPermissionsBoundary(ctx, conn, permissionSetARN, instanceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Permissions Boundary Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permissions Boundary Attachment (%s): %s", d.Id(), err)
	}

	if err := d.Set("permissions_boundary", []interface{}{flattenPermissionsBoundary(policy)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions_boundary: %s", err)
	}
	d.Set("instance_arn", instanceARN)
	d.Set("permission_set_arn", permissionSetARN)

	return diags
}

func resourcePermissionsBoundaryAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Permissions Boundary Attachment (%s): %s", d.Id(), err)
	}

	input := &ssoadmin.DeletePermissionsBoundaryFromPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	log.Printf("[INFO] Detaching permissions boundary from permission set: %s", input)
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, permissionsBoundaryAttachmentTimeout, func() (interface{}, error) {
		return conn.DeletePermissionsBoundaryFromPermissionSetWithContext(ctx, input)
	}, ssoadmin.ErrCodeConflictException, ssoadmin.ErrCodeThrottlingException)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Permissions Boundary Attachment (%s): %s", d.Id(), err)
	}

	// After the policy has been detached from the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", permissionSetARN, err)
	}

	return diags
}

const permissionsBoundaryAttachmentIDSeparator = ","

func PermissionsBoundaryAttachmentCreateResourceID(permissionSetARN, instanceARN string) string {
	parts := []string{permissionSetARN, instanceARN}
	id := strings.Join(parts, permissionsBoundaryAttachmentIDSeparator)

	return id
}

func PermissionsBoundaryAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, permissionsBoundaryAttachmentIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected PERMISSION_SET_ARN%[2]sINSTANCE_ARN", id, permissionsBoundaryAttachmentIDSeparator)
}

func expandPermissionsBoundary(tfMap map[string]interface{}) *ssoadmin.PermissionsBoundary {
	if tfMap == nil {
		return nil
	}

	apiObject := &ssoadmin.PermissionsBoundary{}

	if v, ok := tfMap["customer_managed_policy_reference"].([]interface{}); ok && len(v) > 0 {
		if cmpr, ok := v[0].(map[string]interface{}); ok {
			apiObject.CustomerManagedPolicyReference = expandCustomerManagedPolicyReference(cmpr)
		}
	}
	if v, ok := tfMap["managed_policy_arn"].(string); ok && v != "" {
		apiObject.ManagedPolicyArn = aws.String(v)
	}

	return apiObject
}

func flattenPermissionsBoundary(apiObject *ssoadmin.PermissionsBoundary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ManagedPolicyArn; v != nil {
		tfMap["managed_policy_arn"] = aws.StringValue(v)
	} else if v := apiObject.CustomerManagedPolicyReference; v != nil {
		tfMap["customer_managed_policy_reference"] = []map[string]interface{}{flattenCustomerManagedPolicyReference(v)}
	}

	return tfMap
}
