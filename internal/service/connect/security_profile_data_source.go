// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_security_profile", name="Security Profile")
// @Tags
func dataSourceSecurityProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityProfileRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "security_profile_id"},
			},
			"organization_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPermissions: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"security_profile_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"security_profile_id", names.AttrName},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeSecurityProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("security_profile_id"); ok {
		input.SecurityProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		securityProfileSummary, err := findSecurityProfileSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Security Profile (%s) summary: %s", name, err)
		}

		input.SecurityProfileId = securityProfileSummary.Id
	}

	securityProfile, err := findSecurityProfile(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Security Profile: %s", err)
	}

	securityProfileID := aws.ToString(securityProfile.Id)
	id := securityProfileCreateResourceID(instanceID, securityProfileID)
	d.SetId(id)
	d.Set(names.AttrARN, securityProfile.Arn)
	d.Set(names.AttrDescription, securityProfile.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, securityProfile.SecurityProfileName)
	d.Set("organization_resource_id", securityProfile.OrganizationResourceId)
	d.Set("security_profile_id", securityProfileID)

	permissions, err := findSecurityProfilePermissionsByTwoPartKey(ctx, conn, instanceID, securityProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Security Profile (%s) permissions: %s", d.Id(), err)
	}

	d.Set(names.AttrPermissions, permissions)

	setTagsOut(ctx, securityProfile.Tags)

	return diags
}

func findSecurityProfileSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.SecurityProfileSummary, error) {
	const maxResults = 60
	input := &connect.ListSecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findSecurityProfileSummary(ctx, conn, input, func(v *awstypes.SecurityProfileSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findSecurityProfileSummary(ctx context.Context, conn *connect.Client, input *connect.ListSecurityProfilesInput, filter tfslices.Predicate[*awstypes.SecurityProfileSummary]) (*awstypes.SecurityProfileSummary, error) {
	output, err := findSecurityProfileSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSecurityProfileSummaries(ctx context.Context, conn *connect.Client, input *connect.ListSecurityProfilesInput, filter tfslices.Predicate[*awstypes.SecurityProfileSummary]) ([]awstypes.SecurityProfileSummary, error) {
	var output []awstypes.SecurityProfileSummary

	pages := connect.NewListSecurityProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.SecurityProfileSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
