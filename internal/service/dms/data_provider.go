// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dms_data_provider", name="Data Provider")
// @Tags(identifierAttribute="data_provider_arn")
func resourceDataProvider(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataProviderResource{}, nil
}

type dataProviderResource struct {
	framework.ResourceWithModel[dataProviderResourceModel]
}

func (r *dataProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"data_provider_arn": framework.ARNAttributeComputedOnly(),
			"data_provider_name": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrEngine: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"virtual": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"docdb_settings":                dataProviderSettingsBlock(ctx),
						"ibm_db2_luw_settings":          dataProviderSettingsBlock(ctx),
						"ibm_db2_zos_settings":          dataProviderSettingsBlock(ctx),
						"mariadb_settings":              dataProviderSettingsBlock(ctx),
						"microsoft_sql_server_settings": dataProviderSettingsBlock(ctx),
						"mongodb_settings":              dataProviderSettingsBlock(ctx),
						"mysql_settings":                dataProviderSettingsBlock(ctx),
						"oracle_settings":               dataProviderSettingsBlock(ctx),
						"postgres_settings":             dataProviderSettingsBlock(ctx),
						"redshift_settings":             dataProviderSettingsBlock(ctx),
						"sybase_ase_settings":           dataProviderSettingsBlock(ctx),
					},
				},
			},
		},
	}
}

func dataProviderSettingsBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderDBSettingsModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrCertificateARN: schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Optional:   true,
				},
				names.AttrDatabaseName: schema.StringAttribute{
					Optional: true,
				},
				names.AttrPort: schema.Int64Attribute{
					Optional: true,
				},
				"server_name": schema.StringAttribute{
					Optional: true,
				},
				"ssl_mode": schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}
}

func (r *dataProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data dataProviderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DMSClient(ctx)

	input := &dms.CreateDataProviderInput{
		DataProviderName: fwflex.StringFromFramework(ctx, data.DataProviderName),
		Description:      fwflex.StringFromFramework(ctx, data.Description),
		Engine:           fwflex.StringFromFramework(ctx, data.Engine),
		Tags:             getTagsIn(ctx),
	}

	if !data.Settings.IsNull() {
		settings, diags := data.Settings.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Settings = expandDataProviderSettings(ctx, settings)
	}

	output, err := conn.CreateDataProvider(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating DMS Data Provider (%s)", data.DataProviderName.ValueString()), err.Error())
		return
	}

	data.DataProviderARN = fwflex.StringToFramework(ctx, output.DataProvider.DataProviderArn)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data dataProviderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DMSClient(ctx)

	output, err := findDataProviderByARN(ctx, conn, data.DataProviderARN.ValueString())
	if errs.IsA[*retry.NotFoundError](err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading DMS Data Provider (%s)", data.DataProviderARN.ValueString()), err.Error())
		return
	}

	data.DataProviderName = fwflex.StringToFramework(ctx, output.DataProviderName)
	data.Description = fwflex.StringToFramework(ctx, output.Description)
	data.Engine = fwflex.StringToFramework(ctx, output.Engine)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new dataProviderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DMSClient(ctx)

	if !new.DataProviderName.Equal(old.DataProviderName) ||
		!new.Description.Equal(old.Description) ||
		!new.Settings.Equal(old.Settings) ||
		!new.Virtual.Equal(old.Virtual) {
		input := &dms.ModifyDataProviderInput{
			DataProviderIdentifier: new.DataProviderARN.ValueStringPointer(),
			DataProviderName:       fwflex.StringFromFramework(ctx, new.DataProviderName),
			Description:            fwflex.StringFromFramework(ctx, new.Description),
		}

		if !new.Settings.IsNull() {
			settings, diags := new.Settings.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Settings = expandDataProviderSettings(ctx, settings)
		}

		_, err := conn.ModifyDataProvider(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating DMS Data Provider (%s)", new.DataProviderARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *dataProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data dataProviderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DMSClient(ctx)

	input := &dms.DeleteDataProviderInput{
		DataProviderIdentifier: data.DataProviderARN.ValueStringPointer(),
	}
	_, err := conn.DeleteDataProvider(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting DMS Data Provider (%s)", data.DataProviderARN.ValueString()), err.Error())
	}
}

func (r *dataProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("data_provider_arn"), request, response)
}

type dataProviderResourceModel struct {
	framework.WithRegionModel
	DataProviderARN  types.String                                               `tfsdk:"data_provider_arn"`
	DataProviderName types.String                                               `tfsdk:"data_provider_name"`
	Description      types.String                                               `tfsdk:"description"`
	Engine           types.String                                               `tfsdk:"engine"`
	Settings         fwtypes.ListNestedObjectValueOf[dataProviderSettingsModel] `tfsdk:"settings"`
	Tags             tftags.Map                                                 `tfsdk:"tags"`
	TagsAll          tftags.Map                                                 `tfsdk:"tags_all"`
	Virtual          types.Bool                                                 `tfsdk:"virtual"`
}

type dataProviderSettingsModel struct {
	DocDBSettings              fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"docdb_settings"`
	IBMDB2LUWSettings          fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"ibm_db2_luw_settings"`
	IBMDB2ZOSSettings          fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"ibm_db2_zos_settings"`
	MariaDBSettings            fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"mariadb_settings"`
	MicrosoftSQLServerSettings fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"microsoft_sql_server_settings"`
	MongoDBSettings            fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"mongodb_settings"`
	MySQLSettings              fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"mysql_settings"`
	OracleSettings             fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"oracle_settings"`
	PostgresSettings           fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"postgres_settings"`
	RedshiftSettings           fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"redshift_settings"`
	SybaseASESettings          fwtypes.ListNestedObjectValueOf[dataProviderDBSettingsModel] `tfsdk:"sybase_ase_settings"`
}

type dataProviderDBSettingsModel struct {
	CertificateARN fwtypes.ARN  `tfsdk:"certificate_arn"`
	DatabaseName   types.String `tfsdk:"database_name"`
	Port           types.Int64  `tfsdk:"port"`
	ServerName     types.String `tfsdk:"server_name"`
	SSLMode        types.String `tfsdk:"ssl_mode"`
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
		return nil, &retry.NotFoundError{}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DataProviders) == 0 {
		return nil, &retry.NotFoundError{}
	}

	return &output.DataProviders[0], nil
}

