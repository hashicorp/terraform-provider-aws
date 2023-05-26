package glue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_data_quality_ruleset", name="Data Quality Ruleset")
// @Tags(identifierAttribute="arn")
func ResourceDataQualityRuleset() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: resourceDataQualityRulesetRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"last_modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"recommendation_run_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ruleset": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 65536),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// not found in SDK
						// "catalog_id": {
						// 	Type:         schema.TypeString,
						// 	Optional:     true,
						// 	ForceNew:     true,
						// 	ValidateFunc: validation.StringLenBetween(1, 255),
						// },
						"database_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"table_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
		},
	}
}

func resourceDataQualityRulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	name := d.Id()

	dataQualityRuleset, err := FindDataQualityRulesetByName(ctx, conn, name)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Data Quality Ruleset (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Data Quality Ruleset (%s): %s", d.Id(), err)
	}

	dataQualityRulesetArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dataQualityRuleset/%s", aws.StringValue(dataQualityRuleset.Name)),
	}.String()

	d.Set("arn", dataQualityRulesetArn)
	d.Set("created_on", dataQualityRuleset.CreatedOn.Format(time.RFC3339))
	d.Set("name", dataQualityRuleset.Name)
	d.Set("description", dataQualityRuleset.Description)
	d.Set("last_modified_on", dataQualityRuleset.CreatedOn.Format(time.RFC3339))
	d.Set("recommendation_run_id", dataQualityRuleset.RecommendationRunId)
	d.Set("ruleset", dataQualityRuleset.Ruleset)

	if err := d.Set("target_table", flattenTargetTable(dataQualityRuleset.TargetTable)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
	}

	return diags
}

func flattenTargetTable(apiObject *glue.DataQualityTargetTable) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"database_name": aws.StringValue(apiObject.DatabaseName),
		"table_name":    aws.StringValue(apiObject.TableName),
	}

	return []interface{}{tfMap}
}
