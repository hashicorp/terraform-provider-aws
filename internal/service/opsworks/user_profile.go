// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_opsworks_user_profile", name="Profile")
func resourceUserProfile() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage:   "This resource is deprecated and will be removed in the next major version of the AWS Provider. Consider the AWS Systems Manager service instead.",
		CreateWithoutTimeout: resourceUserProfileCreate,
		ReadWithoutTimeout:   resourceUserProfileRead,
		UpdateWithoutTimeout: resourceUserProfileUpdate,
		DeleteWithoutTimeout: resourceUserProfileDelete,

		Schema: map[string]*schema.Schema{
			"allow_self_management": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	iamUserARN := d.Get("user_arn").(string)
	input := &opsworks.CreateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(iamUserARN),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	if v, ok := d.GetOk("ssh_public_key"); ok {
		input.SshPublicKey = aws.String(v.(string))
	}

	_, err := conn.CreateUserProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks User Profile (%s): %s", iamUserARN, err)
	}

	d.SetId(iamUserARN)

	return append(diags, resourceUserProfileUpdate(ctx, d, meta)...)
}

func resourceUserProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	profile, err := findUserProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks User Profile %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	d.Set("allow_self_management", profile.AllowSelfManagement)
	d.Set("ssh_public_key", profile.SshPublicKey)
	d.Set("ssh_username", profile.SshUsername)
	d.Set("user_arn", profile.IamUserArn)

	return diags
}

func resourceUserProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	input := &opsworks.UpdateUserProfileInput{
		AllowSelfManagement: aws.Bool(d.Get("allow_self_management").(bool)),
		IamUserArn:          aws.String(d.Get("user_arn").(string)),
		SshPublicKey:        aws.String(d.Get("ssh_public_key").(string)),
		SshUsername:         aws.String(d.Get("ssh_username").(string)),
	}

	_, err := conn.UpdateUserProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserProfileRead(ctx, d, meta)...)
}

func resourceUserProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	log.Printf("[DEBUG] Deleting OpsWorks User Profile: %s", d.Id())
	_, err := conn.DeleteUserProfile(ctx, &opsworks.DeleteUserProfileInput{
		IamUserArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks User Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserProfileByARN(ctx context.Context, conn *opsworks.Client, arn string) (*awstypes.UserProfile, error) {
	input := &opsworks.DescribeUserProfilesInput{
		IamUserArns: []string{arn},
	}

	output, err := conn.DescribeUserProfiles(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.UserProfiles) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.UserProfiles); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return tfresource.AssertSingleValueResult(output.UserProfiles)
}
