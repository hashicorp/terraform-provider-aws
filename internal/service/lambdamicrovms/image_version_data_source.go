// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_lambdamicrovms_image_version", name="Image Version")
// @Tags
// @Testing(tagsTest=false)
func newImageVersionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &imageVersionDataSource{}, nil
}

type imageVersionDataSource struct {
	framework.DataSourceWithModel[imageVersionDataSourceModel]
}

func (d *imageVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_os_capabilities": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringEnumType[awstypes.Capability](),
				ElementType: types.StringType,
				Computed:    true,
			},
			"base_image_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"base_image_version": schema.StringAttribute{
				Computed: true,
			},
			"build_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"code_artifact":     framework.DataSourceComputedListOfObjectAttribute[codeArtifactModel](ctx),
			"cpu_configuration": framework.DataSourceComputedListOfObjectAttribute[cpuConfigurationModel](ctx),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"egress_network_connectors": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"environment_variables": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"hooks": framework.DataSourceComputedListOfObjectAttribute[hooksModel](ctx),
			"image_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"image_identifier": schema.StringAttribute{
				Required: true,
			},
			"image_version": schema.StringAttribute{
				Required: true,
			},
			"logging":           framework.DataSourceComputedListOfObjectAttribute[imageVersionLoggingModel](ctx),
			names.AttrResources: framework.DataSourceComputedListOfObjectAttribute[resourcesModel](ctx),
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MicrovmImageVersionState](),
				Computed:   true,
			},
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MicrovmImageVersionStatus](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *imageVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().LambdaMicroVMsClient(ctx)

	var data imageVersionDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	imageIdentifier, imageVersion := fwflex.StringValueFromFramework(ctx, data.ImageIdentifier), fwflex.StringValueFromFramework(ctx, data.ImageVersion)
	out, err := findImageVersionByTwoPartKey(ctx, conn, imageIdentifier, imageVersion)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, imageIdentifier, "image_version", imageVersion)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, imageIdentifier, "image_version", imageVersion)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, imageIdentifier, "image_version", imageVersion)
}

func findImageVersionByTwoPartKey(ctx context.Context, conn *lambdamicrovms.Client, imageIdentifier, imageVersion string) (*lambdamicrovms.GetMicrovmImageVersionOutput, error) {
	input := lambdamicrovms.GetMicrovmImageVersionInput{
		ImageIdentifier: aws.String(imageIdentifier),
		ImageVersion:    aws.String(imageVersion),
	}

	return findImageVersion(ctx, conn, &input)
}

