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

// @SDKDataSource("aws_quicksight_group", name="Group")
func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrGroupName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  DefaultGroupNamespace,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
					),
				},
				"principal_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	groupName := d.Get(names.AttrGroupName).(string)
	namespace := d.Get(names.AttrNamespace).(string)
	in := &quicksight.DescribeGroupInput{
		GroupName:    aws.String(groupName),
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
	}

	out, err := conn.DescribeGroupWithContext(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Group (%s): %s", groupName, err)
	}
	if out == nil || out.Group == nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Group (%s): %s", groupName, tfresource.NewEmptyResultError(in))
	}

	group := out.Group
	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(group.GroupName)))
	d.Set(names.AttrARN, group.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrGroupName, group.GroupName)
	d.Set("principal_id", group.PrincipalId)

	return diags
}
