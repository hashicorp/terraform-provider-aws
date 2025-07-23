// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_data_set", name="Data Set")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.DataSet")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceDataSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSetCreate,
		ReadWithoutTimeout:   resourceDataSetRead,
		UpdateWithoutTimeout: resourceDataSetUpdate,
		DeleteWithoutTimeout: resourceDataSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"column_groups":                 quicksightschema.DataSetColumnGroupsSchema(),
				"column_level_permission_rules": quicksightschema.DataSetColumnLevelPermissionRulesSchema(),
				"data_set_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"data_set_usage_configuration": quicksightschema.DataSetUsageConfigurationSchema(),
				"field_folders":                quicksightschema.DataSetFieldFoldersSchema(),
				"import_mode": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.DataSetImportMode](),
				},
				"logical_table_map": quicksightschema.DataSetLogicalTableMapSchema(),
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"output_columns":                         quicksightschema.DataSetOutputColumnsSchema(),
				names.AttrPermissions:                    quicksightschema.PermissionsSchema(),
				"physical_table_map":                     quicksightschema.DataSetPhysicalTableMapSchema(),
				"row_level_permission_data_set":          quicksightschema.DataSetRowLevelPermissionDataSetSchema(),
				"row_level_permission_tag_configuration": quicksightschema.DataSetRowLevelPermissionTagConfigurationSchema(),
				"refresh_properties":                     quicksightschema.DataSetRefreshPropertiesSchema(),
				names.AttrTags:                           tftags.TagsSchema(),
				names.AttrTagsAll:                        tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				mode := diff.Get("import_mode").(string)
				if v, ok := diff.Get("refresh_properties").([]any); ok && v != nil && len(v) > 0 && mode == "DIRECT_QUERY" {
					return fmt.Errorf("refresh_properties cannot be set when import_mode is 'DIRECT_QUERY'")
				}
				return nil
			},
		),
	}
}

func resourceDataSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dataSetID := d.Get("data_set_id").(string)
	id := dataSetCreateResourceID(awsAccountID, dataSetID)
	input := &quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(awsAccountID),
		DataSetId:        aws.String(dataSetID),
		ImportMode:       awstypes.DataSetImportMode(d.Get("import_mode").(string)),
		PhysicalTableMap: quicksightschema.ExpandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("column_groups"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ColumnGroups = quicksightschema.ExpandColumnGroups(v.([]any))
	}

	if v, ok := d.GetOk("column_level_permission_rules"); ok && len(v.([]any)) > 0 {
		input.ColumnLevelPermissionRules = quicksightschema.ExpandColumnLevelPermissionRules(v.([]any))
	}

	if v, ok := d.GetOk("data_set_usage_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DataSetUsageConfiguration = quicksightschema.ExpandDataSetUsageConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("field_folders"); ok && v.(*schema.Set).Len() != 0 {
		input.FieldFolders = quicksightschema.ExpandFieldFolders(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("logical_table_map"); ok && v.(*schema.Set).Len() != 0 {
		input.LogicalTableMap = quicksightschema.ExpandLogicalTableMap(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("row_level_permission_data_set"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RowLevelPermissionDataSet = quicksightschema.ExpandRowLevelPermissionDataSet(v.([]any))
	}

	if v, ok := d.GetOk("row_level_permission_tag_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RowLevelPermissionTagConfiguration = quicksightschema.ExpandRowLevelPermissionTagConfiguration(v.([]any))
	}

	_, err := conn.CreateDataSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Set (%s): %s", id, err)
	}

	d.SetId(id)

	if v, ok := d.GetOk("refresh_properties"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input := &quicksight.PutDataSetRefreshPropertiesInput{
			AwsAccountId:             aws.String(awsAccountID),
			DataSetId:                aws.String(dataSetID),
			DataSetRefreshProperties: quicksightschema.ExpandDataSetRefreshProperties(v.([]any)),
		}

		_, err := conn.PutDataSetRefreshProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dataSet, err := findDataSetByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Data Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s): %s", d.Id(), err)
	}

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
	if err := d.Set("output_columns", quicksightschema.FlattenOutputColumns(dataSet.OutputColumns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_columns: %s", err)
	}
	if err := d.Set("physical_table_map", quicksightschema.FlattenPhysicalTableMap(dataSet.PhysicalTableMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_table_map: %s", err)
	}
	if err := d.Set("row_level_permission_data_set", quicksightschema.FlattenRowLevelPermissionDataSet(dataSet.RowLevelPermissionDataSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_data_set: %s", err)
	}
	if err := d.Set("row_level_permission_tag_configuration", quicksightschema.FlattenRowLevelPermissionTagConfiguration(dataSet.RowLevelPermissionTagConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting row_level_permission_tag_configuration: %s", err)
	}

	permissions, err := findDataSetPermissionsByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	refreshProperties, err := findDataSetRefreshPropertiesByTwoPartKey(ctx, conn, awsAccountID, dataSetID)

	switch {
	case tfresource.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
	default:
		if err := d.Set("refresh_properties", quicksightschema.FlattenDataSetRefreshProperties(refreshProperties)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting refresh properties: %s", err)
		}
	}

	return diags
}

func resourceDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll, "refresh_properties") {
		input := &quicksight.UpdateDataSetInput{
			AwsAccountId:                       aws.String(awsAccountID),
			ColumnGroups:                       quicksightschema.ExpandColumnGroups(d.Get("column_groups").([]any)),
			ColumnLevelPermissionRules:         quicksightschema.ExpandColumnLevelPermissionRules(d.Get("column_level_permission_rules").([]any)),
			DataSetId:                          aws.String(dataSetID),
			DataSetUsageConfiguration:          quicksightschema.ExpandDataSetUsageConfiguration(d.Get("data_set_usage_configuration").([]any)),
			FieldFolders:                       quicksightschema.ExpandFieldFolders(d.Get("field_folders").(*schema.Set).List()),
			ImportMode:                         awstypes.DataSetImportMode(d.Get("import_mode").(string)),
			LogicalTableMap:                    quicksightschema.ExpandLogicalTableMap(d.Get("logical_table_map").(*schema.Set).List()),
			Name:                               aws.String(d.Get(names.AttrName).(string)),
			PhysicalTableMap:                   quicksightschema.ExpandPhysicalTableMap(d.Get("physical_table_map").(*schema.Set).List()),
			RowLevelPermissionDataSet:          quicksightschema.ExpandRowLevelPermissionDataSet(d.Get("row_level_permission_data_set").([]any)),
			RowLevelPermissionTagConfiguration: quicksightschema.ExpandRowLevelPermissionTagConfiguration(d.Get("row_level_permission_tag_configuration").([]any)),
		}

		_, err = conn.UpdateDataSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateDataSetPermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDataSetPermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Set (%s) permissions: %s", d.Id(), err)
		}
	}

	if d.HasChange("refresh_properties") {
		o, n := d.GetChange("refresh_properties")

		if old, new := o.([]any), n.([]any); len(old) == 1 && len(new) == 0 {
			input := &quicksight.DeleteDataSetRefreshPropertiesInput{
				AwsAccountId: aws.String(awsAccountID),
				DataSetId:    aws.String(dataSetID),
			}

			_, err := conn.DeleteDataSetRefreshProperties(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
			}
		} else {
			input := &quicksight.PutDataSetRefreshPropertiesInput{
				AwsAccountId:             aws.String(awsAccountID),
				DataSetId:                aws.String(dataSetID),
				DataSetRefreshProperties: quicksightschema.ExpandDataSetRefreshProperties(new),
			}

			_, err = conn.PutDataSetRefreshProperties(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting QuickSight Data Set (%s) refresh properties: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSetID, err := dataSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Data Set: %s", d.Id())
	_, err = conn.DeleteDataSet(ctx, &quicksight.DeleteDataSetInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Set (%s): %s", d.Id(), err)
	}

	return diags
}

const dataSetResourceIDSeparator = ","

func dataSetCreateResourceID(awsAccountID, dataSetID string) string {
	parts := []string{awsAccountID, dataSetID}
	id := strings.Join(parts, dataSetResourceIDSeparator)

	return id
}

func dataSetParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, dataSetResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDATA_SET_ID", id, dataSetResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDataSetByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) (*awstypes.DataSet, error) {
	input := &quicksight.DescribeDataSetInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSet(ctx, conn, input)
}

func findDataSet(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetInput) (*awstypes.DataSet, error) {
	output, err := conn.DescribeDataSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSet, nil
}

func findDataSetRefreshPropertiesByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) (*awstypes.DataSetRefreshProperties, error) {
	input := &quicksight.DescribeDataSetRefreshPropertiesInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSetRefreshProperties(ctx, conn, input)
}

func findDataSetRefreshProperties(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetRefreshPropertiesInput) (*awstypes.DataSetRefreshProperties, error) {
	output, err := conn.DescribeDataSetRefreshProperties(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "not a SPICE dataset") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSetRefreshProperties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSetRefreshProperties, nil
}

func findDataSetPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeDataSetPermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
	}

	return findDataSetPermissions(ctx, conn, input)
}

func findDataSetPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSetPermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeDataSetPermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Permissions, nil
}
