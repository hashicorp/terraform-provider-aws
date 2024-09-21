// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @@SDKDataSource("aws_iam_service_linked_role", name="Service Linked Role")
func dataSourceServiceLinkedRole() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceLinkedRoleRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`\.`), "must be a full service hostname e.g. elasticbeanstalk.amazonaws.com"),
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_suffix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceLinkedRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var nameRegex string
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	serviceName, _, customSuffix, err := DecodeServiceLinkedRoleID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	//AWS API does not provide a Get method for Service Linked Roles.
	//Matching the role path prefix and role name using regex is the only option to find Service Linked roles
	pathPrefix := fmt.Sprintf("/aws-service-role/%s", serviceName)
	if customSuffix == "" {
		nameRegex = `AWSServiceRole[^_\s]+` //regex to match AWSServiceRole prefix and 1 or more characters exluding _ and white spaces
	} else {
		nameRegex = fmt.Sprintf(`AWSServiceRole\w+_%s`, customSuffix) //regex to match AWSServiceRole prefix, 1 or more characters and the provided suffix
	}

	roles, err := findRoles(ctx, conn, pathPrefix, nameRegex)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Linked Role (%s): %s", d.Id(), err)
	}
	if len(roles) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Linked Role (%s): role not found", d.Id())
	}
	if len(roles) > 1 {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Linked Role (%s): more than one role was returned", d.Id())
	}
	role := roles[0]
	d.Set(names.AttrARN, role.Arn)
	d.Set("aws_service_name", serviceName)
	d.Set("create_date", aws.ToTime(role.CreateDate).Format(time.RFC3339))
	d.Set("custom_suffix", customSuffix)
	d.Set(names.AttrDescription, role.Description)
	d.Set(names.AttrName, role.RoleName)
	d.Set(names.AttrPath, role.Path)
	d.Set("unique_id", role.RoleId)

	setTagsOut(ctx, role.Tags)

	return diags
}
