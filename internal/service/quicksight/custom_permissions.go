// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package quicksight

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_custom_permissions", name="Custom Permissions")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.CustomPermissions")
// @Testing(skipEmptyTags=true, skipNullTags=true)
// @Testing(importStateIdFunc="testAccCustomPermissionsImportStateID", importStateIdAttribute="custom_permissions_name")
func newCustomPermissionsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &customPermissionsResource{}

	return r, nil
}

type customPermissionsResource struct {
	framework.ResourceWithModel[customPermissionsResourceModel]
}

func (r *customPermissionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:          framework.ARNAttributeComputedOnly(),
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"custom_permissions_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9+=,.@_-]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"capabilities": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capabilitiesModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: tfmaps.ApplyToAllValues(fwtypes.AttributeTypesMust[capabilitiesModel](ctx), func(attr.Type) schema.Attribute {
						return schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.CapabilityState](),
							Optional:   true,
						}
					}),
				},
			},
		},
	}
}

func (r *customPermissionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data customPermissionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.CustomPermissionsName)
	var input quicksight.CreateCustomPermissionsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateCustomPermissions(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Custom Permissions (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *customPermissionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data customPermissionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID, name := fwflex.StringValueFromFramework(ctx, data.AWSAccountID), fwflex.StringValueFromFramework(ctx, data.CustomPermissionsName)
	output, err := findCustomPermissionsByTwoPartKey(ctx, conn, accountID, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Quicksight Custom Permissions (%s)", name), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customPermissionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old customPermissionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	diff, diags := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, new.CustomPermissionsName)
		var input quicksight.UpdateCustomPermissionsInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateCustomPermissions(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Quicksight Custom Permissions (%s)", name), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *customPermissionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data customPermissionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID, name := fwflex.StringValueFromFramework(ctx, data.AWSAccountID), fwflex.StringValueFromFramework(ctx, data.CustomPermissionsName)
	input := quicksight.DeleteCustomPermissionsInput{
		AwsAccountId:          aws.String(accountID),
		CustomPermissionsName: aws.String(name),
	}
	_, err := conn.DeleteCustomPermissions(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight Custom Permissions (%s)", name), err.Error())

		return
	}
}

func (r *customPermissionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		customPermissionsIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, customPermissionsIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAWSAccountID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("custom_permissions_name"), parts[1])...)
}

func findCustomPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, customPermissionsName string) (*awstypes.CustomPermissions, error) {
	input := &quicksight.DescribeCustomPermissionsInput{
		AwsAccountId:          aws.String(awsAccountID),
		CustomPermissionsName: aws.String(customPermissionsName),
	}

	return findCustomPermissions(ctx, conn, input)
}

func findCustomPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeCustomPermissionsInput) (*awstypes.CustomPermissions, error) {
	output, err := conn.DescribeCustomPermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CustomPermissions == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.CustomPermissions, nil
}

type customPermissionsResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String                                       `tfsdk:"arn"`
	AWSAccountID          types.String                                       `tfsdk:"aws_account_id"`
	Capabilities          fwtypes.ListNestedObjectValueOf[capabilitiesModel] `tfsdk:"capabilities"`
	CustomPermissionsName types.String                                       `tfsdk:"custom_permissions_name"`
	Tags                  tftags.Map                                         `tfsdk:"tags"`
	TagsAll               tftags.Map                                         `tfsdk:"tags_all"`
}

type capabilitiesModel struct {
	AddOrRunAnomalyDetectionForAnalyses   fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"add_or_run_anomaly_detection_for_analyses"`
	CreateAndUpdateDashboardEmailReports  fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_and_update_dashboard_email_reports"`
	CreateAndUpdateDatasets               fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_and_update_datasets"`
	CreateAndUpdateDataSources            fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_and_update_data_sources"`
	CreateAndUpdateThemes                 fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_and_update_themes"`
	CreateAndUpdateThresholdAlerts        fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_and_update_threshold_alerts"`
	CreateSharedFolders                   fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_shared_folders"`
	CreateSPICEDataset                    fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"create_spice_dataset"`
	ExportToCSV                           fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_csv"`
	ExportToCSVInScheduledReports         fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_csv_in_scheduled_reports"`
	ExportToExcel                         fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_excel"`
	ExportToExcelInScheduledReports       fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_excel_in_scheduled_reports"`
	ExportToPDF                           fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_pdf"`
	ExportToPDFInScheduledReports         fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"export_to_pdf_in_scheduled_reports"`
	IncludeContentInScheduledReportsEmail fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"include_content_in_scheduled_reports_email"`
	PrintReports                          fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"print_reports"`
	RenameSharedFolders                   fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"rename_shared_folders"`
	ShareAnalyses                         fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"share_analyses"`
	ShareDashboards                       fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"share_dashboards"`
	ShareDatasets                         fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"share_datasets"`
	ShareDataSources                      fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"share_data_sources"`
	SubscribeDashboardEmailReports        fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"subscribe_dashboard_email_reports"`
	ViewAccountSPICECapacity              fwtypes.StringEnum[awstypes.CapabilityState] `tfsdk:"view_account_spice_capacity"`
}
