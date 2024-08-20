// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_findings_filter", name="Findings Filter")
// @Tags
func ResourceFindingsFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFindingsFilterCreate,
		ReadWithoutTimeout:   resourceFindingsFilterRead,
		UpdateWithoutTimeout: resourceFindingsFilterUpdate,
		DeleteWithoutTimeout: resourceFindingsFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"finding_criteria": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrField: {
										Type:     schema.TypeString,
										Required: true,
									},
									"eq_exact_match": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"eq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"neq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"lt": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"lte": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"gt": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"gte": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(3, 64),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(3, 64-id.UniqueIDSuffixLength),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			names.AttrAction: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(macie2.FindingsFilterAction_Values(), false),
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFindingsFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.CreateFindingsFilterInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))),
		Action:      aws.String(d.Get(names.AttrAction).(string)),
		Tags:        getTagsIn(ctx),
	}

	var err error
	input.FindingCriteria, err = expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie FindingsFilter: %s", err)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("position"); ok {
		input.Position = aws.Int64(int64(v.(int)))
	}

	var output *macie2.CreateFindingsFilterOutput
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		output, err = conn.CreateFindingsFilterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateFindingsFilterWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie FindingsFilter: %s", err)
	}

	d.SetId(aws.StringValue(output.Id))

	return append(diags, resourceFindingsFilterRead(ctx, d, meta)...)
}

func resourceFindingsFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.GetFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetFindingsFilterWithContext(ctx, input)

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled")) {
		log.Printf("[WARN] Macie FindingsFilter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie FindingsFilter (%s): %s", d.Id(), err)
	}

	if err = d.Set("finding_criteria", flattenFindingCriteriaFindingsFilter(resp.FindingCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie FindingsFilter (%s): %s", "finding_criteria", d.Id(), err)
	}
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set(names.AttrDescription, resp.Description)
	d.Set(names.AttrAction, resp.Action)
	d.Set("position", resp.Position)

	setTagsOut(ctx, resp.Tags)

	d.Set(names.AttrARN, resp.Arn)

	return diags
}

func resourceFindingsFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.UpdateFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	var err error
	if d.HasChange("finding_criteria") {
		input.FindingCriteria, err = expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{}))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie FindingsFilter (%s): %s", d.Id(), err)
		}
	}
	if d.HasChange(names.AttrName) {
		input.Name = aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string)))
	}
	if d.HasChange(names.AttrNamePrefix) {
		input.Name = aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string)))
	}
	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}
	if d.HasChange(names.AttrAction) {
		input.Action = aws.String(d.Get(names.AttrAction).(string))
	}
	if d.HasChange("position") {
		input.Position = aws.Int64(int64(d.Get("position").(int)))
	}

	_, err = conn.UpdateFindingsFilterWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Macie FindingsFilter (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFindingsFilterRead(ctx, d, meta)...)
}

func resourceFindingsFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.DeleteFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteFindingsFilterWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Macie FindingsFilter (%s): %s", d.Id(), err)
	}
	return diags
}

func expandFindingCriteriaFilter(findingCriterias []interface{}) (*macie2.FindingCriteria, error) {
	if len(findingCriterias) == 0 {
		return nil, nil
	}

	criteria := map[string]*macie2.CriterionAdditionalProperties{}
	findingCriteria := findingCriterias[0].(map[string]interface{})
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	for _, criterion := range inputFindingCriteria {
		crit := criterion.(map[string]interface{})
		field := crit[names.AttrField].(string)
		conditional := macie2.CriterionAdditionalProperties{}

		if v, ok := crit["eq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.Eq = foo
		}
		if v, ok := crit["neq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.Neq = foo
		}
		if v, ok := crit["eq_exact_match"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.EqExactMatch = foo
		}
		if v, ok := crit["lt"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("parsing condition %q for field %q: %w", "lt", field, err)
			}
			conditional.Lt = aws.Int64(i)
		}
		if v, ok := crit["lte"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("parsing condition %q for field %q: %w", "lte", field, err)
			}
			conditional.Lte = aws.Int64(i)
		}
		if v, ok := crit["gt"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("parsing condition %q for field %q: %w", "gt", field, err)
			}
			conditional.Gt = aws.Int64(i)
		}
		if v, ok := crit["gte"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("parsing condition %q for field %q: %w", "gte", field, err)
			}
			conditional.Gte = aws.Int64(i)
		}
		criteria[field] = &conditional
	}

	return &macie2.FindingCriteria{Criterion: criteria}, nil
}

func flattenFindingCriteriaFindingsFilter(findingCriteria *macie2.FindingCriteria) []interface{} {
	if findingCriteria == nil {
		return nil
	}

	var flatCriteria []interface{}

	for field, conditions := range findingCriteria.Criterion {
		criterion := map[string]interface{}{
			names.AttrField: field,
		}
		if len(conditions.Eq) != 0 {
			criterion["eq"] = aws.StringValueSlice(conditions.Eq)
		}
		if len(conditions.Neq) != 0 {
			criterion["neq"] = aws.StringValueSlice(conditions.Neq)
		}
		if len(conditions.EqExactMatch) != 0 {
			criterion["eq_exact_match"] = aws.StringValueSlice(conditions.EqExactMatch)
		}
		if v := aws.Int64Value(conditions.Lt); v != 0 {
			criterion["lt"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Lte); v != 0 {
			criterion["lte"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Gt); v != 0 {
			criterion["gt"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Gte); v != 0 {
			criterion["gte"] = flattenConditionIntField(field, v)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []interface{}{
		map[string][]interface{}{
			"criterion": flatCriteria,
		},
	}
}

func expandConditionIntField(field, v string) (int64, error) {
	if field == "updatedAt" {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		return date.UnixNano() / 1000000, nil
	}

	return strconv.ParseInt(v, 10, 64)
}

func flattenConditionIntField(field string, v int64) string {
	if field == "updatedAt" {
		seconds := v / 1000
		nanoseconds := v % 1000
		date := time.Unix(seconds, nanoseconds).UTC()
		return date.Format(time.RFC3339)
	}
	return strconv.FormatInt(v, 10)
}
