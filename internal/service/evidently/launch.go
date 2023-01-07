package evidently

import (
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLaunch() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
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
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 160),
			},
			"execution": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ended_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"started_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"groups": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 160),
						},
						"feature": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
						"variation": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 127),
								validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
							),
						},
					},
				},
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metric_monitors": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_definition": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"event_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 1024),
											validation.StringIsJSON,
										),
										DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
										StateFunc: func(v interface{}) string {
											json, _ := structure.NormalizeJsonString(v)
											return json
										},
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"unit_label": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"value_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// case 1: User-defined string (old) is a name and is the suffix of API-returned string (new). Check non-empty old in resoure creation scenario
					// case 2: after setting API-returned string.  User-defined string (new) is suffix of API-returned string (old)
					return (strings.HasSuffix(new, old) && old != "") || strings.HasSuffix(old, new)
				},
			},
			"randomization_salt": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 127),
				// Default: set to the launch name if not specified
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == d.Get("name").(string) && new == ""
				},
			},
			"scheduled_splits_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"steps": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 6,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_weights": {
										Type:     schema.TypeMap,
										Required: true,
										ValidateDiagFunc: verify.ValidAllDiag(
											validation.MapKeyLenBetween(1, 127),
											validation.MapKeyMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
										),
										Elem: &schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntBetween(0, 100000),
										},
									},
									"segment_overrides": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 6,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"evaluation_order": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"segment": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 2048),
														validation.StringMatch(regexp.MustCompile(`(^[a-zA-Z0-9._-]*$)|(arn:[^:]*:[^:]*:[^:]*:[^:]*:segment/[a-zA-Z0-9._-]*)`), "name or arn of the segment"),
													),
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														// case 1: User-defined string (old) is a name and is the suffix of API-returned string (new). Check non-empty old in resoure creation scenario
														// case 2: after setting API-returned string.  User-defined string (new) is suffix of API-returned string (old)
														return (strings.HasSuffix(new, old) && old != "") || strings.HasSuffix(old, new)
													},
												},
												"weights": {
													Type:     schema.TypeMap,
													Required: true,
													ValidateDiagFunc: verify.ValidAllDiag(
														validation.MapKeyLenBetween(1, 127),
														validation.MapKeyMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeInt,
														ValidateFunc: validation.IntBetween(0, 100000),
													},
												},
											},
										},
									},
									"start_time": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}
