// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cleanrooms

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cleanrooms_configured_table", name="Configured Table")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
// @IdentityAttribute("id")
// @Testing(preIdentityVersion="v6.14.1")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cleanrooms;cleanrooms.GetConfiguredTableOutput")
func resourceConfiguredTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfiguredTableCreate,
		ReadWithoutTimeout:   resourceConfiguredTableRead,
		UpdateWithoutTimeout: resourceConfiguredTableUpdate,
		DeleteWithoutTimeout: resourceConfiguredTableDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"allowed_columns": {
					Type:     schema.TypeSet,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					MinItems: 1,
					MaxItems: 225,
				},
				"analysis_method": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.AnalysisMethod](),
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrCreateTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"selected_analysis_methods": {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 2,
					MaxItems: 2,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[awstypes.SelectedAnalysisMethod](),
					},
				},
				"table_reference": {
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabaseName: {
								Type:          schema.TypeString,
								Optional:      true,
								ForceNew:      true,
								RequiredWith:  []string{"table_reference.0.table_name"},
								AtLeastOneOf:  []string{"table_reference.0.database_name", "table_reference.0.athena", "table_reference.0.snowflake"},
								ConflictsWith: []string{"table_reference.0.athena", "table_reference.0.snowflake"},
							},
							names.AttrTableName: {
								Type:          schema.TypeString,
								Optional:      true,
								ForceNew:      true,
								RequiredWith:  []string{"table_reference.0.database_name"},
								ConflictsWith: []string{"table_reference.0.athena", "table_reference.0.snowflake"},
							},
							names.AttrRegion: {
								Type:             schema.TypeString,
								Optional:         true,
								ForceNew:         true,
								ConflictsWith:    []string{"table_reference.0.athena", "table_reference.0.snowflake"},
								ValidateDiagFunc: enum.Validate[awstypes.CommercialRegion](),
							},
							"athena": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								ConflictsWith: []string{
									"table_reference.0.database_name",
									"table_reference.0.table_name",
									"table_reference.0.snowflake",
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabaseName: {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										names.AttrTableName: {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"workgroup": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"catalog_name": {
											Type:     schema.TypeString,
											Optional: true,
											ForceNew: true,
										},
										"output_location": {
											Type:     schema.TypeString,
											Optional: true,
											ForceNew: true,
										},
										names.AttrRegion: {
											Type:             schema.TypeString,
											Optional:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CommercialRegion](),
										},
									},
								},
							},
							"snowflake": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								ConflictsWith: []string{
									"table_reference.0.database_name",
									"table_reference.0.table_name",
									"table_reference.0.athena",
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"account_identifier": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										names.AttrDatabaseName: {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"schema_name": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"secret_arn": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										names.AttrTableName: {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"table_schema": {
											Type:     schema.TypeList,
											Required: true,
											ForceNew: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"v1": {
														Type:     schema.TypeList,
														Required: true,
														ForceNew: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"column_name": {
																	Type:     schema.TypeString,
																	Required: true,
																	ForceNew: true,
																},
																"column_type": {
																	Type:     schema.TypeString,
																	Required: true,
																	ForceNew: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"update_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

const (
	ResNameConfiguredTable = "Configured Table"
)

func resourceConfiguredTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	input := cleanrooms.CreateConfiguredTableInput{
		Name:           aws.String(d.Get(names.AttrName).(string)),
		AllowedColumns: flex.ExpandStringValueSet(d.Get("allowed_columns").(*schema.Set)),
		AnalysisMethod: awstypes.AnalysisMethod(d.Get("analysis_method").(string)),
		TableReference: expandTableReference(d.Get("table_reference").([]any)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("selected_analysis_methods"); ok {
		input.SelectedAnalysisMethods = flex.ExpandStringyValueSet[awstypes.SelectedAnalysisMethod](v.(*schema.Set))
	}

	out, err := conn.CreateConfiguredTable(ctx, &input)
	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTable, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.ConfiguredTable == nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTable, d.Get(names.AttrName).(string), errors.New("empty output"))
	}
	d.SetId(aws.ToString(out.ConfiguredTable.Id))

	return append(diags, resourceConfiguredTableRead(ctx, d, meta)...)
}

func resourceConfiguredTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	out, err := findConfiguredTableByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Clean Rooms Configured Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionReading, ResNameConfiguredTable, d.Id(), err)
	}

	resourceConfiguredTableFlatten(ctx, d, out)

	return diags
}

func resourceConfiguredTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := cleanrooms.UpdateConfiguredTableInput{
			ConfiguredTableIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges("allowed_columns") {
			input.AllowedColumns = flex.ExpandStringValueSet(d.Get("allowed_columns").(*schema.Set))
		}

		if d.HasChanges("analysis_method", "selected_analysis_methods") {
			input.AnalysisMethod = awstypes.AnalysisMethod(d.Get("analysis_method").(string))
			if input.AnalysisMethod == awstypes.AnalysisMethodMultiple {
				if v, ok := d.GetOk("selected_analysis_methods"); ok {
					input.SelectedAnalysisMethods = flex.ExpandStringyValueSet[awstypes.SelectedAnalysisMethod](v.(*schema.Set))
				}
			}
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateConfiguredTable(ctx, &input)
		if err != nil {
			return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionUpdating, ResNameConfiguredTable, d.Id(), err)
		}
	}

	return append(diags, resourceConfiguredTableRead(ctx, d, meta)...)
}

func resourceConfiguredTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	log.Printf("[INFO] Deleting Clean Rooms Configured Table %s", d.Id())
	input := cleanrooms.DeleteConfiguredTableInput{
		ConfiguredTableIdentifier: aws.String(d.Id()),
	}

	_, err := conn.DeleteConfiguredTable(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionDeleting, ResNameConfiguredTable, d.Id(), err)
	}

	return diags
}

