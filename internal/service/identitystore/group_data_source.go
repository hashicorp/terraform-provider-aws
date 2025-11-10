// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_group", name="Group")
func dataSourceGroup() *schema.Resource {
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
				ConflictsWith: []string{"group_id"},
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
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexache.MustCompile(`^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`), "must match ([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"),
				),
				AtLeastOneOf:  []string{"alternate_identifier", "group_id"},
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

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	var groupID string

	if v, ok := d.GetOk("alternate_identifier"); ok && len(v.([]any)) > 0 {
		input := identitystore.GetGroupIdInput{
			AlternateIdentifier: expandAlternateIdentifier(v.([]any)[0].(map[string]any)),
			IdentityStoreId:     aws.String(identityStoreID),
		}

		output, err := conn.GetGroupId(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IdentityStore Group (%s): %s", identityStoreID, err)
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

	group, err := findGroupByTwoPartKey(ctx, conn, identityStoreID, groupID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IdentityStore Group (%s): %s", groupID, err)
	}

	d.SetId(aws.ToString(group.GroupId))
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrDisplayName, group.DisplayName)
	if err := d.Set("external_ids", flattenExternalIDs(group.ExternalIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting external_ids: %s", err)
	}
	d.Set("group_id", group.GroupId)

	return diags
}
