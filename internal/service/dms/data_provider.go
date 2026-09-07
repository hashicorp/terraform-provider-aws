// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dms_data_provider", name="Data Provider")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types;awstypes;awstypes.DataProvider")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importStateIdAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
func newDataProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataProviderResource{}, nil
}

const resNameDataProvider = "Data Provider"

type dataProviderResource struct {
	framework.ResourceWithModel[dataProviderResourceModel]
	framework.WithImportByIdentity
}

func (r *dataProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			names.AttrEngine: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("aurora", "aurora-postgresql", "db2", "db2-zos", "docdb", "mariadb", "mongodb", "mysql", "oracle", "postgres", "redshift", "sqlserver", "sybase"),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"virtual": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"settings": dataProviderSettingsBlock(ctx),
		},
	}
}

func (r *dataProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan dataProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input databasemigrationservice.CreateDataProviderInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("DataProvider")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDataProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.DataProvider == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.DataProvider, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *dataProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state dataProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDataProviderByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *dataProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan, state dataProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input databasemigrationservice.ModifyDataProviderInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("DataProvider")))
		if resp.Diagnostics.HasError() {
			return
		}
		input.DataProviderIdentifier = state.ARN.ValueStringPointer()
		// Replace rather than merge settings so removed arguments are cleared.
		input.ExactSettings = aws.Bool(true)

		out, err := conn.ModifyDataProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
			return
		}
		if out == nil || out.DataProvider == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, state.ARN.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.DataProvider, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *dataProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state dataProviderResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := databasemigrationservice.DeleteDataProviderInput{
		DataProviderIdentifier: state.ARN.ValueStringPointer(),
	}
	_, err := conn.DeleteDataProvider(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}
}

func (r *dataProviderResource) flatten(ctx context.Context, out *awstypes.DataProvider, data *dataProviderResourceModel) diag.Diagnostics {
	return flex.Flatten(ctx, out, data, flex.WithFieldNamePrefix("DataProvider"))
}

