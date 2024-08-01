// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_filter", name="Filter")
// @Tags(identifierAttribute="arn")
func ResourceFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFilterCreate,
		ReadWithoutTimeout:   resourceFilterRead,
		UpdateWithoutTimeout: resourceFilterUpdate,
		DeleteWithoutTimeout: resourceFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(guardduty.FilterAction_Values(), false),
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
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"finding_criteria": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrField: {
										Type:     schema.TypeString,
										Required: true,
									},
									"greater_than": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"greater_than_or_equal": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"less_than": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"less_than_or_equal": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidStringDateOrPositiveInt,
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_.-]+$`),
						"only alphanumeric characters, hyphens, underscores, and periods are allowed"),
				),
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	input := guardduty.CreateFilterInput{
		Action:      aws.String(d.Get(names.AttrAction).(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get(names.AttrName).(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
		Tags:        getTagsIn(ctx),
	}

	var err error
	input.FindingCriteria, err = expandFindingCriteria(d.Get("finding_criteria").([]interface{}))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Filter: %s", err)
	}

	log.Printf("[DEBUG] Creating GuardDuty Filter: %s", input)
	output, err := conn.CreateFilterWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Filter: %s", err)
	}

	d.SetId(filterCreateID(d.Get("detector_id").(string), aws.StringValue(output.Name)))

	return append(diags, resourceFilterRead(ctx, d, meta)...)
}

func resourceFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	var detectorID, name string
	var err error

	if _, ok := d.GetOk("detector_id"); !ok {
		// If there is no "detector_id" passed, then it's an import.
		detectorID, name, err = FilterParseID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading GuardDuty Filter (%s): %s", name, err)
		}
	} else {
		detectorID = d.Get("detector_id").(string)
		name = d.Get(names.AttrName).(string)
	}

	input := guardduty.GetFilterInput{
		DetectorId: aws.String(detectorID),
		FilterName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading GuardDuty Filter: %s", input)
	filter, err := conn.GetFilterWithContext(ctx, &input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected since no such resource found.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Filter (%s): %s", name, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s/filter/%s", detectorID, name),
	}.String()
	d.Set(names.AttrARN, arn)

	err = d.Set("finding_criteria", flattenFindingCriteria(filter.FindingCriteria))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Setting GuardDuty Filter FindingCriteria failed: %s", err)
	}

	d.Set(names.AttrAction, filter.Action)
	d.Set(names.AttrDescription, filter.Description)
	d.Set(names.AttrName, filter.Name)
	d.Set("detector_id", detectorID)
	d.Set("rank", filter.Rank)

	setTagsOut(ctx, filter.Tags)

	d.SetId(filterCreateID(detectorID, name))

	return diags
}

func resourceFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	if d.HasChanges(names.AttrAction, names.AttrDescription, "finding_criteria", "rank") {
		input := guardduty.UpdateFilterInput{
			Action:      aws.String(d.Get(names.AttrAction).(string)),
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			DetectorId:  aws.String(d.Get("detector_id").(string)),
			FilterName:  aws.String(d.Get(names.AttrName).(string)),
			Rank:        aws.Int64(int64(d.Get("rank").(int))),
		}

		var err error
		input.FindingCriteria, err = expandFindingCriteria(d.Get("finding_criteria").([]interface{}))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Filter %s: %s", d.Id(), err)
		}

		_, err = conn.UpdateFilterWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Filter %s: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFilterRead(ctx, d, meta)...)
}

func resourceFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorId := d.Get("detector_id").(string)
	name := d.Get(names.AttrName).(string)

	input := guardduty.DeleteFilterInput{
		FilterName: aws.String(name),
		DetectorId: aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Filter: %s", input)

	_, err := conn.DeleteFilterWithContext(ctx, &input)
	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected since no such resource found.") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Filter %s: %s", d.Id(), err)
	}
	return diags
}

const filterIDSeparator = ":"

func filterCreateID(detectorID, filterName string) string {
	return detectorID + filterIDSeparator + filterName
}

func FilterParseID(importedId string) (string, string, error) {
	parts := strings.Split(importedId, filterIDSeparator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("GuardDuty filter ID must be of the form <Detector ID>:<Filter name>. Got %q.", importedId)
	}
	return parts[0], parts[1], nil
}

func expandFindingCriteria(raw []interface{}) (*guardduty.FindingCriteria, error) {
	findingCriteria := raw[0].(map[string]interface{})
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	criteria := map[string]*guardduty.Condition{}
	for _, criterion := range inputFindingCriteria {
		typedCriterion := criterion.(map[string]interface{})
		field := typedCriterion[names.AttrField].(string)

		condition := guardduty.Condition{}
		if x, ok := typedCriterion["equals"]; ok {
			if v, ok := x.([]interface{}); ok && len(v) > 0 {
				foo := make([]*string, len(v))
				for i := range v {
					s := v[i].(string)
					foo[i] = &s
				}
				condition.Equals = foo
			}
		}
		if x, ok := typedCriterion["not_equals"]; ok {
			if v, ok := x.([]interface{}); ok && len(v) > 0 {
				foo := make([]*string, len(v))
				for i := range v {
					s := v[i].(string)
					foo[i] = &s
				}
				condition.NotEquals = foo
			}
		}
		if x, ok := typedCriterion["greater_than"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("parsing condition %q for field %q: %w", "greater_than", field, err)
				}
				condition.GreaterThan = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["greater_than_or_equal"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("parsing condition %q for field %q: %w", "greater_than_or_equal", field, err)
				}
				condition.GreaterThanOrEqual = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["less_than"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("parsing condition %q for field %q: %w", "less_than", field, err)
				}
				condition.LessThan = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["less_than_or_equal"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("parsing condition %q for field %q: %w", "less_than_or_equal", field, err)
				}
				condition.LessThanOrEqual = aws.Int64(i)
			}
		}
		criteria[field] = &condition
	}

	return &guardduty.FindingCriteria{Criterion: criteria}, nil
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

func flattenFindingCriteria(findingCriteriaRemote *guardduty.FindingCriteria) []interface{} {
	var flatCriteria []interface{}

	for field, conditions := range findingCriteriaRemote.Criterion {
		criterion := map[string]interface{}{
			names.AttrField: field,
		}
		if len(conditions.Equals) > 0 {
			criterion["equals"] = aws.StringValueSlice(conditions.Equals)
		}
		if len(conditions.NotEquals) > 0 {
			criterion["not_equals"] = aws.StringValueSlice(conditions.NotEquals)
		}
		if v := aws.Int64Value(conditions.GreaterThan); v > 0 {
			criterion["greater_than"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.GreaterThanOrEqual); v > 0 {
			criterion["greater_than_or_equal"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.LessThan); v > 0 {
			criterion["less_than"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.LessThanOrEqual); v > 0 {
			criterion["less_than_or_equal"] = flattenConditionIntField(field, v)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []interface{}{
		map[string][]interface{}{
			"criterion": flatCriteria,
		},
	}
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