func findConfiguredTableByID(ctx context.Context, conn *cleanrooms.Client, id string) (*cleanrooms.GetConfiguredTableOutput, error) {
	input := cleanrooms.GetConfiguredTableInput{
		ConfiguredTableIdentifier: aws.String(id),
	}

	out, err := conn.GetConfiguredTable(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ConfiguredTable == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func expandTableReference(data []any) awstypes.TableReference {
	tableReference := data[0].(map[string]any)

	if v, ok := tableReference["athena"].([]any); ok && len(v) > 0 {
		return expandAthenaTableReference(v[0].(map[string]any))
	}

	if v, ok := tableReference["snowflake"].([]any); ok && len(v) > 0 {
		return expandSnowflakeTableReference(v[0].(map[string]any))
	}

	ref := awstypes.GlueTableReference{
		DatabaseName: aws.String(tableReference[names.AttrDatabaseName].(string)),
		TableName:    aws.String(tableReference[names.AttrTableName].(string)),
	}
	if v, ok := tableReference[names.AttrRegion].(string); ok && v != "" {
		ref.Region = awstypes.CommercialRegion(v)
	}
	return &awstypes.TableReferenceMemberGlue{Value: ref}
}

func expandAthenaTableReference(m map[string]any) awstypes.TableReference {
	ref := awstypes.AthenaTableReference{
		DatabaseName: aws.String(m[names.AttrDatabaseName].(string)),
		TableName:    aws.String(m[names.AttrTableName].(string)),
		WorkGroup:    aws.String(m["workgroup"].(string)),
	}
	if v, ok := m["catalog_name"].(string); ok && v != "" {
		ref.CatalogName = aws.String(v)
	}
	if v, ok := m["output_location"].(string); ok && v != "" {
		ref.OutputLocation = aws.String(v)
	}
	if v, ok := m[names.AttrRegion].(string); ok && v != "" {
		ref.Region = awstypes.CommercialRegion(v)
	}
	return &awstypes.TableReferenceMemberAthena{Value: ref}
}

func expandSnowflakeTableReference(m map[string]any) awstypes.TableReference {
	ref := awstypes.SnowflakeTableReference{
		AccountIdentifier: aws.String(m["account_identifier"].(string)),
		DatabaseName:      aws.String(m[names.AttrDatabaseName].(string)),
		SchemaName:        aws.String(m["schema_name"].(string)),
		SecretArn:         aws.String(m["secret_arn"].(string)),
		TableName:         aws.String(m[names.AttrTableName].(string)),
		TableSchema:       expandSnowflakeTableSchema(m["table_schema"].([]any)),
	}
	return &awstypes.TableReferenceMemberSnowflake{Value: ref}
}

func expandSnowflakeTableSchema(data []any) awstypes.SnowflakeTableSchema {
	schemaMap := data[0].(map[string]any)
	v1Raw := schemaMap["v1"].([]any)
	cols := make([]awstypes.SnowflakeTableSchemaV1, len(v1Raw))
	for i, col := range v1Raw {
		c := col.(map[string]any)
		cols[i] = awstypes.SnowflakeTableSchemaV1{
			ColumnName: aws.String(c["column_name"].(string)),
			ColumnType: aws.String(c["column_type"].(string)),
		}
	}
	return &awstypes.SnowflakeTableSchemaMemberV1{Value: cols}
}

func flattenTableReference(tableReference awstypes.TableReference) []any {
	switch v := tableReference.(type) {
	case *awstypes.TableReferenceMemberGlue:
		return []any{map[string]any{
			names.AttrDatabaseName: aws.ToString(v.Value.DatabaseName),
			names.AttrTableName:    aws.ToString(v.Value.TableName),
			names.AttrRegion:       string(v.Value.Region),
		}}
	case *awstypes.TableReferenceMemberAthena:
		inner := map[string]any{
			names.AttrDatabaseName: aws.ToString(v.Value.DatabaseName),
			names.AttrTableName:    aws.ToString(v.Value.TableName),
			"workgroup":            aws.ToString(v.Value.WorkGroup),
			"catalog_name":         aws.ToString(v.Value.CatalogName),
			"output_location":      aws.ToString(v.Value.OutputLocation),
			names.AttrRegion:       string(v.Value.Region),
		}
		return []any{map[string]any{"athena": []any{inner}}}
	case *awstypes.TableReferenceMemberSnowflake:
		inner := map[string]any{
			"account_identifier":   aws.ToString(v.Value.AccountIdentifier),
			names.AttrDatabaseName: aws.ToString(v.Value.DatabaseName),
			"schema_name":          aws.ToString(v.Value.SchemaName),
			"secret_arn":           aws.ToString(v.Value.SecretArn),
			names.AttrTableName:    aws.ToString(v.Value.TableName),
			"table_schema":         flattenSnowflakeTableSchema(v.Value.TableSchema),
		}
		return []any{map[string]any{"snowflake": []any{inner}}}
	default:
		return nil
	}
}

func flattenSnowflakeTableSchema(schema awstypes.SnowflakeTableSchema) []any {
	switch v := schema.(type) {
	case *awstypes.SnowflakeTableSchemaMemberV1:
		cols := make([]any, len(v.Value))
		for i, col := range v.Value {
			cols[i] = map[string]any{
				"column_name": aws.ToString(col.ColumnName),
				"column_type": aws.ToString(col.ColumnType),
			}
		}
		return []any{map[string]any{"v1": cols}}
	default:
		return nil
	}
}

func resourceConfiguredTableFlatten(_ context.Context, d *schema.ResourceData, out *cleanrooms.GetConfiguredTableOutput) {
	configuredTable := out.ConfiguredTable
	d.Set(names.AttrARN, configuredTable.Arn)
	d.Set(names.AttrName, configuredTable.Name)
	d.Set(names.AttrDescription, configuredTable.Description)
	d.Set("allowed_columns", configuredTable.AllowedColumns)
	d.Set("analysis_method", configuredTable.AnalysisMethod)
	d.Set("selected_analysis_methods", configuredTable.SelectedAnalysisMethods)
	d.Set(names.AttrCreateTime, configuredTable.CreateTime.String())
	d.Set("update_time", configuredTable.UpdateTime.String())
	d.Set("table_reference", flattenTableReference(configuredTable.TableReference))
}
