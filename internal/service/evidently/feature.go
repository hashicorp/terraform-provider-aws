package evidently

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFeature() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFeatureCreate,
		ReadWithoutTimeout:   resourceFeatureRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_variation": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 160),
			},
			"entity_overrides": {
				Type:     schema.TypeMap,
				Optional: true,
				ValidateDiagFunc: verify.ValidAllDiag(
					validation.MapKeyLenBetween(1, 512),
					validation.MapValueLenBetween(1, 127),
					validation.MapValueMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"evaluation_rules": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"evaluation_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(cloudwatchevidently.FeatureEvaluationStrategy_Values(), false),
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringMatch(regexp.MustCompile(`(^[a-zA-Z0-9._-]*$)|(arn:[^:]*:[^:]*:[^:]*:[^:]*:project/[a-zA-Z0-9._-]*)`), "name or arn of the project"),
				),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"value_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"variations": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
						"value": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bool_value": {
										Type:         nullable.TypeNullableBool,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableBool,
										// unable to index parent list
										// ConflictsWith: []string{"double_value", "long_value", "string_value"},
									},
									"double_value": {
										Type:     nullable.TypeNullableFloat,
										Optional: true,
										// unable to index parent list
										// ConflictsWith: []string{"bool_value", "long_value", "string_value"},
									},
									"long_value": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(-9007199254740991, 9007199254740991),
										// unable to index parent list
										// ConflictsWith: []string{"bool_value", "double_value", "string_value"},
									},
									"string_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 512),
										// unable to index parent list
										// ConflictsWith: []string{"bool_value", "double_value", "long_value"},
									},
								},
							},
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFeatureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	project := d.Get("project").(string)

	input := &cloudwatchevidently.CreateFeatureInput{
		Name:       aws.String(name),
		Project:    aws.String(project),
		Variations: expandVariations(d.Get("variations").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("default_variation"); ok {
		input.DefaultVariation = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v := d.Get("entity_overrides").(map[string]interface{}); len(v) > 0 {
		input.EntityOverrides = flex.ExpandStringMap(v)
	}

	if v, ok := d.GetOk("evaluation_strategy"); ok {
		input.EvaluationStrategy = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating CloudWatch Evidently Feature: %s", input)
	output, err := conn.CreateFeatureWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Evidently Feature (%s) for Project (%s): %s", name, project, err)
	}

	// the GetFeature API call uses the Feature name and Project ARN
	// concat Feature name and Project Name or ARN to be used in Read for imports
	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.Feature.Name), project))

	return resourceFeatureRead(ctx, d, meta)
}

func resourceFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	featureName, projectNameOrARN, err := FeatureParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	feature, err := FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Feature (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Evidently Feature (%s) for Project (%s): %s", featureName, projectNameOrARN, err)
	}

	if err := d.Set("evaluation_rules", flattenEvaluationRules(feature.EvaluationRules)); err != nil {
		return diag.Errorf("setting evaluation_rules: %s", err)
	}

	if err := d.Set("variations", flattenVariations(feature.Variations)); err != nil {
		return diag.Errorf("setting variations: %s", err)
	}

	d.Set("arn", feature.Arn)
	d.Set("created_time", aws.TimeValue(feature.CreatedTime).Format(time.RFC3339))
	d.Set("default_variation", feature.DefaultVariation)
	d.Set("description", feature.Description)
	d.Set("entity_overrides", aws.StringValueMap(feature.EntityOverrides))
	d.Set("evaluation_strategy", feature.EvaluationStrategy)
	d.Set("last_updated_time", aws.TimeValue(feature.LastUpdatedTime).Format(time.RFC3339))
	d.Set("name", feature.Name)
	d.Set("project", projectNameOrARN)
	d.Set("status", feature.Status)
	d.Set("value_type", feature.ValueType)

	tags := KeyValueTags(feature.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func FeatureParseID(id string) (string, string, error) {
	featureName, projectNameOrARN, _ := strings.Cut(id, ":")

	if featureName == "" || projectNameOrARN == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected featureName:projectNameOrARN", id)
	}

	return featureName, projectNameOrARN, nil
}

func expandVariations(variations []interface{}) []*cloudwatchevidently.VariationConfig {
	if len(variations) == 0 {
		return nil
	}

	variationsFormatted := make([]*cloudwatchevidently.VariationConfig, len(variations))

	for i, variation := range variations {
		variationsFormatted[i] = expandVariation(variation.(map[string]interface{}))
	}

	return variationsFormatted
}

func expandVariation(variation map[string]interface{}) *cloudwatchevidently.VariationConfig {
	return &cloudwatchevidently.VariationConfig{
		Name:  aws.String(variation["name"].(string)),
		Value: expandValue(variation["value"].([]interface{})),
	}
}

func expandValue(value []interface{}) *cloudwatchevidently.VariableValue {
	if len(value) == 0 || value[0] == nil {
		return nil
	}

	tfMap, ok := value[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &cloudwatchevidently.VariableValue{}

	// Only one of these values can be set at a time
	if val, null, _ := nullable.Bool(tfMap["bool_value"].(string)).Value(); !null {
		result.BoolValue = aws.Bool(val)
	} else if v, null, _ := nullable.Int(tfMap["long_value"].(string)).Value(); !null {
		result.LongValue = aws.Int64(v)
	} else if v, null, _ := nullable.Float(tfMap["double_value"].(string)).Value(); !null {
		result.DoubleValue = aws.Float64(float64(v))
	} else if v, ok := tfMap["string_value"].(string); ok {
		result.StringValue = aws.String(v)
	}
	return result
}

func flattenVariations(apiObject []*cloudwatchevidently.Variation) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	variationsFormatted := []interface{}{}
	for _, variation := range apiObject {
		variationFormatted := map[string]interface{}{
			"name":  aws.StringValue(variation.Name),
			"value": flattenValue(variation.Value),
		}
		variationsFormatted = append(variationsFormatted, variationFormatted)
	}
	return variationsFormatted
}

func flattenValue(apiObject *cloudwatchevidently.VariableValue) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	// only one of these values should be set at a time
	if v := apiObject.BoolValue; v != nil {
		m["bool_value"] = strconv.FormatBool(aws.BoolValue(v))
	} else if v := apiObject.LongValue; v != nil {
		m["long_value"] = strconv.FormatInt(aws.Int64Value(v), 10)
	} else if v := apiObject.DoubleValue; v != nil {
		m["double_value"] = strconv.FormatFloat(aws.Float64Value(v), 'f', -1, 64)
	} else {
		m["string_value"] = aws.StringValue(apiObject.StringValue)
	}

	return []interface{}{m}
}

func flattenEvaluationRules(apiObject []*cloudwatchevidently.EvaluationRule) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	evaluationRulesFormatted := []interface{}{}
	for _, evaluationRule := range apiObject {
		evaluationRuleFormatted := map[string]interface{}{
			"name": aws.StringValue(evaluationRule.Name),
			"type": aws.StringValue(evaluationRule.Type),
		}
		evaluationRulesFormatted = append(evaluationRulesFormatted, evaluationRuleFormatted)
	}
	return evaluationRulesFormatted
}
