// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3tables

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table", name="Table")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableOutput")
// @Testing(importStateIdAttribute="arn")
// @Testing(importStateIdFunc="testAccTableImportStateIdFunc")
// @Testing(preCheck="testAccPreCheck")
func newTableResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableResource{}, nil
}

type tableResource struct {
	framework.ResourceWithModel[tableResourceModel]
}

func (r *tableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEncryptionConfiguration: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[encryptionConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OpenTableFormat](),
				Required:   true,
				// TODO: Only one format is currently supported. When a new value is added, we can determine if `format` can be changed in-place or must recreate the resource
			},
			// TODO: Once Protocol v6 is supported, convert this to a `schema.SingleNestedAttribute` with full schema information
			// Validations needed:
			// * iceberg_compaction.settings.target_file_size_mb:  int32validator.Between(64, 512)
			// * iceberg_snapshot_management.settings.max_snapshot_age_hours: int32validator.AtLeast(1)
			// * iceberg_snapshot_management.settings.min_snapshots_to_keep:  int32validator.AtLeast(1)
			"maintenance_configuration": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[tableMaintenanceConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"modified_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required:   true,
				Validators: tableNameValidator,
			},
			names.AttrNamespace: schema.StringAttribute{
				Required:   true,
				Validators: namespaceNameValidator,
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"table_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"warehouse_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"metadata": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tableMetadataModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"iceberg": schema.ListNestedBlock{
							Description: "Iceberg metadata configuration.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[icebergMetadataModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									names.AttrSchema: schema.ListNestedBlock{
										Description: "Schema configuration for the Iceberg table.",
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrField: schema.ListNestedBlock{
													Description: "List of schema fields for the Iceberg table.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required:    true,
																Description: "The name of the field.",
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
															"required": schema.BoolAttribute{
																Optional:    true,
																Computed:    true,
																Default:     booldefault.StaticBool(false),
																Description: "A Boolean value that specifies whether values are required for each row in this field. Default: false.",
																PlanModifiers: []planmodifier.Bool{
																	boolplanmodifier.RequiresReplace(),
																},
															},
															names.AttrType: schema.StringAttribute{
																Required:    true,
																Description: "The field type. S3 Tables supports all Apache Iceberg primitive types.",
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
													},
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
												},
											},
										},
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeBetween(1, 1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	var input s3tables.CreateTableInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	// Handle metadata separately since it's an interface type.
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadataModel, diags := data.Metadata.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(fwflex.Expand(ctx, metadataModel, &input.Metadata)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	_, err := conn.CreateTable(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table (%s)", name), err.Error())

		return
	}

	if !data.MaintenanceConfiguration.IsUnknown() && !data.MaintenanceConfiguration.IsNull() {
		mc, diags := data.MaintenanceConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if !mc.IcebergCompaction.IsNull() {
			typ := awstypes.TableMaintenanceTypeIcebergCompaction
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           aws.String(name),
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, diags := expandTableMaintenanceIcebergCompaction(ctx, mc.IcebergCompaction)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}

		if !mc.IcebergSnapshotManagement.IsNull() {
			typ := awstypes.TableMaintenanceTypeIcebergSnapshotManagement
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           aws.String(name),
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, diags := expandTableMaintenanceIcebergSnapshotManagement(ctx, mc.IcebergSnapshotManagement)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}
	}

	outputGT, err := findTableByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGT, &data, fwflex.WithFieldNamePrefix("Table"))...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Namespace = types.StringValue(outputGT.Namespace[0])

	outputGTMC, err := findTableMaintenanceConfigurationByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s) maintenance configuration", name), err.Error())

		return
	default:
		value, diags := flattenTableMaintenanceConfiguration(ctx, outputGTMC)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.MaintenanceConfiguration = value
	}

	awsEncryptionConfig, err := findTableEncryptionByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s) encryption", name), err.Error())

		return
	default:
		var encryptionConfiguration encryptionConfigurationModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, awsEncryptionConfig, &encryptionConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		var diags diag.Diagnostics
		data.EncryptionConfiguration, diags = fwtypes.NewObjectValueOf(ctx, &encryptionConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	outputGT, err := findTableByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGT, &data, fwflex.WithFieldNamePrefix("Table"))...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Namespace = types.StringValue(outputGT.Namespace[0])

	outputGTMC, err := findTableMaintenanceConfigurationByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s) maintenance configuration", name), err.Error())

		return
	default:
		value, diags := flattenTableMaintenanceConfiguration(ctx, outputGTMC)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.MaintenanceConfiguration = value
	}

	awsEncryptionConfig, err := findTableEncryptionByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s) encryption", name), err.Error())

		return
	default:
		var encryptionConfiguration encryptionConfigurationModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, awsEncryptionConfig, &encryptionConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		var diags diag.Diagnostics
		data.EncryptionConfiguration, diags = fwtypes.NewObjectValueOf(ctx, &encryptionConfiguration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old tableResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	// New name and namespace.
	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, new.Name), fwflex.StringValueFromFramework(ctx, new.Namespace), fwflex.StringValueFromFramework(ctx, new.TableBucketARN)

	if !new.Name.Equal(old.Name) || !new.Namespace.Equal(old.Namespace) {
		input := s3tables.RenameTableInput{
			Name:           old.Name.ValueStringPointer(),
			Namespace:      old.Namespace.ValueStringPointer(),
			TableBucketARN: aws.String(tableBucketARN),
		}

		if !new.Name.Equal(old.Name) {
			input.NewName = aws.String(name)
		}

		if !new.Namespace.Equal(old.Namespace) {
			input.NewNamespaceName = aws.String(namespace)
		}

		_, err := conn.RenameTable(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("renaming S3 Tables Table (%s)", name), err.Error())

			return
		}
	}

	if !new.MaintenanceConfiguration.Equal(old.MaintenanceConfiguration) {
		newMC, d := new.MaintenanceConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		oldMC, d := old.MaintenanceConfiguration.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		if !newMC.IcebergCompaction.Equal(oldMC.IcebergCompaction) {
			typ := awstypes.TableMaintenanceTypeIcebergCompaction
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           aws.String(name),
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, diags := expandTableMaintenanceIcebergCompaction(ctx, newMC.IcebergCompaction)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}

		if !newMC.IcebergSnapshotManagement.Equal(oldMC.IcebergSnapshotManagement) {
			typ := awstypes.TableMaintenanceTypeIcebergSnapshotManagement
			input := s3tables.PutTableMaintenanceConfigurationInput{
				Name:           aws.String(name),
				Namespace:      aws.String(namespace),
				TableBucketARN: aws.String(tableBucketARN),
				Type:           typ,
			}

			value, d := expandTableMaintenanceIcebergSnapshotManagement(ctx, newMC.IcebergSnapshotManagement)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Value = &value

			_, err := conn.PutTableMaintenanceConfiguration(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("putting S3 Tables Table (%s) maintenance configuration (%s)", name, typ), err.Error())

				return
			}
		}
	}

	outputGT, err := findTableByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGT, &new, fwflex.WithFieldNamePrefix("Table"))...)
	if response.Diagnostics.HasError() {
		return
	}
	new.Namespace = types.StringValue(outputGT.Namespace[0])

	outputGTMC, err := findTableMaintenanceConfigurationByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	switch {
	case retry.NotFound(err):
	case err != nil:
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table (%s) maintenance configuration", name), err.Error())

		return
	default:
		value, diags := flattenTableMaintenanceConfiguration(ctx, outputGTMC)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		new.MaintenanceConfiguration = value
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.DeleteTableInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := conn.DeleteTable(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table (%s)", name), err.Error())

		return
	}
}

