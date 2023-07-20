// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_quicksight_group", name="Group")
func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"aws_account_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"group_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"namespace": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  DefaultGroupNamespace,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
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
	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}
	groupName := d.Get("group_name").(string)
	namespace := d.Get("namespace").(string)
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
	d.Set("arn", group.Arn)
	d.Set("aws_account_id", awsAccountID)
	d.Set("description", group.Description)
	d.Set("group_name", group.GroupName)
	d.Set("principal_id", group.PrincipalId)

	return diags
}
