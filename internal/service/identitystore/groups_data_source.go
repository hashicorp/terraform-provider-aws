// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_groups", name="Groups")
func DataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

const (
	DSNameGroups = "Groups Data Source"
)

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	input := &identitystore.ListGroupsInput{
		IdentityStoreId: aws.String(identityStoreID),
	}

	paginator := identitystore.NewListGroupsPaginator(conn, input)
	groups := make([]types.Group, 0)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroups, identityStoreID, err)
		}

		groups = append(groups, page.Groups...)
	}

	d.SetId(identityStoreID)

	if err := d.Set("groups", flattenGroups(groups)); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, DSNameGroups, d.Id(), err)
	}

	return nil
}

func flattenGroups(groups []types.Group) []map[string]interface{} {
	if len(groups) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	for _, group := range groups {
		g := make(map[string]interface{}, 1)
		g["description"] = aws.ToString(group.Description)
		g["display_name"] = aws.ToString(group.DisplayName)
		g["external_ids"] = flattenExternalIds(group.ExternalIds)
		g["group_id"] = aws.ToString(group.GroupId)

		result = append(result, g)
	}

	return result
}
