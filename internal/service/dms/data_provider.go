// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_data_provider", name="Data Provider")
// @Tags(identifierAttribute="data_provider_arn")
func resourceDataProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataProviderCreate,
		ReadWithoutTimeout:   resourceDataProviderRead,
		UpdateWithoutTimeout: resourceDataProviderUpdate,
		DeleteWithoutTimeout: resourceDataProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"data_provider_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_provider_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"docdb_settings":                dataProviderSettingsSchema(),
						"ibm_db2_luw_settings":          dataProviderSettingsSchema(),
						"ibm_db2_zos_settings":          dataProviderSettingsSchema(),
						"mariadb_settings":              dataProviderSettingsSchema(),
						"microsoft_sql_server_settings": dataProviderSettingsSchema(),
						"mongodb_settings":              dataProviderSettingsSchema(),
						"mysql_settings":                dataProviderSettingsSchema(),
						"oracle_settings":               dataProviderSettingsSchema(),
						"postgres_settings":             dataProviderSettingsSchema(),
						"redshift_settings":             dataProviderSettingsSchema(),
						"sybase_ase_settings":           dataProviderSettingsSchema(),
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"virtual": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataProviderSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrCertificateARN: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrDatabaseName: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrPort: {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"server_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"ssl_mode": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func resourceDataProviderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	input := &dms.CreateDataProviderInput{
		Engine:   aws.String(d.Get(names.AttrEngine).(string)),
		Settings: expandDataProviderSettings(d.Get("settings").([]interface{})),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_provider_name"); ok {
		input.DataProviderName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("virtual"); ok {
		input.Virtual = aws.Bool(v.(bool))
	}

	output, err := conn.CreateDataProvider(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Data Provider: %s", err)
	}

	d.SetId(aws.ToString(output.DataProvider.DataProviderArn))

	return append(diags, resourceDataProviderRead(ctx, d, meta)...)
}

func resourceDataProviderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	provider, err := findDataProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Data Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Data Provider (%s): %s", d.Id(), err)
	}

	d.Set("data_provider_arn", provider.DataProviderArn)
	d.Set("data_provider_name", provider.DataProviderName)
	d.Set(names.AttrDescription, provider.Description)
	d.Set(names.AttrEngine, provider.Engine)
	d.Set("virtual", provider.Virtual)
	if err := d.Set("settings", flattenDataProviderSettings(provider.Settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting settings: %s", err)
	}

	return diags
}

func resourceDataProviderUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &dms.ModifyDataProviderInput{
			DataProviderIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("data_provider_name") {
			input.DataProviderName = aws.String(d.Get("data_provider_name").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("settings") {
			input.Settings = expandDataProviderSettings(d.Get("settings").([]interface{}))
		}

		_, err := conn.ModifyDataProvider(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Data Provider (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataProviderRead(ctx, d, meta)...)
}

func resourceDataProviderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	log.Printf("[DEBUG] Deleting DMS Data Provider: %s", d.Id())
	input := dms.DeleteDataProviderInput{
		DataProviderIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteDataProvider(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Data Provider (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataProviderByARN(ctx context.Context, conn *dms.Client, arn string) (*awstypes.DataProvider, error) {
	input := &dms.DescribeDataProvidersInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("data-provider-arn"),
				Values: []string{arn},
			},
		},
	}

	output, err := conn.DescribeDataProviders(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.DataProviders)
}

func expandDataProviderSettings(tfList []interface{}) awstypes.DataProviderSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["docdb_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberDocDbSettings{
			Value: *expandDocDBDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["ibm_db2_luw_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberIbmDb2LuwSettings{
			Value: *expandIBMDB2LUWDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["ibm_db2_zos_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberIbmDb2zOsSettings{
			Value: *expandIBMDB2zOSDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["mariadb_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberMariaDbSettings{
			Value: *expandMariaDBDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["microsoft_sql_server_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings{
			Value: *expandMicrosoftSQLServerDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["mongodb_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberMongoDbSettings{
			Value: *expandMongoDBDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["mysql_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberMySqlSettings{
			Value: *expandMySQLDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["oracle_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberOracleSettings{
			Value: *expandOracleDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["postgres_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberPostgreSqlSettings{
			Value: *expandPostgreSQLDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["redshift_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberRedshiftSettings{
			Value: *expandRedshiftDataProviderSettings(v),
		}
	}

	if v, ok := tfMap["sybase_ase_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return &awstypes.DataProviderSettingsMemberSybaseAseSettings{
			Value: *expandSybaseAseDataProviderSettings(v),
		}
	}

	return nil
}

func expandPostgreSQLDataProviderSettings(tfList []interface{}) *awstypes.PostgreSqlDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.PostgreSqlDataProviderSettings](tfList)
}

func expandMySQLDataProviderSettings(tfList []interface{}) *awstypes.MySqlDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.MySqlDataProviderSettings](tfList)
}

func expandDocDBDataProviderSettings(tfList []interface{}) *awstypes.DocDbDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.DocDbDataProviderSettings](tfList)
}

func expandIBMDB2LUWDataProviderSettings(tfList []interface{}) *awstypes.IbmDb2LuwDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.IbmDb2LuwDataProviderSettings](tfList)
}

func expandIBMDB2zOSDataProviderSettings(tfList []interface{}) *awstypes.IbmDb2zOsDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.IbmDb2zOsDataProviderSettings](tfList)
}

