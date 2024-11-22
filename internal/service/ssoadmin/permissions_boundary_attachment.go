// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									names.AttrPath: {
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
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	tfMap := d.Get("permissions_boundary").([]interface{})[0].(map[string]interface{})
	instanceARN := d.Get("instance_arn").(string)
	permissionSetARN := d.Get("permission_set_arn").(string)
	id := PermissionsBoundaryAttachmentCreateResourceID(permissionSetARN, instanceARN)
	input := &ssoadmin.PutPermissionsBoundaryToPermissionSetInput{
		InstanceArn:         aws.String(instanceARN),
		PermissionSetArn:    aws.String(permissionSetARN),
		PermissionsBoundary: expandPermissionsBoundary(tfMap),
	}

	_, err := conn.PutPermissionsBoundaryToPermissionSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Permissions Boundary Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	// After the policy has been attached to the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourcePermissionsBoundaryAttachmentRead(ctx, d, meta)...)
}

func resourcePermissionsBoundaryAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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

	d.Set("instance_arn", instanceARN)
	d.Set("permission_set_arn", permissionSetARN)
	if err := d.Set("permissions_boundary", []interface{}{flattenPermissionsBoundary(policy)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions_boundary: %s", err)
	}

	return diags
}

func resourcePermissionsBoundaryAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &ssoadmin.DeletePermissionsBoundaryFromPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	_, err = conn.DeletePermissionsBoundaryFromPermissionSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Permissions Boundary Attachment (%s): %s", d.Id(), err)
	}

	// After the policy has been detached from the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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

func FindPermissionsBoundary(ctx context.Context, conn *ssoadmin.Client, permissionSetARN, instanceARN string) (*awstypes.PermissionsBoundary, error) {
	input := &ssoadmin.GetPermissionsBoundaryForPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	output, err := conn.GetPermissionsBoundaryForPermissionSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PermissionsBoundary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PermissionsBoundary, nil
}

func expandPermissionsBoundary(tfMap map[string]interface{}) *awstypes.PermissionsBoundary {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PermissionsBoundary{}

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

func flattenPermissionsBoundary(apiObject *awstypes.PermissionsBoundary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ManagedPolicyArn; v != nil {
		tfMap["managed_policy_arn"] = aws.ToString(v)
	} else if v := apiObject.CustomerManagedPolicyReference; v != nil {
		tfMap["customer_managed_policy_reference"] = []map[string]interface{}{flattenCustomerManagedPolicyReference(v)}
	}

	return tfMap
}
