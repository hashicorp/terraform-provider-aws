// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_data_set", name="Data Set")
// @Testing(tagsTest=true)
// @Testing(tagsIdentifierAttribute="arn")
func dataSourceDataSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDataSetRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"column_groups":                 quicksightschema.DataSetColumnGroupsSchemaDataSourceSchema(),
				"column_level_permission_rules": quicksightschema.DataSetColumnLevelPermissionRulesSchemaDataSourceSchema(),
				"data_set_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"data_set_usage_configuration": quicksightschema.DataSetUsageConfigurationSchemaDataSourceSchema(),
				"field_folders":                quicksightschema.DataSetFieldFoldersSchemaDataSourceSchema(),
				"import_mode": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"logical_table_map": quicksightschema.DataSetLogicalTableMapSchemaDataSourceSchema(),
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrPermissions:                    quicksightschema.PermissionsDataSourceSchema(),
				"physical_table_map":                     quicksightschema.DataSetPhysicalTableMapSchemaDataSourceSchema(),
				"row_level_permission_data_set":          quicksightschema.DataSetRowLevelPermissionDataSetSchemaDataSourceSchema(),
				"row_level_permission_tag_configuration": quicksightschema.DataSetRowLevelPermissionTagConfigurationSchemaDataSourceSchema(),
				names.AttrTags:                           tftags.TagsSchemaComputed(),
				names.AttrTagsAll: {
					Type:       schema.TypeMap,
					Optional:   true,
					Computed:   true,
					Elem:       &schema.Schema{Type: schema.TypeString},
					Deprecated: "tags_all is deprecated. This argument will be removed in a future major version.",
				},
			}
		},
	}
}

func dataSourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dataSetID := d.Get("data_set_id").(string)
	id := dataSetCreateResourceID(awsAccountID, dataSetID)

	dataSet, err := findDataSetByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrARN, dataSet.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	if err := d.Set("column_groups", quicksightschema.FlattenColumnGroups(dataSet.ColumnGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_groups: %s", err)
	}
	if err := d.Set("column_level_permission_rules", quicksightschema.FlattenColumnLevelPermissionRules(dataSet.ColumnLevelPermissionRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting column_level_permission_rules: %s", err)
	}
	d.Set("data_set_id", dataSet.DataSetId)
	if err := d.Set("data_set_usage_configuration", quicksightschema.FlattenDataSetUsageConfiguration(dataSet.DataSetUsageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_set_usage_configuration: %s", err)
	}
	if err := d.Set("field_folders", quicksightschema.FlattenFieldFolders(dataSet.FieldFolders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting field_folders: %s", err)
	}
	d.Set("import_mode", dataSet.ImportMode)
	if err := d.Set("logical_table_map", quicksightschema.FlattenLogicalTableMap(dataSet.LogicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logical_table_map: %s", err)
	}
	d.Set(names.AttrName, dataSet.Name)
	if err := d.Set("physical_table_map", quicksightschema.FlattenPhysicalTableMap(dataSet.PhysicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_table_map: %s", err)
	}
	if err := d.Set("row_level_permission_data_set", quicksightschema.FlattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_data_set: %s", err)
	}
	if err := d.Set("row_level_permission_tag_configuration", quicksightschema.FlattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_tag_configuration: %s", err)
	}

	// Cannot use transparent tagging because it has to handle `tags_all` as well
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for QuickSight Data Set (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	permissions, err := findDataSetPermissionsByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}
