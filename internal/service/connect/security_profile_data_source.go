// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_connect_security_profile")
func DataSourceSecurityProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityProfileRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "security_profile_id"},
			},
			"organization_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
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
				ExactlyOneOf: []string{"security_profile_id", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeSecurityProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("security_profile_id"); ok {
		input.SecurityProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		securityProfileSummary, err := dataSourceGetSecurityProfileSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Security Profile Summary by name (%s): %s", name, err)
		}

		if securityProfileSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Security Profile Summary by name (%s): not found", name)
		}

		input.SecurityProfileId = securityProfileSummary.Id
	}

	resp, err := conn.DescribeSecurityProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Security Profile: %s", err)
	}

	if resp == nil || resp.SecurityProfile == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Security Profile: empty response")
	}

	securityProfile := resp.SecurityProfile

	d.Set("arn", resp.SecurityProfile.Arn)
	d.Set("description", resp.SecurityProfile.Description)
	d.Set("instance_id", instanceID)
	d.Set("organization_resource_id", resp.SecurityProfile.OrganizationResourceId)
	d.Set("security_profile_id", resp.SecurityProfile.Id)
	d.Set("name", resp.SecurityProfile.SecurityProfileName)

	// reading permissions requires a separate API call
	permissions, err := getSecurityProfilePermissions(ctx, conn, instanceID, *resp.SecurityProfile.Id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Security Profile Permissions for Security Profile (%s): %s", *resp.SecurityProfile.Id, err)
	}

	if permissions != nil {
		d.Set("permissions", flex.FlattenStringValueSet(permissions))
	}

	if err := d.Set("tags", KeyValueTags(ctx, securityProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(resp.SecurityProfile.Id)))

	return diags
}

func dataSourceGetSecurityProfileSummaryByName(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.SecurityProfileSummary, error) {
	var result *awstypes.SecurityProfileSummary

	input := &connect.ListSecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListSecurityProfilesMaxResults),
	}

	pages := connect.NewListSecurityProfilesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, qs := range page.SecurityProfileSummaryList {
			if aws.ToString(qs.Name) == name {
				result = &qs
			}
		}
	}

	return result, nil
}
