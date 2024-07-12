// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lex_slot_type")
func DataSourceSlotType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSlotTypeRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enumeration_value": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"synonyms": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  SlotTypeVersionLatest,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`\$LATEST|[0-9]+`), ""),
				),
			},
		},
	}
}

func dataSourceSlotTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	name := d.Get(names.AttrName).(string)
	version := d.Get(names.AttrVersion).(string)
	output, err := FindSlotTypeVersionByName(ctx, conn, name, version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Slot Type (%s/%s): %s", name, version, err)
	}

	d.SetId(name)
	d.Set("checksum", output.Checksum)
	d.Set(names.AttrCreatedDate, output.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("enumeration_value", flattenEnumerationValues(output.EnumerationValues))
	d.Set(names.AttrLastUpdatedDate, output.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, output.Name)
	d.Set("value_selection_strategy", output.ValueSelectionStrategy)
	d.Set(names.AttrVersion, output.Version)

	return diags
}