func (r *tableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	identifier, err := parseTableIdentifier(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Tables must use the format <table bucket ARN>"+tableIDSeparator+"<namespace>"+tableIDSeparator+"<table name>.\n"+
				fmt.Sprintf("Had %q", request.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &response.State, &response.Diagnostics)
}

func findTableByThreePartKey(ctx context.Context, conn *s3tables.Client, tableBucketARN, namespace, name string) (*s3tables.GetTableOutput, error) {
	input := s3tables.GetTableInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTable(ctx, conn, &input)
}

func findTable(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableInput) (*s3tables.GetTableOutput, error) {
	output, err := conn.GetTable(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findTableEncryptionByThreePartKey(ctx context.Context, conn *s3tables.Client, tableBucketARN, namespace, name string) (*awstypes.EncryptionConfiguration, error) {
	input := s3tables.GetTableEncryptionInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTableEncryption(ctx, conn, &input)
}

func findTableEncryption(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableEncryptionInput) (*awstypes.EncryptionConfiguration, error) {
	output, err := conn.GetTableEncryption(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.EncryptionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.EncryptionConfiguration, nil
}

func findTableMaintenanceConfigurationByThreePartKey(ctx context.Context, conn *s3tables.Client, tableBucketARN, namespace, name string) (*s3tables.GetTableMaintenanceConfigurationOutput, error) {
	input := s3tables.GetTableMaintenanceConfigurationInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTableMaintenanceConfiguration(ctx, conn, &input)
}

func findTableMaintenanceConfiguration(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableMaintenanceConfigurationInput) (*s3tables.GetTableMaintenanceConfigurationOutput, error) {
	output, err := conn.GetTableMaintenanceConfiguration(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tableResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                              `tfsdk:"arn"`
	CreatedAt                timetypes.RFC3339                                         `tfsdk:"created_at"`
	CreatedBy                types.String                                              `tfsdk:"created_by"`
	EncryptionConfiguration  fwtypes.ObjectValueOf[encryptionConfigurationModel]       `tfsdk:"encryption_configuration"`
	Format                   fwtypes.StringEnum[awstypes.OpenTableFormat]              `tfsdk:"format"`
	MaintenanceConfiguration fwtypes.ObjectValueOf[tableMaintenanceConfigurationModel] `tfsdk:"maintenance_configuration" autoflex:"-"`
	Metadata                 fwtypes.ListNestedObjectValueOf[tableMetadataModel]       `tfsdk:"metadata"`
	MetadataLocation         types.String                                              `tfsdk:"metadata_location"`
	ModifiedAt               timetypes.RFC3339                                         `tfsdk:"modified_at"`
	ModifiedBy               types.String                                              `tfsdk:"modified_by"`
	Name                     types.String                                              `tfsdk:"name"`
	Namespace                types.String                                              `tfsdk:"namespace" autoflex:",noflatten"` // On read, Namespace is an array
	OwnerAccountID           types.String                                              `tfsdk:"owner_account_id"`
	TableBucketARN           fwtypes.ARN                                               `tfsdk:"table_bucket_arn"`
	Tags                     tftags.Map                                                `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                `tfsdk:"tags_all"`
	Type                     fwtypes.StringEnum[awstypes.TableType]                    `tfsdk:"type"`
	VersionToken             types.String                                              `tfsdk:"version_token"`
	WarehouseLocation        types.String                                              `tfsdk:"warehouse_location"`
}

type tableMaintenanceConfigurationModel struct {
	IcebergCompaction         fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]]         `tfsdk:"iceberg_compaction"`
	IcebergSnapshotManagement fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]] `tfsdk:"iceberg_snapshot_management"`
}

type tableMaintenanceConfigurationValueModel[T any] struct {
	Settings fwtypes.ObjectValueOf[T]                       `tfsdk:"settings"`
	Status   fwtypes.StringEnum[awstypes.MaintenanceStatus] `tfsdk:"status"`
}

type icebergCompactionSettingsModel struct {
	TargetFileSizeMB types.Int32 `tfsdk:"target_file_size_mb"`
}

type icebergSnapshotManagementSettingsModel struct {
	MaxSnapshotAgeHours types.Int32 `tfsdk:"max_snapshot_age_hours"`
	MinSnapshotsToKeep  types.Int32 `tfsdk:"min_snapshots_to_keep"`
}

func flattenTableMaintenanceConfiguration(ctx context.Context, in *s3tables.GetTableMaintenanceConfigurationOutput) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	compactionConfig := in.Configuration[string(awstypes.TableMaintenanceTypeIcebergCompaction)]
	compactionConfigModel, d := flattenTableMaintenanceIcebergCompaction(ctx, &compactionConfig)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	snapshotManagementConfig := in.Configuration[string(awstypes.TableMaintenanceTypeIcebergSnapshotManagement)]
	snapshotManagementConfigModel, d := flattenTableMaintenanceIcebergSnapshotManagement(ctx, &snapshotManagementConfig)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationModel{
		IcebergCompaction:         compactionConfigModel,
		IcebergSnapshotManagement: snapshotManagementConfigModel,
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandTableMaintenanceIcebergCompaction(ctx context.Context, in fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]]) (result awstypes.TableMaintenanceConfigurationValue, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	settings, d := expandIcebergCompactionSettings(ctx, model.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	result.Settings = settings
	result.Status = model.Status.ValueEnum()

	return result, diags
}

func flattenTableMaintenanceIcebergCompaction(ctx context.Context, in *awstypes.TableMaintenanceConfigurationValue) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	iceberg, d := flattenIcebergCompactionSettings(ctx, in.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationValueModel[icebergCompactionSettingsModel]{
		Settings: iceberg,
		Status:   fwtypes.StringEnumValue(in.Status),
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandIcebergCompactionSettings(ctx context.Context, in fwtypes.ObjectValueOf[icebergCompactionSettingsModel]) (result *awstypes.TableMaintenanceSettingsMemberIcebergCompaction, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	var value awstypes.IcebergCompactionSettings

	diags.Append(fwflex.Expand(ctx, model, &value)...)

	return &awstypes.TableMaintenanceSettingsMemberIcebergCompaction{
		Value: value,
	}, diags
}

func flattenIcebergCompactionSettings(ctx context.Context, in awstypes.TableMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergCompactionSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableMaintenanceSettingsMemberIcebergCompaction:
		var model icebergCompactionSettingsModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		result = fwtypes.NewObjectValueOfMust(ctx, &model)

	case *awstypes.UnknownUnionMember:
		tflog.Warn(ctx, "Unexpected tagged union member", map[string]any{
			"tag": t.Tag,
		})

	default:
		tflog.Warn(ctx, "Unexpected nil tagged union value")
	}
	return result, diags
}

func expandTableMaintenanceIcebergSnapshotManagement(ctx context.Context, in fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]]) (result awstypes.TableMaintenanceConfigurationValue, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	settings, d := expandIcebergSnapshotManagementSettings(ctx, model.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	result.Settings = settings
	result.Status = model.Status.ValueEnum()

	return result, diags
}

func flattenTableMaintenanceIcebergSnapshotManagement(ctx context.Context, in *awstypes.TableMaintenanceConfigurationValue) (result fwtypes.ObjectValueOf[tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	iceberg, d := flattenIcebergSnapshotManagementSettings(ctx, in.Settings)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	model := tableMaintenanceConfigurationValueModel[icebergSnapshotManagementSettingsModel]{
		Settings: iceberg,
		Status:   fwtypes.StringEnumValue(in.Status),
	}

	result, d = fwtypes.NewObjectValueOf(ctx, &model)
	diags.Append(d...)
	return result, diags
}

func expandIcebergSnapshotManagementSettings(ctx context.Context, in fwtypes.ObjectValueOf[icebergSnapshotManagementSettingsModel]) (result *awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement, diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	model, d := in.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}

	var value awstypes.IcebergSnapshotManagementSettings

	diags.Append(fwflex.Expand(ctx, model, &value)...)

	return &awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement{
		Value: value,
	}, diags
}

func flattenIcebergSnapshotManagementSettings(ctx context.Context, in awstypes.TableMaintenanceSettings) (result fwtypes.ObjectValueOf[icebergSnapshotManagementSettingsModel], diags diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	switch t := in.(type) {
	case *awstypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement:
		var model icebergSnapshotManagementSettingsModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		result = fwtypes.NewObjectValueOfMust(ctx, &model)

	case *awstypes.UnknownUnionMember:
		tflog.Warn(ctx, "Unexpected tagged union member", map[string]any{
			"tag": t.Tag,
		})

	default:
		tflog.Warn(ctx, "Unexpected nil tagged union value")
	}
	return result, diags
}

func tableIDFromTableARN(s string) (string, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	return tableIDFromTableARNResource(arn.Resource), nil
}

func tableIDFromTableARNResource(s string) string {
	parts := strings.SplitN(s, "/", 4)
	return parts[3]
}

type tableIdentifier struct {
	TableBucketARN string
	Namespace      string
	Name           string
}

const (
	tableIDSeparator = ";"
	tableIDParts     = 3
)

func parseTableIdentifier(s string) (tableIdentifier, error) {
	parts := strings.Split(s, tableIDSeparator)
	if len(parts) != tableIDParts {
		return tableIdentifier{}, errors.New("not enough parts")
	}
	for i := range tableIDParts {
		if parts[i] == "" {
			return tableIdentifier{}, errors.New("empty part")
		}
	}

	return tableIdentifier{
		TableBucketARN: parts[0],
		Namespace:      parts[1],
		Name:           parts[2],
	}, nil
}

func (id tableIdentifier) String() string {
	return id.TableBucketARN + tableIDSeparator +
		id.Namespace + tableIDSeparator +
		id.Name
}

func (id tableIdentifier) PopulateState(ctx context.Context, s *tfsdk.State, diags *diag.Diagnostics) {
	diags.Append(s.SetAttribute(ctx, path.Root("table_bucket_arn"), id.TableBucketARN)...)
	diags.Append(s.SetAttribute(ctx, path.Root(names.AttrNamespace), id.Namespace)...)
	diags.Append(s.SetAttribute(ctx, path.Root(names.AttrName), id.Name)...)
}

var tableNameValidator = []validator.String{
	stringvalidator.LengthBetween(1, 255),
	tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersUnderscores,
	tfstringvalidator.StartsWithLetterOrNumber,
	tfstringvalidator.EndsWithLetterOrNumber,
}

type tableMetadataModel struct {
	Iceberg fwtypes.ListNestedObjectValueOf[icebergMetadataModel] `tfsdk:"iceberg"`
}

type icebergMetadataModel struct {
	Schema fwtypes.ListNestedObjectValueOf[icebergSchemaModel] `tfsdk:"schema"`
}

type icebergSchemaModel struct {
	Fields fwtypes.ListNestedObjectValueOf[icebergSchemaFieldModel] `tfsdk:"field"`
}

type icebergSchemaFieldModel struct {
	Name     types.String `tfsdk:"name"`
	Required types.Bool   `tfsdk:"required"`
	Type     types.String `tfsdk:"type"`
}

var (
	_ fwflex.Expander = tableMetadataModel{}
)

func (m tableMetadataModel) Expand(ctx context.Context) (out any, diags diag.Diagnostics) {
	// If Iceberg metadata is set, expand it
	if !m.Iceberg.IsNull() && !m.Iceberg.IsUnknown() {
		icebergModel, d := m.Iceberg.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		// Create Iceberg schema
		var schema awstypes.IcebergMetadata

		diags.Append(fwflex.Expand(ctx, icebergModel, &schema)...)
		if diags.HasError() {
			return nil, diags
		}

		out = &awstypes.TableMetadataMemberIceberg{
			Value: schema,
		}
	}

	return out, diags
}