func findDataProviderByARN(ctx context.Context, conn *databasemigrationservice.Client, arn string) (*awstypes.DataProvider, error) {
	input := databasemigrationservice.DescribeDataProvidersInput{
		Filters: []awstypes.Filter{{
			Name:   aws.String("data-provider-identifier"),
			Values: []string{arn},
		}},
	}

	var output []awstypes.DataProvider
	for item, err := range listDataProviders(ctx, conn, &input) {
		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		output = append(output, item)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

type dataProviderResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                                               `tfsdk:"arn"`
	CreationTime timetypes.RFC3339                                          `tfsdk:"creation_time"`
	Description  types.String                                               `tfsdk:"description"`
	Engine       types.String                                               `tfsdk:"engine"`
	Name         types.String                                               `tfsdk:"name"`
	Settings     fwtypes.ListNestedObjectValueOf[dataProviderSettingsModel] `tfsdk:"settings"`
	Tags         tftags.Map                                                 `tfsdk:"tags"`
	TagsAll      tftags.Map                                                 `tfsdk:"tags_all"`
	Virtual      types.Bool                                                 `tfsdk:"virtual"`
}

func dataProviderSettingsBlock(ctx context.Context) schema.ListNestedBlock {
	certificateARN := schema.StringAttribute{CustomType: fwtypes.ARNType, Optional: true}
	accessRoleARN := schema.StringAttribute{CustomType: fwtypes.ARNType, Optional: true}
	port := schema.Int32Attribute{
		Optional:   true,
		Validators: []validator.Int32{int32validator.Between(1, 65535)},
	}
	sslMode := schema.StringAttribute{
		CustomType: fwtypes.StringEnumType[awstypes.DmsSslModeValue](),
		Optional:   true,
		Computed:   true,
	}
	db2SSLMode := sslMode
	db2SSLMode.Validators = []validator.String{stringvalidator.OneOf("none", "verify-ca")}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderSettingsModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"doc_db_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderDocDBSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"ibm_db2_luw_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderIBMDb2LUWSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							"encryption_algorithm":   schema.Int32Attribute{Optional: true, Computed: true},
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"security_mechanism":     schema.Int32Attribute{Optional: true, Computed: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               db2SSLMode,
						},
					},
				},
				"ibm_db2_zos_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderIBMDb2ZOSSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               db2SSLMode,
						},
					},
				},
				"maria_db_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderMariaDBSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"microsoft_sql_server_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderMicrosoftSQLServerSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"mongo_db_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderMongoDBSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"auth_mechanism": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.AuthMechanismValue](),
								Optional:   true,
								Computed:   true,
							},
							"auth_source": schema.StringAttribute{Optional: true, Computed: true},
							"auth_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.AuthTypeValue](),
								Optional:   true,
								Computed:   true,
							},
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"mysql_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderMySQLSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"oracle_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderOracleSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"asm_server":             schema.StringAttribute{Optional: true},
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"secrets_manager_oracle_asm_access_role_arn":             accessRoleARN,
							"secrets_manager_oracle_asm_secret_id":                   schema.StringAttribute{Optional: true},
							"secrets_manager_security_db_encryption_access_role_arn": accessRoleARN,
							"secrets_manager_security_db_encryption_secret_id":       schema.StringAttribute{Optional: true},
							"server_name": schema.StringAttribute{Optional: true},
							"ssl_mode":    sslMode,
						},
					},
				},
				"postgresql_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderPostgreSQLSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							names.AttrPort:           port,
							"s3_access_role_arn":     accessRoleARN,
							"s3_path":                schema.StringAttribute{Optional: true},
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
				"redshift_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderRedshiftSettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrDatabaseName: schema.StringAttribute{Optional: true},
							names.AttrPort:         port,
							"s3_access_role_arn":   accessRoleARN,
							"s3_path":              schema.StringAttribute{Optional: true},
							"server_name":          schema.StringAttribute{Optional: true},
						},
					},
				},
				"sybase_ase_settings": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[dataProviderSybaseASESettingsModel](ctx),
					Validators: dataProviderSettingsUnionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrCertificateARN: certificateARN,
							names.AttrDatabaseName:   schema.StringAttribute{Optional: true},
							"encrypt_password":       schema.BoolAttribute{Optional: true, Computed: true},
							names.AttrPort:           port,
							"server_name":            schema.StringAttribute{Optional: true},
							"ssl_mode":               sslMode,
						},
					},
				},
			},
		},
	}
}

func dataProviderSettingsUnionValidators() []validator.List {
	return []validator.List{
		listvalidator.SizeAtMost(1),
		listvalidator.ExactlyOneOf(
			path.MatchRelative().AtParent().AtName("doc_db_settings"),
			path.MatchRelative().AtParent().AtName("ibm_db2_luw_settings"),
			path.MatchRelative().AtParent().AtName("ibm_db2_zos_settings"),
			path.MatchRelative().AtParent().AtName("maria_db_settings"),
			path.MatchRelative().AtParent().AtName("microsoft_sql_server_settings"),
			path.MatchRelative().AtParent().AtName("mongo_db_settings"),
			path.MatchRelative().AtParent().AtName("mysql_settings"),
			path.MatchRelative().AtParent().AtName("oracle_settings"),
			path.MatchRelative().AtParent().AtName("postgresql_settings"),
			path.MatchRelative().AtParent().AtName("redshift_settings"),
			path.MatchRelative().AtParent().AtName("sybase_ase_settings"),
		),
	}
}