func findImageVersion(ctx context.Context, conn *lambdamicrovms.Client, input *lambdamicrovms.GetMicrovmImageVersionInput) (*lambdamicrovms.GetMicrovmImageVersionOutput, error) {
	out, err := conn.GetMicrovmImageVersion(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type imageVersionDataSourceModel struct {
	framework.WithRegionModel
	AdditionalOSCapabilities fwtypes.ListOfStringEnum[awstypes.Capability]             `tfsdk:"additional_os_capabilities"`
	BaseImageARN             fwtypes.ARN                                               `tfsdk:"base_image_arn"`
	BaseImageVersion         types.String                                              `tfsdk:"base_image_version"`
	BuildRoleARN             fwtypes.ARN                                               `tfsdk:"build_role_arn"`
	CodeArtifact             fwtypes.ListNestedObjectValueOf[codeArtifactModel]        `tfsdk:"code_artifact"`
	CPUConfigurations        fwtypes.ListNestedObjectValueOf[cpuConfigurationModel]    `tfsdk:"cpu_configuration"`
	CreatedAt                timetypes.RFC3339                                         `tfsdk:"created_at"`
	Description              types.String                                              `tfsdk:"description"`
	EgressNetworkConnectors  fwtypes.ListOfString                                      `tfsdk:"egress_network_connectors"`
	EnvironmentVariables     fwtypes.MapOfString                                       `tfsdk:"environment_variables"`
	Hooks                    fwtypes.ListNestedObjectValueOf[hooksModel]               `tfsdk:"hooks"`
	ImageARN                 fwtypes.ARN                                               `tfsdk:"image_arn"`
	ImageIdentifier          types.String                                              `tfsdk:"image_identifier"`
	ImageVersion             types.String                                              `tfsdk:"image_version"`
	Logging                  fwtypes.ListNestedObjectValueOf[imageVersionLoggingModel] `tfsdk:"logging"`
	Resources                fwtypes.ListNestedObjectValueOf[resourcesModel]           `tfsdk:"resources"`
	State                    fwtypes.StringEnum[awstypes.MicrovmImageVersionState]     `tfsdk:"state"`
	StateReason              types.String                                              `tfsdk:"state_reason"`
	Status                   fwtypes.StringEnum[awstypes.MicrovmImageVersionStatus]    `tfsdk:"status"`
	Tags                     tftags.Map                                                `tfsdk:"tags"`
	UpdatedAt                timetypes.RFC3339                                         `tfsdk:"updated_at"`
}

type hooksModel struct {
	MicroVMHooks      fwtypes.ListNestedObjectValueOf[microVMHooksModel]      `tfsdk:"microvm_hooks"`
	MicroVMImageHooks fwtypes.ListNestedObjectValueOf[microVMImageHooksModel] `tfsdk:"microvm_image_hooks"`
	Port              types.Int32                                             `tfsdk:"port"`
}

type microVMHooksModel struct {
	Resume                    fwtypes.StringEnum[awstypes.HookState] `tfsdk:"resume"`
	ResumeTimeoutInSeconds    types.Int32                            `tfsdk:"resume_timeout_in_seconds"`
	Run                       fwtypes.StringEnum[awstypes.HookState] `tfsdk:"run"`
	RunTimeoutInSeconds       types.Int32                            `tfsdk:"run_timeout_in_seconds"`
	Suspend                   fwtypes.StringEnum[awstypes.HookState] `tfsdk:"suspend"`
	SuspendTimeoutInSeconds   types.Int32                            `tfsdk:"suspend_timeout_in_seconds"`
	Terminate                 fwtypes.StringEnum[awstypes.HookState] `tfsdk:"terminate"`
	TerminateTimeoutInSeconds types.Int32                            `tfsdk:"terminate_timeout_in_seconds"`
}

type microVMImageHooksModel struct {
	Ready                    fwtypes.StringEnum[awstypes.HookState] `tfsdk:"ready"`
	ReadyTimeoutInSeconds    types.Int32                            `tfsdk:"ready_timeout_in_seconds"`
	Validate                 fwtypes.StringEnum[awstypes.HookState] `tfsdk:"validate"`
	ValidateTimeoutInSeconds types.Int32                            `tfsdk:"validate_timeout_in_seconds"`
}

type resourcesModel struct {
	MinimumMemoryInMiB types.Int32 `tfsdk:"minimum_memory_in_mib"`
}

type imageVersionLoggingModel struct {
	CloudWatch fwtypes.ListNestedObjectValueOf[cloudWatchLoggingModel] `tfsdk:"cloudwatch"`
	Disabled   fwtypes.ListNestedObjectValueOf[loggingDisabledModel]   `tfsdk:"disabled"`
}

var _ fwflex.Flattener = &imageVersionLoggingModel{}

func (m *imageVersionLoggingModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.LoggingMemberCloudWatch:
		var model cloudWatchLoggingModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		if diags.HasError() {
			return diags
		}
		m.CloudWatch = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		m.Disabled = fwtypes.NewListNestedObjectValueOfNull[loggingDisabledModel](ctx)

	case awstypes.LoggingMemberDisabled:
		m.CloudWatch = fwtypes.NewListNestedObjectValueOfNull[cloudWatchLoggingModel](ctx)
		m.Disabled = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &loggingDisabledModel{})

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("imageVersionLoggingModel.Flatten: %T", v),
		)
	}

	return diags
}
