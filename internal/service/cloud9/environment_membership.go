// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloud9"
	"github.com/aws/aws-sdk-go-v2/service/cloud9/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloud9_environment_membership", name="Environment Membership")
func resourceEnvironmentMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentMembershipCreate,
		ReadWithoutTimeout:   resourceEnvironmentMembershipRead,
		UpdateWithoutTimeout: resourceEnvironmentMembershipUpdate,
		DeleteWithoutTimeout: resourceEnvironmentMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPermissions: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.Permissions](),
			},
			"user_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEnvironmentMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	envID := d.Get("environment_id").(string)
	userARN := d.Get("user_arn").(string)
	id := environmentMembershipCreateResourceID(envID, userARN)
	input := &cloud9.CreateEnvironmentMembershipInput{
		EnvironmentId: aws.String(envID),
		Permissions:   types.MemberPermissions(d.Get(names.AttrPermissions).(string)),
		UserArn:       aws.String(userARN),
	}

	_, err := conn.CreateEnvironmentMembership(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cloud9 Environment Membership (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceEnvironmentMembershipRead(ctx, d, meta)...)
}

func resourceEnvironmentMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	envID, userARN, err := environmentMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	env, err := findEnvironmentMembershipByTwoPartKey(ctx, conn, envID, userARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloud9 Environment Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cloud9 Environment Membership (%s): %s", d.Id(), err)
	}

	d.Set("environment_id", env.EnvironmentId)
	d.Set(names.AttrPermissions, env.Permissions)
	d.Set("user_arn", env.UserArn)
	d.Set("user_id", env.UserId)

	return diags
}

func resourceEnvironmentMembershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	envID, userARN, err := environmentMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := cloud9.UpdateEnvironmentMembershipInput{
		EnvironmentId: aws.String(envID),
		Permissions:   types.MemberPermissions(d.Get(names.AttrPermissions).(string)),
		UserArn:       aws.String(userARN),
	}

	_, err = conn.UpdateEnvironmentMembership(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cloud9 Environment Membership (%s): %s", d.Id(), err)
	}

	return append(diags, resourceEnvironmentMembershipRead(ctx, d, meta)...)
}

func resourceEnvironmentMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Client(ctx)

	envID, userARN, err := environmentMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Cloud9 Environment Membership: %s", d.Id())
	_, err = conn.DeleteEnvironmentMembership(ctx, &cloud9.DeleteEnvironmentMembershipInput{
		EnvironmentId: aws.String(envID),
		UserArn:       aws.String(userARN),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cloud9 Environment Membership (%s): %s", d.Id(), err)
	}

	return diags
}

const environmentMembershipResourceIDSeparator = "#"

func environmentMembershipCreateResourceID(envID, userARN string) string {
	parts := []string{envID, userARN}
	id := strings.Join(parts, environmentMembershipResourceIDSeparator)

	return id
}

func environmentMembershipParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, environmentMembershipResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ENVIRONMENTID%[2]sUSERARN", id, environmentMembershipResourceIDSeparator)
}

func findEnvironmentMembershipByTwoPartKey(ctx context.Context, conn *cloud9.Client, envID, userARN string) (*types.EnvironmentMember, error) {
	input := &cloud9.DescribeEnvironmentMembershipsInput{
		EnvironmentId: aws.String(envID),
		UserArn:       aws.String(userARN),
	}

	output, err := conn.DescribeEnvironmentMemberships(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

	return tfresource.AssertSingleValueResult(output.Memberships)
}