type dataProviderSettingsModel struct {
	DocDBSettings              fwtypes.ListNestedObjectValueOf[dataProviderDocDBSettingsModel]              `tfsdk:"doc_db_settings"`
	IBMDb2LUWSettings          fwtypes.ListNestedObjectValueOf[dataProviderIBMDb2LUWSettingsModel]          `tfsdk:"ibm_db2_luw_settings"`
	IBMDb2ZOSSettings          fwtypes.ListNestedObjectValueOf[dataProviderIBMDb2ZOSSettingsModel]          `tfsdk:"ibm_db2_zos_settings"`
	MariaDBSettings            fwtypes.ListNestedObjectValueOf[dataProviderMariaDBSettingsModel]            `tfsdk:"maria_db_settings"`
	MicrosoftSQLServerSettings fwtypes.ListNestedObjectValueOf[dataProviderMicrosoftSQLServerSettingsModel] `tfsdk:"microsoft_sql_server_settings"`
	MongoDBSettings            fwtypes.ListNestedObjectValueOf[dataProviderMongoDBSettingsModel]            `tfsdk:"mongo_db_settings"`
	MySQLSettings              fwtypes.ListNestedObjectValueOf[dataProviderMySQLSettingsModel]              `tfsdk:"mysql_settings"`
	OracleSettings             fwtypes.ListNestedObjectValueOf[dataProviderOracleSettingsModel]             `tfsdk:"oracle_settings"`
	PostgreSQLSettings         fwtypes.ListNestedObjectValueOf[dataProviderPostgreSQLSettingsModel]         `tfsdk:"postgresql_settings"`
	RedshiftSettings           fwtypes.ListNestedObjectValueOf[dataProviderRedshiftSettingsModel]           `tfsdk:"redshift_settings"`
	SybaseASESettings          fwtypes.ListNestedObjectValueOf[dataProviderSybaseASESettingsModel]          `tfsdk:"sybase_ase_settings"`
}

