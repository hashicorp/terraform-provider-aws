// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
		CreateWithoutTimeout: resourceDataQualityRulesetCreate,
		ReadWithoutTimeout:   resourceDataQualityRulesetRead,
		UpdateWithoutTimeout: resourceDataQualityRulesetUpdate,
		DeleteWithoutTimeout: resourceDataQualityRulesetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"last_modified_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
						names.AttrCatalogID: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrDatabaseName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrTableName: {
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

func resourceDataQualityRulesetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)

	input := &glue.CreateDataQualityRulesetInput{
		Name:    aws.String(name),
		Ruleset: aws.String(d.Get("ruleset").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TargetTable = expandTargetTable(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateDataQualityRuleset(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Data Quality Ruleset (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDataQualityRulesetRead(ctx, d, meta)...)
}

func resourceDataQualityRulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

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
		Resource:  fmt.Sprintf("dataQualityRuleset/%s", aws.ToString(dataQualityRuleset.Name)),
	}.String()

	d.Set(names.AttrARN, dataQualityRulesetArn)
	d.Set("created_on", dataQualityRuleset.CreatedOn.Format(time.RFC3339))
	d.Set(names.AttrName, dataQualityRuleset.Name)
	d.Set(names.AttrDescription, dataQualityRuleset.Description)
	d.Set("last_modified_on", dataQualityRuleset.CreatedOn.Format(time.RFC3339))
	d.Set("recommendation_run_id", dataQualityRuleset.RecommendationRunId)
	d.Set("ruleset", dataQualityRuleset.Ruleset)

	if err := d.Set("target_table", flattenTargetTable(dataQualityRuleset.TargetTable)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
	}

	return diags
}

func resourceDataQualityRulesetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges(names.AttrDescription, "ruleset") {
		name := d.Id()

		input := &glue.UpdateDataQualityRulesetInput{
			Name: aws.String(name),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ruleset"); ok {
			input.Ruleset = aws.String(v.(string))
		}

		if _, err := conn.UpdateDataQualityRuleset(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Data Quality Ruleset (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataQualityRulesetRead(ctx, d, meta)...)
}

func resourceDataQualityRulesetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Glue Data Quality Ruleset: %s", d.Id())
	_, err := conn.DeleteDataQualityRuleset(ctx, &glue.DeleteDataQualityRulesetInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Data Quality Ruleset (%s): %s", d.Id(), err)
	}

	return diags
}

func expandTargetTable(tfMap map[string]interface{}) *awstypes.DataQualityTargetTable {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataQualityTargetTable{
		DatabaseName: aws.String(tfMap[names.AttrDatabaseName].(string)),
		TableName:    aws.String(tfMap[names.AttrTableName].(string)),
	}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	return apiObject
}

func flattenTargetTable(apiObject *awstypes.DataQualityTargetTable) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrDatabaseName: aws.ToString(apiObject.DatabaseName),
		names.AttrTableName:    aws.ToString(apiObject.TableName),
	}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
