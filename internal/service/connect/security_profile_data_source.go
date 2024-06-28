// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_security_profile")
func DataSourceSecurityProfile() *schema.Resource {
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

func dataSourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeSecurityProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("security_profile_id"); ok {
		input.SecurityProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
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

	resp, err := conn.DescribeSecurityProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Security Profile: %s", err)
	}

	if resp == nil || resp.SecurityProfile == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Security Profile: empty response")
	}

	securityProfile := resp.SecurityProfile

	d.Set(names.AttrARN, resp.SecurityProfile.Arn)
	d.Set(names.AttrDescription, resp.SecurityProfile.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("organization_resource_id", resp.SecurityProfile.OrganizationResourceId)
	d.Set("security_profile_id", resp.SecurityProfile.Id)
	d.Set(names.AttrName, resp.SecurityProfile.SecurityProfileName)

	// reading permissions requires a separate API call
	permissions, err := getSecurityProfilePermissions(ctx, conn, instanceID, *resp.SecurityProfile.Id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Security Profile Permissions for Security Profile (%s): %s", *resp.SecurityProfile.Id, err)
	}

	if permissions != nil {
		d.Set(names.AttrPermissions, flex.FlattenStringSet(permissions))
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, securityProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(resp.SecurityProfile.Id)))

	return diags
}

func dataSourceGetSecurityProfileSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.SecurityProfileSummary, error) {
	var result *connect.SecurityProfileSummary

	input := &connect.ListSecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListSecurityProfilesMaxResults),
	}

	err := conn.ListSecurityProfilesPagesWithContext(ctx, input, func(page *connect.ListSecurityProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.SecurityProfileSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
				result = qs
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
