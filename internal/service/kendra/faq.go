package kendra

import (
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFaq() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(faqFileFormatValues(types.FaqFileFormat("").Values()...), false),
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"language_code": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z-]*`),
						"Must have alphanumeric characters or hyphens.",
					),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_path": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

// Helpers added. Could be generated or somehow use go 1.18 generics?
func faqFileFormatValues(input ...types.FaqFileFormat) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func FaqStatusValues(input ...types.FaqStatus) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	output = append(output, "PENDING_CREATION")
	output = append(output, "PENDING_DELETION")

	return output
}
