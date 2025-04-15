// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_findings_filter", name="Findings Filter")
// @Tags(identifierAttribute="arn")
func resourceFindingsFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFindingsFilterCreate,
		ReadWithoutTimeout:   resourceFindingsFilterRead,
		UpdateWithoutTimeout: resourceFindingsFilterUpdate,
		DeleteWithoutTimeout: resourceFindingsFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FindingsFilterAction](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
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
									"eq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"eq_exact_match": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrField: {
										Type:     schema.TypeString,
										Required: true,
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
									"neq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
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
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
		},
	}
}

func resourceFindingsFilterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := macie2.CreateFindingsFilterInput{
		Action:      awstypes.FindingsFilterAction(d.Get(names.AttrAction).(string)),
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, err := expandFindingCriteriaFilter(d.Get("finding_criteria").([]any)); err == nil {
		input.FindingCriteria = v
	} else {
		return sdkdiag.AppendErrorf(diags, "expanding finding_criteria: %s", err)
	}

	if v, ok := d.GetOk("position"); ok {
		input.Position = aws.Int32(int32(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateFindingsFilter(ctx, &input)
	}, errCodeClientError)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Findings Filter (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*macie2.CreateFindingsFilterOutput).Id))

	return append(diags, resourceFindingsFilterRead(ctx, d, meta)...)
}

func resourceFindingsFilterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	output, err := findFindingsFilterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Findings Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie Findings Filter (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAction, output.Action)
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	if err = d.Set("finding_criteria", flattenFindingCriteriaFindingsFilter(output.FindingCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting finding_criteria: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set("position", output.Position)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceFindingsFilterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := macie2.UpdateFindingsFilterInput{
			Id: aws.String(d.Id()),
		}

		var err error
		if d.HasChange("finding_criteria") {
			input.FindingCriteria, err = expandFindingCriteriaFilter(d.Get("finding_criteria").([]any))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Macie FindingsFilter (%s): %s", d.Id(), err)
			}
		}

		if d.HasChange(names.AttrAction) {
			input.Action = awstypes.FindingsFilterAction(d.Get(names.AttrAction).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("finding_criteria") {
			if v, err := expandFindingCriteriaFilter(d.Get("finding_criteria").([]any)); err == nil {
				input.FindingCriteria = v
			} else {
				return sdkdiag.AppendErrorf(diags, "expanding finding_criteria: %s", err)
			}
		}

		if d.HasChanges(names.AttrName, names.AttrNamePrefix) {
			input.Name = aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string)))
		}

		if d.HasChange("position") {
			input.Position = aws.Int32(int32(d.Get("position").(int)))
		}

		_, err = conn.UpdateFindingsFilter(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie Findings Filter (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFindingsFilterRead(ctx, d, meta)...)
}

func resourceFindingsFilterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	_, err := conn.DeleteFindingsFilter(ctx, &macie2.DeleteFindingsFilterInput{
		Id: aws.String(d.Id()),
	})

	if isFindingsFilterNotFoundError(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Findings Filter (%s): %s", d.Id(), err)
	}

	return diags
}

func findFindingsFilterByID(ctx context.Context, conn *macie2.Client, id string) (*macie2.GetFindingsFilterOutput, error) {
	input := macie2.GetFindingsFilterInput{
		Id: aws.String(id),
	}

	return findFindingsFilter(ctx, conn, &input)
}

func findFindingsFilter(ctx context.Context, conn *macie2.Client, input *macie2.GetFindingsFilterInput) (*macie2.GetFindingsFilterOutput, error) {
	output, err := conn.GetFindingsFilter(ctx, input)

	if isFindingsFilterNotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func isFindingsFilterNotFoundError(err error) bool {
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
		return true
	}

	return false
}

func expandFindingCriteriaFilter(findingCriterias []any) (*awstypes.FindingCriteria, error) {
	if len(findingCriterias) == 0 {
		return nil, nil
	}

	criteria := map[string]awstypes.CriterionAdditionalProperties{}
	findingCriteria := findingCriterias[0].(map[string]any)
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	for _, criterion := range inputFindingCriteria {
		crit := criterion.(map[string]any)
		field := crit[names.AttrField].(string)
		conditional := awstypes.CriterionAdditionalProperties{}

		if v, ok := crit["eq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = s
			}
			conditional.Eq = foo
		}
		if v, ok := crit["neq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = s
			}
			conditional.Neq = foo
		}
		if v, ok := crit["eq_exact_match"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = s
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
		criteria[field] = conditional
	}

	return &awstypes.FindingCriteria{Criterion: criteria}, nil
}

func flattenFindingCriteriaFindingsFilter(findingCriteria *awstypes.FindingCriteria) []any {
	if findingCriteria == nil {
		return nil
	}

	var flatCriteria []any

	for field, conditions := range findingCriteria.Criterion {
		criterion := map[string]any{
			names.AttrField: field,
		}
		if len(conditions.Eq) != 0 {
			criterion["eq"] = conditions.Eq
		}
		if len(conditions.Neq) != 0 {
			criterion["neq"] = conditions.Neq
		}
		if len(conditions.EqExactMatch) != 0 {
			criterion["eq_exact_match"] = conditions.EqExactMatch
		}
		if v := aws.ToInt64(conditions.Lt); v != 0 {
			criterion["lt"] = flattenConditionIntField(field, v)
		}
		if v := aws.ToInt64(conditions.Lte); v != 0 {
			criterion["lte"] = flattenConditionIntField(field, v)
		}
		if v := aws.ToInt64(conditions.Gt); v != 0 {
			criterion["gt"] = flattenConditionIntField(field, v)
		}
		if v := aws.ToInt64(conditions.Gte); v != 0 {
			criterion["gte"] = flattenConditionIntField(field, v)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []any{
		map[string][]any{
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