func expandMariaDBDataProviderSettings(tfList []interface{}) *awstypes.MariaDbDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.MariaDbDataProviderSettings](tfList)
}

func expandSybaseAseDataProviderSettings(tfList []interface{}) *awstypes.SybaseAseDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.SybaseAseDataProviderSettings](tfList)
}

func expandMicrosoftSQLServerDataProviderSettings(tfList []interface{}) *awstypes.MicrosoftSqlServerDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.MicrosoftSqlServerDataProviderSettings](tfList)
}

func expandMongoDBDataProviderSettings(tfList []interface{}) *awstypes.MongoDbDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.MongoDbDataProviderSettings](tfList)
}

func expandOracleDataProviderSettings(tfList []interface{}) *awstypes.OracleDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.OracleDataProviderSettings](tfList)
}

func expandRedshiftDataProviderSettings(tfList []interface{}) *awstypes.RedshiftDataProviderSettings {
	return expandGenericDataProviderSettings[awstypes.RedshiftDataProviderSettings](tfList)
}

func expandGenericDataProviderSettings[T any](tfList []interface{}) *T {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	var settings T
	v := &settings

	// Use reflection to set common fields
	val := reflect.ValueOf(v).Elem()

	if certArn, ok := tfMap[names.AttrCertificateARN].(string); ok && certArn != "" {
		if field := val.FieldByName("CertificateArn"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(aws.String(certArn)))
		}
	}

	if dbName, ok := tfMap[names.AttrDatabaseName].(string); ok && dbName != "" {
		if field := val.FieldByName("DatabaseName"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(aws.String(dbName)))
		}
	}

	if port, ok := tfMap[names.AttrPort].(int); ok && port != 0 {
		if field := val.FieldByName("Port"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(aws.Int32(int32(port))))
		}
	}

	if serverName, ok := tfMap["server_name"].(string); ok && serverName != "" {
		if field := val.FieldByName("ServerName"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(aws.String(serverName)))
		}
	}

	if sslMode, ok := tfMap["ssl_mode"].(string); ok && sslMode != "" {
		if field := val.FieldByName("SslMode"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(awstypes.DmsSslModeValue(sslMode)))
		}
	}

	return v
}

func flattenDataProviderSettings(settings awstypes.DataProviderSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	switch v := settings.(type) {
	case *awstypes.DataProviderSettingsMemberDocDbSettings:
		m["docdb_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberIbmDb2LuwSettings:
		m["ibm_db2_luw_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberIbmDb2zOsSettings:
		m["ibm_db2_zos_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberMariaDbSettings:
		m["mariadb_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings:
		m["microsoft_sql_server_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberMongoDbSettings:
		m["mongodb_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberMySqlSettings:
		m["mysql_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberOracleSettings:
		m["oracle_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberPostgreSqlSettings:
		m["postgres_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberRedshiftSettings:
		m["redshift_settings"] = flattenGenericDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberSybaseAseSettings:
		m["sybase_ase_settings"] = flattenGenericDataProviderSettings(&v.Value)
	}

	return []interface{}{m}
}

func flattenGenericDataProviderSettings(settings interface{}) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}
	val := reflect.ValueOf(settings).Elem()

	if field := val.FieldByName("CertificateArn"); field.IsValid() && !field.IsNil() {
		m[names.AttrCertificateARN] = aws.ToString(field.Interface().(*string))
	}

	if field := val.FieldByName("DatabaseName"); field.IsValid() && !field.IsNil() {
		m[names.AttrDatabaseName] = aws.ToString(field.Interface().(*string))
	}

	if field := val.FieldByName("Port"); field.IsValid() && !field.IsNil() {
		m[names.AttrPort] = aws.ToInt32(field.Interface().(*int32))
	}

	if field := val.FieldByName("ServerName"); field.IsValid() && !field.IsNil() {
		m["server_name"] = aws.ToString(field.Interface().(*string))
	}

	if field := val.FieldByName("SslMode"); field.IsValid() {
		if sslMode := field.Interface().(awstypes.DmsSslModeValue); sslMode != "" {
			m["ssl_mode"] = string(sslMode)
		}
	}

	return []interface{}{m}
}
