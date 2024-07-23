// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_session_context", name="Session Context")
func dataSourceSessionContext() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSessionContextRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"issuer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSessionContextRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	arn := d.Get(names.AttrARN).(string)

	d.SetId(arn)

	roleName := ""
	sessionName := ""
	var err error

	if roleName, sessionName = RoleNameSessionFromARN(arn); roleName == "" {
		d.Set("issuer_arn", arn)
		d.Set("issuer_id", "")
		d.Set("issuer_name", "")
		d.Set("session_name", "")

		return diags
	}

	var role *awstypes.Role

	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		role, err = findRoleByName(ctx, conn, roleName)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		role, err = findRoleByName(ctx, conn, roleName)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to get role (%s): %s", roleName, err)
	}

	if role == nil || role.Arn == nil {
		return sdkdiag.AppendErrorf(diags, "empty role returned (%s)", roleName)
	}

	d.Set("issuer_arn", role.Arn)
	d.Set("issuer_id", role.RoleId)
	d.Set("issuer_name", roleName)
	d.Set("session_name", sessionName)

	return diags
}

// RoleNameSessionFromARN returns the role and session names in an ARN if any.
// Otherwise, it returns empty strings.
func RoleNameSessionFromARN(rawARN string) (string, string) {
	parsedARN, err := arn.Parse(rawARN)

	if err != nil {
		return "", ""
	}

	reAssume := regexache.MustCompile(`^assumed-role/.{1,}/.{2,}`)

	if !reAssume.MatchString(parsedARN.Resource) || parsedARN.Service != "sts" {
		return "", ""
	}

	parts := strings.Split(parsedARN.Resource, "/")

	if len(parts) < 3 {
		return "", ""
	}

	return parts[len(parts)-2], parts[len(parts)-1]
}
