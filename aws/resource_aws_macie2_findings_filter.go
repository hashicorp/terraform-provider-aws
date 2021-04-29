package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsMacie2FindingsFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2FindingsFilterCreate,
		ReadWithoutTimeout:   resourceMacie2FindingsFilterRead,
		UpdateWithoutTimeout: resourceMacie2FindingsFilterUpdate,
		DeleteWithoutTimeout: resourceMacie2FindingsFilterDelete,
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
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:     schema.TypeString,
										Required: true,
									},
									"eq_exact_match": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"eq": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"neq": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"lt": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"lte": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"gt": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"gte": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(macie2.FindingsFilterAction_Values(), false),
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2FindingsFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.CreateFindingsFilterInput{
		ClientToken:     aws.String(resource.UniqueId()),
		FindingCriteria: expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{})),
		Name:            aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
		Action:          aws.String(d.Get("action").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("position"); ok {
		input.SetPosition(int64(v.(int)))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().AppsyncTags())
	}

	var err error
	var output macie2.CreateFindingsFilterOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		resp, err := conn.CreateFindingsFilterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}
		output = *resp

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateFindingsFilterWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie2 FindingsFilter: %w", err))
	}

	d.SetId(aws.StringValue(output.Id))

	return resourceMacie2FindingsFilterRead(ctx, d, meta)
}

func resourceMacie2FindingsFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetFindingsFilterWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			log.Printf("[WARN] Macie2 FindingsFilter does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Macie2 FindingsFilter (%s): %w", d.Id(), err))
	}

	if err = d.Set("finding_criteria", flattenFindingCriteriaFindingsFilter(resp.FindingCriteria)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie2 FindingsFilter (%s): %w", "finding_criteria", d.Id(), err))
	}
	d.Set("name", aws.StringValue(resp.Name))
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set("description", aws.StringValue(resp.Description))
	d.Set("action", aws.StringValue(resp.Action))
	d.Set("position", aws.Int64Value(resp.Position))
	if err = d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie2 FindingsFilter (%s): %w", "tags", d.Id(), err))
	}
	d.Set("arn", aws.StringValue(resp.Arn))

	return nil
}

func resourceMacie2FindingsFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("finding_criteria") {
		input.SetFindingCriteria(expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{})))
	}
	if d.HasChange("name") {
		input.SetName(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	}
	if d.HasChange("name_prefix") {
		input.SetName(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	}
	if d.HasChange("description") {
		input.SetDescription(d.Get("description").(string))
	}
	if d.HasChange("action") {
		input.SetAction(d.Get("action").(string))
	}
	if d.HasChange("position") {
		input.SetPosition(int64(d.Get("position").(int)))
	}

	_, err := conn.UpdateFindingsFilterWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Macie2 FindingsFilter (%s): %w", d.Id(), err))
	}

	return resourceMacie2FindingsFilterRead(ctx, d, meta)
}

func resourceMacie2FindingsFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteFindingsFilterWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie2 FindingsFilter (%s): %w", d.Id(), err))
	}
	return nil
}

func expandFindingCriteriaFilter(findingCriterias []interface{}) *macie2.FindingCriteria {
	if len(findingCriterias) == 0 {
		return nil
	}

	criteria := map[string]*macie2.CriterionAdditionalProperties{}
	findingCriteria := findingCriterias[0].(map[string]interface{})
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	for _, criterion := range inputFindingCriteria {
		crit := criterion.(map[string]interface{})
		field := crit["field"].(string)
		conditional := macie2.CriterionAdditionalProperties{}

		if v, ok := crit["eq"].([]interface{}); ok && len(v) > 0 {
			foo := make([]*string, len(v))
			for i := range v {
				s := v[i].(string)
				foo[i] = &s
			}
			conditional.Eq = foo
		}
		if v, ok := crit["neq"].([]interface{}); ok && len(v) > 0 {
			foo := make([]*string, len(v))
			for i := range v {
				s := v[i].(string)
				foo[i] = &s
			}
			conditional.Neq = foo
		}
		if v, ok := crit["eq_exact_match"].([]interface{}); ok && len(v) > 0 {
			foo := make([]*string, len(v))
			for i := range v {
				s := v[i].(string)
				foo[i] = &s
			}
			conditional.EqExactMatch = foo
		}
		if v, ok := crit["lt"].(int); ok && v > 0 {
			conditional.Lt = aws.Int64(int64(v))
		}
		if v, ok := crit["lte"].(int); ok && v > 0 {
			conditional.Lte = aws.Int64(int64(v))
		}
		if v, ok := crit["gt"].(int); ok && v > 0 {
			conditional.Gt = aws.Int64(int64(v))
		}
		if v, ok := crit["gte"].(int); ok && v > 0 {
			conditional.Gte = aws.Int64(int64(v))
		}
		criteria[field] = &conditional
	}

	return &macie2.FindingCriteria{Criterion: criteria}
}

func flattenFindingCriteriaFindingsFilter(findingCriteria *macie2.FindingCriteria) []interface{} {
	if findingCriteria == nil {
		return nil
	}

	var flatCriteria []interface{}

	for field, conditions := range findingCriteria.Criterion {
		criterion := map[string]interface{}{
			"field": field,
		}
		if len(conditions.Eq) > 0 {
			criterion["eq"] = aws.StringValueSlice(conditions.Eq)
		}
		if len(conditions.Neq) > 0 {
			criterion["neq"] = aws.StringValueSlice(conditions.Neq)
		}
		if len(conditions.EqExactMatch) > 0 {
			criterion["eq_exact_match"] = aws.StringValueSlice(conditions.EqExactMatch)
		}
		if v := aws.Int64Value(conditions.Lt); v > 0 {
			criterion["lt"] = aws.Int64Value(conditions.Lt)
		}
		if v := aws.Int64Value(conditions.Lt); v > 0 {
			criterion["lte"] = aws.Int64Value(conditions.Lte)
		}
		if v := aws.Int64Value(conditions.Lt); v > 0 {
			criterion["gt"] = aws.Int64Value(conditions.Gt)
		}
		if v := aws.Int64Value(conditions.Lt); v > 0 {
			criterion["gte"] = aws.Int64Value(conditions.Gte)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []interface{}{
		map[string][]interface{}{
			"criterion": flatCriteria,
		},
	}
}
