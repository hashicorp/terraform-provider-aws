// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_user", name="User")
func DataSourceUser() *schema.Resource {
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
					Default:  DefaultUserNamespace,
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

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	namespace := d.Get(names.AttrNamespace).(string)
	in := &quicksight.DescribeUserInput{
		UserName:     aws.String(d.Get(names.AttrUserName).(string)),
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
	}

	out, err := conn.DescribeUserWithContext(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", d.Id(), err)
	}
	if out == nil || out.User == nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", d.Id(), tfresource.NewEmptyResultError(in))
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(out.User.UserName)))
	d.Set("active", out.User.Active)
	d.Set(names.AttrARN, out.User.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrEmail, out.User.Email)
	d.Set("identity_type", out.User.IdentityType)
	d.Set("principal_id", out.User.PrincipalId)
	d.Set(names.AttrUserName, out.User.UserName)
	d.Set("user_role", out.User.Role)

	return diags
}