type dataProviderDocDBSettingsModel struct {
	CertificateARN fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName   types.String                                 `tfsdk:"database_name"`
	Port           types.Int32                                  `tfsdk:"port"`
	ServerName     types.String                                 `tfsdk:"server_name"`
	SSLMode        fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderIBMDb2LUWSettingsModel struct {
	CertificateARN      fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName        types.String                                 `tfsdk:"database_name"`
	EncryptionAlgorithm types.Int32                                  `tfsdk:"encryption_algorithm"`
	Port                types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN     fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path              types.String                                 `tfsdk:"s3_path"`
	SecurityMechanism   types.Int32                                  `tfsdk:"security_mechanism"`
	ServerName          types.String                                 `tfsdk:"server_name"`
	SSLMode             fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderIBMDb2ZOSSettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName    types.String                                 `tfsdk:"database_name"`
	Port            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String                                 `tfsdk:"s3_path"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderMariaDBSettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	Port            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String                                 `tfsdk:"s3_path"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderMicrosoftSQLServerSettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName    types.String                                 `tfsdk:"database_name"`
	Port            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String                                 `tfsdk:"s3_path"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderMongoDBSettingsModel struct {
	AuthMechanism  fwtypes.StringEnum[awstypes.AuthMechanismValue] `tfsdk:"auth_mechanism"`
	AuthSource     types.String                                    `tfsdk:"auth_source"`
	AuthType       fwtypes.StringEnum[awstypes.AuthTypeValue]      `tfsdk:"auth_type"`
	CertificateARN fwtypes.ARN                                     `tfsdk:"certificate_arn"`
	DatabaseName   types.String                                    `tfsdk:"database_name"`
	Port           types.Int32                                     `tfsdk:"port"`
	ServerName     types.String                                    `tfsdk:"server_name"`
	SSLMode        fwtypes.StringEnum[awstypes.DmsSslModeValue]    `tfsdk:"ssl_mode"`
}

type dataProviderMySQLSettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	Port            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String                                 `tfsdk:"s3_path"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderOracleSettingsModel struct {
	ASMServer                                       types.String                                 `tfsdk:"asm_server"`
	CertificateARN                                  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName                                    types.String                                 `tfsdk:"database_name"`
	Port                                            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN                                 fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path                                          types.String                                 `tfsdk:"s3_path"`
	SecretsManagerOracleASMAccessRoleARN            fwtypes.ARN                                  `tfsdk:"secrets_manager_oracle_asm_access_role_arn"`
	SecretsManagerOracleASMSecretID                 types.String                                 `tfsdk:"secrets_manager_oracle_asm_secret_id"`
	SecretsManagerSecurityDBEncryptionAccessRoleARN fwtypes.ARN                                  `tfsdk:"secrets_manager_security_db_encryption_access_role_arn"`
	SecretsManagerSecurityDBEncryptionSecretID      types.String                                 `tfsdk:"secrets_manager_security_db_encryption_secret_id"`
	ServerName                                      types.String                                 `tfsdk:"server_name"`
	SSLMode                                         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderPostgreSQLSettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName    types.String                                 `tfsdk:"database_name"`
	Port            types.Int32                                  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN                                  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String                                 `tfsdk:"s3_path"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

type dataProviderRedshiftSettingsModel struct {
	DatabaseName    types.String `tfsdk:"database_name"`
	Port            types.Int32  `tfsdk:"port"`
	S3AccessRoleARN fwtypes.ARN  `tfsdk:"s3_access_role_arn"`
	S3Path          types.String `tfsdk:"s3_path"`
	ServerName      types.String `tfsdk:"server_name"`
}

type dataProviderSybaseASESettingsModel struct {
	CertificateARN  fwtypes.ARN                                  `tfsdk:"certificate_arn"`
	DatabaseName    types.String                                 `tfsdk:"database_name"`
	EncryptPassword types.Bool                                   `tfsdk:"encrypt_password"`
	Port            types.Int32                                  `tfsdk:"port"`
	ServerName      types.String                                 `tfsdk:"server_name"`
	SSLMode         fwtypes.StringEnum[awstypes.DmsSslModeValue] `tfsdk:"ssl_mode"`
}

var (
	_ flex.Expander  = dataProviderSettingsModel{}
	_ flex.Flattener = &dataProviderSettingsModel{}
)

// Expand implements the AutoFlex union conversion.
func (m dataProviderSettingsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	count := 0
	for _, v := range []basetypes.ListValue{
		m.DocDBSettings.ListValue,
		m.IBMDb2LUWSettings.ListValue,
		m.IBMDb2ZOSSettings.ListValue,
		m.MariaDBSettings.ListValue,
		m.MicrosoftSQLServerSettings.ListValue,
		m.MongoDBSettings.ListValue,
		m.MySQLSettings.ListValue,
		m.OracleSettings.ListValue,
		m.PostgreSQLSettings.ListValue,
		m.RedshiftSettings.ListValue,
		m.SybaseASESettings.ListValue,
	} {
		if v.IsUnknown() {
			smerr.AddError(ctx, &diags, fmt.Errorf("data provider settings contain an unknown engine settings block"))
			return nil, diags
		}
		if len(v.Elements()) > 1 {
			smerr.AddError(ctx, &diags, fmt.Errorf("data provider settings must contain exactly one engine settings block"))
			return nil, diags
		}
		for _, element := range v.Elements() {
			if element.IsNull() || element.IsUnknown() {
				smerr.AddError(ctx, &diags, fmt.Errorf("data provider settings contain a null or unknown engine settings object"))
				return nil, diags
			}
		}
		count += len(v.Elements())
	}
	if count != 1 {
		smerr.AddError(ctx, &diags, fmt.Errorf("data provider settings must contain exactly one engine settings block, got %d", count))
		return nil, diags
	}

	switch {
	case len(m.DocDBSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberDocDbSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.DocDBSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.IBMDb2LUWSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberIbmDb2LuwSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.IBMDb2LUWSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.IBMDb2ZOSSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberIbmDb2zOsSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.IBMDb2ZOSSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.MariaDBSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberMariaDbSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.MariaDBSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.MicrosoftSQLServerSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.MicrosoftSQLServerSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.MongoDBSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberMongoDbSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.MongoDBSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.MySQLSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberMySqlSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.MySQLSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.OracleSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberOracleSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.OracleSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.PostgreSQLSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberPostgreSqlSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.PostgreSQLSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.RedshiftSettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberRedshiftSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.RedshiftSettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	case len(m.SybaseASESettings.Elements()) == 1:
		var result awstypes.DataProviderSettingsMemberSybaseAseSettings
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.SybaseASESettings, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	default:
		smerr.AddError(ctx, &diags, fmt.Errorf("data provider settings contain no supported engine settings"))
		return nil, diags
	}
}

// Flatten implements the AutoFlex union conversion.
func (m *dataProviderSettingsModel) Flatten(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics
	result := dataProviderSettingsModel{
		DocDBSettings:              fwtypes.NewListNestedObjectValueOfNull[dataProviderDocDBSettingsModel](ctx),
		IBMDb2LUWSettings:          fwtypes.NewListNestedObjectValueOfNull[dataProviderIBMDb2LUWSettingsModel](ctx),
		IBMDb2ZOSSettings:          fwtypes.NewListNestedObjectValueOfNull[dataProviderIBMDb2ZOSSettingsModel](ctx),
		MariaDBSettings:            fwtypes.NewListNestedObjectValueOfNull[dataProviderMariaDBSettingsModel](ctx),
		MicrosoftSQLServerSettings: fwtypes.NewListNestedObjectValueOfNull[dataProviderMicrosoftSQLServerSettingsModel](ctx),
		MongoDBSettings:            fwtypes.NewListNestedObjectValueOfNull[dataProviderMongoDBSettingsModel](ctx),
		MySQLSettings:              fwtypes.NewListNestedObjectValueOfNull[dataProviderMySQLSettingsModel](ctx),
		OracleSettings:             fwtypes.NewListNestedObjectValueOfNull[dataProviderOracleSettingsModel](ctx),
		PostgreSQLSettings:         fwtypes.NewListNestedObjectValueOfNull[dataProviderPostgreSQLSettingsModel](ctx),
		RedshiftSettings:           fwtypes.NewListNestedObjectValueOfNull[dataProviderRedshiftSettingsModel](ctx),
		SybaseASESettings:          fwtypes.NewListNestedObjectValueOfNull[dataProviderSybaseASESettingsModel](ctx),
	}
	switch v := value.(type) {
	case awstypes.DataProviderSettingsMemberDocDbSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.DocDBSettings))
	case awstypes.DataProviderSettingsMemberIbmDb2LuwSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.IBMDb2LUWSettings))
	case awstypes.DataProviderSettingsMemberIbmDb2zOsSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.IBMDb2ZOSSettings))
	case awstypes.DataProviderSettingsMemberMariaDbSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.MariaDBSettings))
	case awstypes.DataProviderSettingsMemberMicrosoftSqlServerSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.MicrosoftSQLServerSettings))
	case awstypes.DataProviderSettingsMemberMongoDbSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.MongoDBSettings))
	case awstypes.DataProviderSettingsMemberMySqlSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.MySQLSettings))
	case awstypes.DataProviderSettingsMemberOracleSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.OracleSettings))
	case awstypes.DataProviderSettingsMemberPostgreSqlSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.PostgreSQLSettings))
	case awstypes.DataProviderSettingsMemberRedshiftSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.RedshiftSettings))
	case awstypes.DataProviderSettingsMemberSybaseAseSettings:
		smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, v.Value, &result.SybaseASESettings))
	default:
		smerr.AddError(ctx, &diags, fmt.Errorf("flattening data provider settings: unsupported union member %T", value))
	}
	if diags.HasError() {
		return diags
	}
	*m = result
	return diags
}
