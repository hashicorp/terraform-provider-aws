// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
						"postgres_settings": {
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
						},
						"mysql_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCertificateARN: {
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
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
	_, err := conn.DeleteDataProvider(ctx, &dms.DeleteDataProviderInput{
		DataProviderIdentifier: aws.String(d.Id()),
	})

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
		return nil, &retry.NotFoundError{
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

	if v, ok := tfMap["postgres_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return expandPostgreSQLDataProviderSettings(v)
	}

	if v, ok := tfMap["mysql_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		return expandMySQLDataProviderSettings(v)
	}

	return nil
}

func expandPostgreSQLDataProviderSettings(tfList []interface{}) *awstypes.DataProviderSettingsMemberPostgreSqlSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	settings := &awstypes.PostgreSqlDataProviderSettings{}

	if v, ok := tfMap[names.AttrCertificateARN].(string); ok && v != "" {
		settings.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		settings.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
		settings.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["server_name"].(string); ok && v != "" {
		settings.ServerName = aws.String(v)
	}

	if v, ok := tfMap["ssl_mode"].(string); ok && v != "" {
		settings.SslMode = awstypes.DmsSslModeValue(v)
	}

	return &awstypes.DataProviderSettingsMemberPostgreSqlSettings{
		Value: *settings,
	}
}

func expandMySQLDataProviderSettings(tfList []interface{}) *awstypes.DataProviderSettingsMemberMySqlSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	settings := &awstypes.MySqlDataProviderSettings{}

	if v, ok := tfMap[names.AttrCertificateARN].(string); ok && v != "" {
		settings.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
		settings.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["server_name"].(string); ok && v != "" {
		settings.ServerName = aws.String(v)
	}

	if v, ok := tfMap["ssl_mode"].(string); ok && v != "" {
		settings.SslMode = awstypes.DmsSslModeValue(v)
	}

	return &awstypes.DataProviderSettingsMemberMySqlSettings{
		Value: *settings,
	}
}

func flattenDataProviderSettings(settings awstypes.DataProviderSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	switch v := settings.(type) {
	case *awstypes.DataProviderSettingsMemberPostgreSqlSettings:
		m["postgres_settings"] = flattenPostgreSQLDataProviderSettings(&v.Value)
	case *awstypes.DataProviderSettingsMemberMySqlSettings:
		m["mysql_settings"] = flattenMySQLDataProviderSettings(&v.Value)
	}

	return []interface{}{m}
}

func flattenPostgreSQLDataProviderSettings(settings *awstypes.PostgreSqlDataProviderSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := settings.CertificateArn; v != nil {
		m[names.AttrCertificateARN] = aws.ToString(v)
	}

	if v := settings.DatabaseName; v != nil {
		m[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := settings.Port; v != nil {
		m[names.AttrPort] = aws.ToInt32(v)
	}

	if v := settings.ServerName; v != nil {
		m["server_name"] = aws.ToString(v)
	}

	if v := settings.SslMode; v != "" {
		m["ssl_mode"] = string(v)
	}

	return []interface{}{m}
}

func flattenMySQLDataProviderSettings(settings *awstypes.MySqlDataProviderSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := settings.CertificateArn; v != nil {
		m[names.AttrCertificateARN] = aws.ToString(v)
	}

	if v := settings.Port; v != nil {
		m[names.AttrPort] = aws.ToInt32(v)
	}

	if v := settings.ServerName; v != nil {
		m["server_name"] = aws.ToString(v)
	}

	if v := settings.SslMode; v != "" {
		m["ssl_mode"] = string(v)
	}

	return []interface{}{m}
}
