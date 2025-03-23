// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_user", name="User")
func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"active": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrEmail: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"identity_type": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  defaultUserNamespace,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
					),
				},
				"principal_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrUserName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"user_role": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	namespace := d.Get(names.AttrNamespace).(string)
	userName := d.Get(names.AttrUserName).(string)
	id := userCreateResourceID(awsAccountID, namespace, userName)

	user, err := findUserByThreePartKey(ctx, conn, awsAccountID, namespace, userName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("active", user.Active)
	d.Set(names.AttrARN, user.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrEmail, user.Email)
	d.Set("identity_type", user.IdentityType)
	d.Set("principal_id", user.PrincipalId)
	d.Set(names.AttrUserName, user.UserName)
	d.Set("user_role", user.Role)

	return diags
}
