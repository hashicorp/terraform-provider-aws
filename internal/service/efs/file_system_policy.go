// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_file_system_policy", name="File System Policy")
func resourceFileSystemPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileSystemPolicyPut,
		ReadWithoutTimeout:   resourceFileSystemPolicyRead,
		UpdateWithoutTimeout: resourceFileSystemPolicyPut,
		DeleteWithoutTimeout: resourceFileSystemPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrFileSystemID: {
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
		},
	}
}

func resourceFileSystemPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	fsID := d.Get(names.AttrFileSystemID).(string)
	input := &efs.PutFileSystemPolicyInput{
		BypassPolicyLockoutSafetyCheck: d.Get("bypass_policy_lockout_safety_check").(bool),
		FileSystemId:                   aws.String(fsID),
		Policy:                         aws.String(policy),
	}

	_, err = tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidPolicyException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutFileSystemPolicy(ctx, input)
	}, "Policy contains invalid Principal block")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting EFS File System Policy (%s): %s", fsID, err)
	}

	if d.IsNewResource() {
		d.SetId(fsID)
	}

	return append(diags, resourceFileSystemPolicyRead(ctx, d, meta)...)
}

func resourceFileSystemPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	output, err := findFileSystemPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS File System Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS File System Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrFileSystemID, output.FileSystemId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceFileSystemPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	log.Printf("[DEBUG] Deleting EFS File System Policy: %s", d.Id())
	_, err := conn.DeleteFileSystemPolicy(ctx, &efs.DeleteFileSystemPolicyInput{
		FileSystemId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS File System Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findFileSystemPolicyByID(ctx context.Context, conn *efs.Client, id string) (*efs.DescribeFileSystemPolicyOutput, error) {
	input := &efs.DescribeFileSystemPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeFileSystemPolicy(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) || errs.IsA[*awstypes.PolicyNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
