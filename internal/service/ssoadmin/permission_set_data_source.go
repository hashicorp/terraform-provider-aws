// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssoadmin_permission_set")
func DataSourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePermissionSetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrARN, names.AttrName},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
				ExactlyOneOf: []string{names.AttrName, names.AttrARN},
			},
			"relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_duration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceArn := d.Get("instance_arn").(string)

	var permissionSet *awstypes.PermissionSet

	if v, ok := d.GetOk(names.AttrARN); ok {
		arn := v.(string)

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		output, err := conn.DescribePermissionSet(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSO Admin Permission Set (%s): %s", arn, err)
		}

		if output == nil {
			return sdkdiag.AppendErrorf(diags, "reading SSO Admin Permission Set (%s): empty output", arn)
		}

		permissionSet = output.PermissionSet
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)

		input := &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
		}

		var permissionSetArns []string
		paginator := ssoadmin.NewListPermissionSetsPaginator(conn, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing SSO Permission Sets: %s", err)
			}

			permissionSetArns = append(permissionSetArns, page.PermissionSets...)
		}

		for _, permissionSetArn := range permissionSetArns {
			output, err := conn.DescribePermissionSet(ctx, &ssoadmin.DescribePermissionSetInput{
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: aws.String(permissionSetArn),
			})

			if err != nil {
				// Proceed with attempting to describe the remaining permission sets
				continue
			}

			if output == nil || output.PermissionSet == nil {
				continue
			}

			if aws.ToString(output.PermissionSet.Name) == name {
				permissionSet = output.PermissionSet
			}
		}
	}

	if permissionSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set: not found")
	}

	arn := aws.ToString(permissionSet.PermissionSetArn)

	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, permissionSet.Description)
	d.Set("instance_arn", instanceArn)
	d.Set(names.AttrName, permissionSet.Name)
	d.Set("session_duration", permissionSet.SessionDuration)
	d.Set("relay_state", permissionSet.RelayState)

	tags, err := listTags(ctx, conn, arn, instanceArn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SSO Permission Set (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
