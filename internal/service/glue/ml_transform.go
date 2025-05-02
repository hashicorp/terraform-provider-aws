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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_ml_transform", name="ML Transform")
// @Tags(identifierAttribute="arn")
func resourceMLTransform() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMLTransformCreate,
		ReadWithoutTimeout:   resourceMLTransformRead,
		UpdateWithoutTimeout: resourceMLTransformUpdate,
		DeleteWithoutTimeout: resourceMLTransformDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"input_record_tables": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrTableName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrParameters: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"find_matches_parameters": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"accuracy_cost_trade_off": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtMost(1.0),
									},
									"enforce_provided_labels": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"precision_recall_trade_off": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtMost(1.0),
									},
									"primary_key_column_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"transform_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TransformType](),
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrMaxCapacity: {
				Type:          schema.TypeFloat,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
				ValidateFunc:  validation.FloatBetween(2, 100),
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2880,
			},
			"worker_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{names.AttrMaxCapacity},
				ValidateDiagFunc: enum.Validate[awstypes.WorkerType](),
				RequiredWith:     []string{"number_of_workers"},
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{names.AttrMaxCapacity},
				ValidateFunc:  validation.IntAtLeast(1),
				RequiredWith:  []string{"worker_type"},
			},
			"label_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrSchema: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMLTransformCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.CreateMLTransformInput{
		Name:              aws.String(d.Get(names.AttrName).(string)),
		Role:              aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:              getTagsIn(ctx),
		Timeout:           aws.Int32(int32(d.Get(names.AttrTimeout).(int))),
		InputRecordTables: expandMLTransformInputRecordTables(d.Get("input_record_tables").([]any)),
		Parameters:        expandMLTransformParameters(d.Get(names.AttrParameters).([]any)),
	}

	if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = awstypes.WorkerType(v.(string))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue ML Transform: %+v", input)

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidInputException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateMLTransform(ctx, input)
	}, "Unable to assume role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue ML Transform: %s", err)
	}

	output := outputRaw.(*glue.CreateMLTransformOutput)

	d.SetId(aws.ToString(output.TransformId))

	return append(diags, resourceMLTransformRead(ctx, d, meta)...)
}

func resourceMLTransformRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.GetMLTransformInput{
		TransformId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue ML Transform: %+v", input)
	output, err := conn.GetMLTransform(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Glue ML Transform (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue ML Transform (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue ML Transform (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	log.Printf("[DEBUG] setting Glue ML Transform: %#v", output)

	mlTransformArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("mlTransform/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, mlTransformArn)

	d.Set(names.AttrDescription, output.Description)
	d.Set("glue_version", output.GlueVersion)
	d.Set(names.AttrMaxCapacity, output.MaxCapacity)
	d.Set("max_retries", output.MaxRetries)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrRoleARN, output.Role)
	d.Set(names.AttrTimeout, output.Timeout)
	d.Set("worker_type", output.WorkerType)
	d.Set("number_of_workers", output.NumberOfWorkers)
	d.Set("label_count", output.LabelCount)

	if err := d.Set("input_record_tables", flattenMLTransformInputRecordTables(output.InputRecordTables)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_record_tables: %s", err)
	}

	if err := d.Set(names.AttrParameters, flattenMLTransformParameters(output.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}

	if err := d.Set(names.AttrSchema, flattenMLTransformSchemaColumns(output.Schema)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schema: %s", err)
	}

	return diags
}

func resourceMLTransformUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges(names.AttrDescription, "glue_version", names.AttrMaxCapacity, "max_retries", "number_of_workers",
		names.AttrRoleARN, names.AttrTimeout, "worker_type", names.AttrParameters) {
		input := &glue.UpdateMLTransformInput{
			TransformId: aws.String(d.Id()),
			Role:        aws.String(d.Get(names.AttrRoleARN).(string)),
			Timeout:     aws.Int32(int32(d.Get(names.AttrTimeout).(int))),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			input.WorkerType = awstypes.WorkerType(v.(string))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			input.MaxRetries = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			input.NumberOfWorkers = aws.Int32(int32(v.(int)))
		} else {
			if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
				input.MaxCapacity = aws.Float64(v.(float64))
			}
		}

		if v, ok := d.GetOk("glue_version"); ok {
			input.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok {
			input.Parameters = expandMLTransformParameters(v.([]any))
		}

		log.Printf("[DEBUG] Updating Glue ML Transform: %+v", input)
		_, err := conn.UpdateMLTransform(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue ML Transform (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMLTransformRead(ctx, d, meta)...)
}

func resourceMLTransformDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue ML Transform: %s", d.Id())
	input := glue.DeleteMLTransformInput{
		TransformId: aws.String(d.Id()),
	}
	_, err := conn.DeleteMLTransform(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue ML Transform (%s): %s", d.Id(), err)
	}

	if _, err := waitMLTransformDeleted(ctx, conn, d.Id()); err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue ML Transform (%s) to be Deleted: %s", d.Id(), err)
	}

	return diags
}

// statusMLTransform fetches the MLTransform and its Status
func statusMLTransform(ctx context.Context, conn *glue.Client, transformId string) retry.StateRefreshFunc {
	const (
		mlTransformStatusUnknown = "Unknown"
	)
	return func() (any, string, error) {
		input := &glue.GetMLTransformInput{
			TransformId: aws.String(transformId),
		}

		output, err := conn.GetMLTransform(ctx, input)

		if err != nil {
			return nil, mlTransformStatusUnknown, err
		}

		if output == nil {
			return output, mlTransformStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// waitMLTransformDeleted waits for an MLTransform to return Deleted
func waitMLTransformDeleted(ctx context.Context, conn *glue.Client, transformId string) (*glue.GetMLTransformOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransformStatusTypeNotReady, awstypes.TransformStatusTypeReady, awstypes.TransformStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusMLTransform(ctx, conn, transformId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetMLTransformOutput); ok {
		return output, err
	}

	return nil, err
}

func expandMLTransformInputRecordTables(l []any) []awstypes.GlueTable {
	var tables []awstypes.GlueTable

	for _, mRaw := range l {
		m := mRaw.(map[string]any)

		table := awstypes.GlueTable{}

		if v, ok := m[names.AttrTableName].(string); ok {
			table.TableName = aws.String(v)
		}

		if v, ok := m[names.AttrDatabaseName].(string); ok {
			table.DatabaseName = aws.String(v)
		}

		if v, ok := m["connection_name"].(string); ok && v != "" {
			table.ConnectionName = aws.String(v)
		}

		if v, ok := m[names.AttrCatalogID].(string); ok && v != "" {
			table.CatalogId = aws.String(v)
		}

		tables = append(tables, table)
	}

	return tables
}

func flattenMLTransformInputRecordTables(tables []awstypes.GlueTable) []any {
	l := []any{}

	for _, table := range tables {
		m := map[string]any{
			names.AttrTableName:    aws.ToString(table.TableName),
			names.AttrDatabaseName: aws.ToString(table.DatabaseName),
		}

		if table.ConnectionName != nil {
			m["connection_name"] = aws.ToString(table.ConnectionName)
		}

		if table.CatalogId != nil {
			m[names.AttrCatalogID] = aws.ToString(table.CatalogId)
		}

		l = append(l, m)
	}

	return l
}

func expandMLTransformParameters(l []any) *awstypes.TransformParameters {
	m := l[0].(map[string]any)

	param := &awstypes.TransformParameters{
		TransformType: awstypes.TransformType(m["transform_type"].(string)),
	}

	if v, ok := m["find_matches_parameters"]; ok && len(v.([]any)) > 0 {
		param.FindMatchesParameters = expandMLTransformFindMatchesParameters(v.([]any))
	}

	return param
}

func flattenMLTransformParameters(parameters *awstypes.TransformParameters) []map[string]any {
	if parameters == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"transform_type": string(parameters.TransformType),
	}

	if parameters.FindMatchesParameters != nil {
		m["find_matches_parameters"] = flattenMLTransformFindMatchesParameters(parameters.FindMatchesParameters)
	}

	return []map[string]any{m}
}

func expandMLTransformFindMatchesParameters(l []any) *awstypes.FindMatchesParameters {
	m := l[0].(map[string]any)

	param := &awstypes.FindMatchesParameters{}

	if v, ok := m["accuracy_cost_trade_off"]; ok {
		param.AccuracyCostTradeoff = aws.Float64(v.(float64))
	}

	if v, ok := m["precision_recall_trade_off"]; ok {
		param.PrecisionRecallTradeoff = aws.Float64(v.(float64))
	}

	if v, ok := m["enforce_provided_labels"]; ok {
		param.EnforceProvidedLabels = aws.Bool(v.(bool))
	}

	if v, ok := m["primary_key_column_name"]; ok && v != "" {
		param.PrimaryKeyColumnName = aws.String(v.(string))
	}

	return param
}

func flattenMLTransformFindMatchesParameters(parameters *awstypes.FindMatchesParameters) []map[string]any {
	if parameters == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if parameters.PrimaryKeyColumnName != nil {
		m["primary_key_column_name"] = aws.ToString(parameters.PrimaryKeyColumnName)
	}

	if parameters.EnforceProvidedLabels != nil {
		m["enforce_provided_labels"] = aws.ToBool(parameters.EnforceProvidedLabels)
	}

	if parameters.AccuracyCostTradeoff != nil {
		m["accuracy_cost_trade_off"] = aws.ToFloat64(parameters.AccuracyCostTradeoff)
	}

	if parameters.PrimaryKeyColumnName != nil {
		m["precision_recall_trade_off"] = aws.ToFloat64(parameters.PrecisionRecallTradeoff)
	}

	return []map[string]any{m}
}

func flattenMLTransformSchemaColumns(schemaCols []awstypes.SchemaColumn) []any {
	l := []any{}

	for _, schemaCol := range schemaCols {
		m := map[string]any{
			names.AttrName: aws.ToString(schemaCol.Name),
			"data_type":    aws.ToString(schemaCol.DataType),
		}

		l = append(l, m)
	}

	return l
}
