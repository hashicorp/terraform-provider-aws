// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	slotTypeCreateTimeout = 1 * time.Minute
	slotTypeUpdateTimeout = 1 * time.Minute
	slotTypeDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_lex_slot_type", name="Slot Type")
func resourceSlotType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSlotTypeCreate,
		ReadWithoutTimeout:   resourceSlotTypeRead,
		UpdateWithoutTimeout: resourceSlotTypeUpdate,
		DeleteWithoutTimeout: resourceSlotTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(slotTypeCreateTimeout),
			Update: schema.DefaultTimeout(slotTypeUpdateTimeout),
			Delete: schema.DefaultTimeout(slotTypeDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_version": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"enumeration_value": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"synonyms": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 140),
							},
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 140),
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
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^((AMAZON\.)_?|[A-Za-z]_?)+`), ""),
				),
			},
			"value_selection_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.SlotValueSelectionStrategyOriginalValue,
				ValidateDiagFunc: enum.Validate[awstypes.SlotValueSelectionStrategy](),
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: updateComputedAttributesOnSlotTypeCreateVersion,
	}
}

func updateComputedAttributesOnSlotTypeCreateVersion(_ context.Context, d *schema.ResourceDiff, meta any) error {
	createVersion := d.Get("create_version").(bool)
	if createVersion && hasSlotTypeConfigChanges(d) {
		d.SetNewComputed(names.AttrVersion)
	}
	return nil
}

func hasSlotTypeConfigChanges(d sdkv2.ResourceDiffer) bool {
	return slices.ContainsFunc([]string{
		names.AttrDescription,
		"enumeration_value",
		"value_selection_strategy",
	}, d.HasChange)
}

func resourceSlotTypeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &lexmodelbuildingservice.PutSlotTypeInput{
		CreateVersion:          aws.Bool(d.Get("create_version").(bool)),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		Name:                   aws.String(name),
		ValueSelectionStrategy: awstypes.SlotValueSelectionStrategy(d.Get("value_selection_strategy").(string)),
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandEnumerationValues(v.(*schema.Set).List())
	}

	var output *lexmodelbuildingservice.PutSlotTypeOutput
	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		var err error

		if output != nil {
			input.Checksum = output.Checksum
		}
		output, err = conn.PutSlotType(ctx, input)

		return output, err
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lex Slot Type (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceSlotTypeRead(ctx, d, meta)...)
}

func resourceSlotTypeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	output, err := findSlotTypeVersionByName(ctx, conn, d.Id(), SlotTypeVersionLatest)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lex Slot Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Slot Type (%s): %s", d.Id(), err)
	}

	d.Set("checksum", output.Checksum)
	d.Set(names.AttrCreatedDate, output.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrLastUpdatedDate, output.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, output.Name)
	d.Set("value_selection_strategy", output.ValueSelectionStrategy)

	if output.EnumerationValues != nil {
		d.Set("enumeration_value", flattenEnumerationValues(output.EnumerationValues))
	}

	version, err := findLatestSlotTypeVersionByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex Slot Type (%s) latest version: %s", d.Id(), err)
	}

	d.Set(names.AttrVersion, version)

	return diags
}

func resourceSlotTypeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Checksum:               aws.String(d.Get("checksum").(string)),
		CreateVersion:          aws.Bool(d.Get("create_version").(bool)),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		Name:                   aws.String(d.Id()),
		ValueSelectionStrategy: awstypes.SlotValueSelectionStrategy(d.Get("value_selection_strategy").(string)),
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandEnumerationValues(v.(*schema.Set).List())
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
		return conn.PutSlotType(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lex Slot Type (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSlotTypeRead(ctx, d, meta)...)
}

func resourceSlotTypeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	input := &lexmodelbuildingservice.DeleteSlotTypeInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Lex Slot Type: (%s)", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteSlotType(ctx, input)
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Slot Type (%s): %s", d.Id(), err)
	}

	if _, err := waitSlotTypeDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Slot Type (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func flattenEnumerationValues(values []awstypes.EnumerationValue) (flattened []map[string]any) {
	for _, value := range values {
		flattened = append(flattened, map[string]any{
			"synonyms":      flex.FlattenStringValueList(value.Synonyms),
			names.AttrValue: aws.ToString(value.Value),
		})
	}

	return
}

func expandEnumerationValues(rawValues []any) []awstypes.EnumerationValue {
	enums := make([]awstypes.EnumerationValue, 0, len(rawValues))
	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]any)
		if !ok {
			continue
		}

		enums = append(enums, awstypes.EnumerationValue{
			Synonyms: flex.ExpandStringValueSet(value["synonyms"].(*schema.Set)),
			Value:    aws.String(value[names.AttrValue].(string)),
		})
	}
	return enums
}
