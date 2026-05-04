// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_synchronization_configuration", name="Synchronization Configuration")
// @IdentityAttribute("file_system_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetSynchronizationConfigurationOutput")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource="true")
// @Testing(importStateIdAttribute="file_system_id")
func newSynchronizationConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &synchronizationConfigurationResource{}, nil
}

type synchronizationConfigurationResource struct {
	framework.ResourceWithModel[synchronizationConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *synchronizationConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrFileSystemID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "File system ID",
			},
			"latest_version_number": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Latest version number for optimistic locking",
			},
		},
		Blocks: map[string]schema.Block{
			"expiration_data_rule": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[expirationDataRuleModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"days_after_last_access": schema.Int32Attribute{
							Required:    true,
							Description: "Days after last access before data expires",
						},
					},
				},
			},
			"import_data_rule": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[importDataRuleModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPrefix: schema.StringAttribute{
							Required:    true,
							Description: "S3 prefix for import",
						},
						"size_less_than": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Description: "Maximum file size to import",
						},
						"trigger": schema.StringAttribute{
							Required:    true,
							CustomType:  fwtypes.StringEnumType[awstypes.ImportTrigger](),
							Description: "Import trigger type",
						},
					},
				},
			},
		},
	}
}

func (r *synchronizationConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data synchronizationConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.PutSynchronizationConfigurationInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutSynchronizationConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	output, err := findSynchronizationConfigurationByFileSystemID(ctx, conn, data.FileSystemID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *synchronizationConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data synchronizationConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findSynchronizationConfigurationByFileSystemID(ctx, conn, data.FileSystemID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.FileSystemID.ValueString())
		return
	}

	flattenSynchronizationConfigurationResource(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func flattenSynchronizationConfigurationResource(ctx context.Context, output *s3files.GetSynchronizationConfigurationOutput, data *synchronizationConfigurationResourceModel, diags *diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, output, data))
}

func (r *synchronizationConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new synchronizationConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.PutSynchronizationConfigurationInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutSynchronizationConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.FileSystemID.ValueString())
		return
	}

	output, err := findSynchronizationConfigurationByFileSystemID(ctx, conn, new.FileSystemID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &new))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, new))
}

func (r *synchronizationConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data synchronizationConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.PutSynchronizationConfigurationInput{
		FileSystemId: data.FileSystemID.ValueStringPointer(),
		ImportDataRules: []awstypes.ImportDataRule{
			{
				Prefix:       aws.String(""),
				Trigger:      awstypes.ImportTriggerOnDirectoryFirstAccess,
				SizeLessThan: aws.Int64(131072),
			},
		},
		ExpirationDataRules: []awstypes.ExpirationDataRule{
			{
				DaysAfterLastAccess: aws.Int32(365),
			},
		},
		LatestVersionNumber: data.LatestVersionNumber.ValueInt32Pointer(),
	}

	_, err := conn.PutSynchronizationConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.FileSystemID.ValueString())
	}
}

type synchronizationConfigurationResourceModel struct {
	framework.WithRegionModel
	ExpirationDataRules fwtypes.SetNestedObjectValueOf[expirationDataRuleModel] `tfsdk:"expiration_data_rule"`
	FileSystemID        types.String                                            `tfsdk:"file_system_id"`
	ImportDataRules     fwtypes.SetNestedObjectValueOf[importDataRuleModel]     `tfsdk:"import_data_rule"`
	LatestVersionNumber types.Int32                                             `tfsdk:"latest_version_number"`
}

type expirationDataRuleModel struct {
	DaysAfterLastAccess types.Int32 `tfsdk:"days_after_last_access"`
}

type importDataRuleModel struct {
	Prefix       types.String                               `tfsdk:"prefix"`
	SizeLessThan types.Int64                                `tfsdk:"size_less_than"`
	Trigger      fwtypes.StringEnum[awstypes.ImportTrigger] `tfsdk:"trigger"`
}

func findSynchronizationConfigurationByFileSystemID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetSynchronizationConfigurationOutput, error) {
	input := s3files.GetSynchronizationConfigurationInput{
		FileSystemId: &id,
	}

	output, err := conn.GetSynchronizationConfiguration(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}
