// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		CreateWithoutTimeout: resourceDataQualityRulesetCreate,
		ReadWithoutTimeout:   resourceDataQualityRulesetRead,
		UpdateWithoutTimeout: resourceDataQualityRulesetUpdate,
		DeleteWithoutTimeout: resourceDataQualityRulesetDelete,
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
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
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

func resourceDataQualityRulesetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	name := d.Get("name").(string)

	input := &glue.CreateDataQualityRulesetInput{
		Name:    aws.String(name),
		Ruleset: aws.String(d.Get("ruleset").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TargetTable = expandTargetTable(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateDataQualityRulesetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Data Quality Ruleset (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDataQualityRulesetRead(ctx, d, meta)...)
}

func resourceDataQualityRulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

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

func resourceDataQualityRulesetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	if d.HasChanges("description", "ruleset") {
		name := d.Id()

		input := &glue.UpdateDataQualityRulesetInput{
			Name: aws.String(name),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ruleset"); ok {
			input.Ruleset = aws.String(v.(string))
		}

		if _, err := conn.UpdateDataQualityRulesetWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Data Quality Ruleset (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataQualityRulesetRead(ctx, d, meta)...)
}

func resourceDataQualityRulesetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	log.Printf("[DEBUG] Glue Data Quality Ruleset: %s", d.Id())
	_, err := conn.DeleteDataQualityRulesetWithContext(ctx, &glue.DeleteDataQualityRulesetInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Data Quality Ruleset (%s): %s", d.Id(), err)
	}

	return diags
}

func expandTargetTable(tfMap map[string]interface{}) *glue.DataQualityTargetTable {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.DataQualityTargetTable{
		DatabaseName: aws.String(tfMap["database_name"].(string)),
		TableName:    aws.String(tfMap["table_name"].(string)),
	}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	return apiObject
}

func flattenTargetTable(apiObject *glue.DataQualityTargetTable) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"database_name": aws.StringValue(apiObject.DatabaseName),
		"table_name":    aws.StringValue(apiObject.TableName),
	}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
