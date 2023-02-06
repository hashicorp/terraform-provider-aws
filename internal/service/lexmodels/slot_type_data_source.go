package lexmodels

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceSlotType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSlotTypeRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  SlotTypeVersionLatest,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
				),
			},
		},
	}
}

func dataSourceSlotTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn()

	name := d.Get("name").(string)
	version := d.Get("version").(string)
	output, err := FindSlotTypeVersionByName(ctx, conn, name, version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Slot Type (%s/%s): %s", name, version, err)
	}

	d.SetId(name)
	d.Set("checksum", output.Checksum)
	d.Set("created_date", output.CreatedDate.Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("enumeration_value", flattenEnumerationValues(output.EnumerationValues))
	d.Set("last_updated_date", output.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", output.Name)
	d.Set("value_selection_strategy", output.ValueSelectionStrategy)
	d.Set("version", output.Version)

	return diags
}
