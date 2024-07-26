// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_group")
func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"alternate_identifier": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExternalID: {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrID: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrIssuer: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"unique_attribute": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"attribute_value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
				ConflictsWith: []string{names.AttrFilter, "group_id"},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDisplayName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: {
				Deprecated:    "Use the alternate_identifier attribute instead.",
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				AtLeastOneOf:  []string{"alternate_identifier", names.AttrFilter, "group_id"},
				ConflictsWith: []string{"alternate_identifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexache.MustCompile(`^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`), "must match ([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"),
				),
				AtLeastOneOf:  []string{"alternate_identifier", names.AttrFilter, "group_id"},
				ConflictsWith: []string{"alternate_identifier"},
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]*$`), "must match [0-9A-Za-z-]"),
				),
			},
		},
	}
}

const (
	DSNameGroup = "Group Data Source"
)

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	if v, ok := d.GetOk(names.AttrFilter); ok && len(v.([]interface{})) > 0 {
		// Use ListGroups for backwards compat.
		input := &identitystore.ListGroupsInput{
			IdentityStoreId: aws.String(identityStoreID),
			Filters:         expandFilters(d.Get(names.AttrFilter).([]interface{})),
		}
		paginator := identitystore.NewListGroupsPaginator(conn, input)
		var results []types.Group

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)

			if err != nil {
				return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreID, err)
			}

			for _, group := range page.Groups {
				if v, ok := d.GetOk("group_id"); ok && v.(string) != aws.ToString(group.GroupId) {
					continue
				}

				results = append(results, group)
			}
		}

		if len(results) == 0 {
			return sdkdiag.AppendErrorf(diags, "no Identity Store Group found matching criteria\n%v; try different search", input.Filters)
		}

		if len(results) > 1 {
			return sdkdiag.AppendErrorf(diags, "multiple Identity Store Groups found matching criteria\n%v; try different search", input.Filters)
		}

		group := results[0]

		d.SetId(aws.ToString(group.GroupId))
		d.Set(names.AttrDescription, group.Description)
		d.Set(names.AttrDisplayName, group.DisplayName)
		d.Set("group_id", group.GroupId)

		if err := d.Set("external_ids", flattenExternalIds(group.ExternalIds)); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameGroup, d.Id(), err)
		}

		return diags
	}

	var groupID string

	if v, ok := d.GetOk("alternate_identifier"); ok && len(v.([]interface{})) > 0 {
		input := &identitystore.GetGroupIdInput{
			AlternateIdentifier: expandAlternateIdentifier(v.([]interface{})[0].(map[string]interface{})),
			IdentityStoreId:     aws.String(identityStoreID),
		}

		output, err := conn.GetGroupId(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreID, err)
		}

		groupID = aws.ToString(output.GroupId)
	}

	if v, ok := d.GetOk("group_id"); ok && v.(string) != "" {
		if groupID != "" && groupID != v.(string) {
			// We were given a filter, and it found a group different to this one.
			return sdkdiag.AppendErrorf(diags, "no Identity Store Group found matching criteria; try different search")
		}

		groupID = v.(string)
	}

	group, err := FindGroupByTwoPartKey(ctx, conn, identityStoreID, groupID)

	if err != nil {
		if tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "no Identity Store Group found matching criteria; try different search")
		}

		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreID, err)
	}

	d.SetId(aws.ToString(group.GroupId))

	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrDisplayName, group.DisplayName)
	d.Set("group_id", group.GroupId)

	if err := d.Set("external_ids", flattenExternalIds(group.ExternalIds)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameGroup, d.Id(), err)
	}

	return diags
}

func expandFilters(l []interface{}) []types.Filter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	filters := make([]types.Filter, 0, len(l))
	for _, v := range l {
		tfMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		filter := types.Filter{}

		if v, ok := tfMap["attribute_path"].(string); ok && v != "" {
			filter.AttributePath = aws.String(v)
		}

		if v, ok := tfMap["attribute_value"].(string); ok && v != "" {
			filter.AttributeValue = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}
