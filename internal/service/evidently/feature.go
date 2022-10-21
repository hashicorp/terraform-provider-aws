package evidently

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFeature() *schema.Resource {
	return &schema.Resource{

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