func expandDataProviderSettings(ctx context.Context, settings *dataProviderSettingsModel) awstypes.DataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	if !settings.PostgresSettings.IsNull() {
		if s, diags := settings.PostgresSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberPostgreSqlSettings{
				Value: *expandPostgreSQLDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.MySQLSettings.IsNull() {
		if s, diags := settings.MySQLSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberMySqlSettings{
				Value: *expandMySQLDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.OracleSettings.IsNull() {
		if s, diags := settings.OracleSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberOracleSettings{
				Value: *expandOracleDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.MariaDBSettings.IsNull() {
		if s, diags := settings.MariaDBSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberMariaDbSettings{
				Value: *expandMariaDBDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.MicrosoftSQLServerSettings.IsNull() {
		if s, diags := settings.MicrosoftSQLServerSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings{
				Value: *expandMicrosoftSQLServerDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.DocDBSettings.IsNull() {
		if s, diags := settings.DocDBSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberDocDbSettings{
				Value: *expandDocDBDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.MongoDBSettings.IsNull() {
		if s, diags := settings.MongoDBSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberMongoDbSettings{
				Value: *expandMongoDBDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.RedshiftSettings.IsNull() {
		if s, diags := settings.RedshiftSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberRedshiftSettings{
				Value: *expandRedshiftDataProviderSettings(ctx, s),
			}
		}
	}
	if !settings.IBMDB2LUWSettings.IsNull() {
		if s, diags := settings.IBMDB2LUWSettings.ToPtr(ctx); diags == nil && s != nil {
			return &awstypes.DataProviderSettingsMemberIbmDb2LuwSettings{
				Value: *expandIBMDB2LUWDataProviderSettings(ctx, s),
			}
		}
	}

	return nil
}

func expandPostgreSQLDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.PostgreSqlDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.PostgreSqlDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandMySQLDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.MySqlDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.MySqlDataProviderSettings{
		ServerName: fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandOracleDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.OracleDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.OracleDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandMariaDBDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.MariaDbDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.MariaDbDataProviderSettings{
		ServerName: fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandMicrosoftSQLServerDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.MicrosoftSqlServerDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.MicrosoftSqlServerDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandDocDBDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.DocDbDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.DocDbDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandMongoDBDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.MongoDbDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.MongoDbDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}

func expandRedshiftDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.RedshiftDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.RedshiftDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}

	return result
}

func expandIBMDB2LUWDataProviderSettings(ctx context.Context, settings *dataProviderDBSettingsModel) *awstypes.IbmDb2LuwDataProviderSettings { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	if settings == nil {
		return nil
	}

	result := &awstypes.IbmDb2LuwDataProviderSettings{
		DatabaseName: fwflex.StringFromFramework(ctx, settings.DatabaseName),
		ServerName:   fwflex.StringFromFramework(ctx, settings.ServerName),
	}

	if !settings.CertificateARN.IsNull() {
		result.CertificateArn = settings.CertificateARN.ValueStringPointer()
	}
	if !settings.Port.IsNull() {
		result.Port = aws.Int32(int32(settings.Port.ValueInt64()))
	}
	if !settings.SSLMode.IsNull() {
		result.SslMode = awstypes.DmsSslModeValue(settings.SSLMode.ValueString())
	}

	return result
}
