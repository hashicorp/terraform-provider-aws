// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_m2_application", name="Application")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/m2;m2.GetApplicationOutput")
func newApplicationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &applicationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type applicationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*applicationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_application"
}

func (r *applicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: framework.IDAttribute(),
			names.AttrARN:           framework.ARNAttributeComputedOnly(),
			"current_version": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(500),
				},
			},
			"engine_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EngineType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{1,59}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[definitionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrContent: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 65000),
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrContent),
									path.MatchRelative().AtParent().AtName("s3_location"),
								),
							},
						},
						"s3_location": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *applicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	name := data.Name.ValueString()
	input := m2.CreateApplicationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.AccessDeniedException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateApplication(ctx, &input)
	}, "does not have proper Trust Policy for M2 service")

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Mainframe Modernization Application (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ApplicationID = fwflex.StringToFramework(ctx, outputRaw.(*m2.CreateApplicationOutput).ApplicationId)
	data.setID()

	app, err := waitApplicationCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Application (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, app, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.CurrentVersion = fwflex.Int32ToFramework(ctx, app.LatestVersion.ApplicationVersion)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *applicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().M2Client(ctx)

	outputGA, err := findApplicationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Mainframe Modernization Application (%s)", data.ID.ValueString()), err.Error())

		return
	}

	applicationVersion := aws.ToInt32(outputGA.LatestVersion.ApplicationVersion)
	outputGAV, err := findApplicationVersionByTwoPartKey(ctx, conn, data.ID.ValueString(), applicationVersion)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Mainframe Modernization Application (%s) version (%d)", data.ID.ValueString(), applicationVersion), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGA, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.CurrentVersion = fwflex.Int32ToFramework(ctx, outputGAV.ApplicationVersion)
	data.Definition = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &definitionModel{
		Content:    fwflex.StringToFramework(ctx, outputGAV.DefinitionContent),
		S3Location: types.StringNull(),
	})

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *applicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new applicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	if !new.Definition.Equal(old.Definition) || !new.Description.Equal(old.Description) {
		input := &m2.UpdateApplicationInput{
			ApplicationId:             fwflex.StringFromFramework(ctx, new.ID),
			CurrentApplicationVersion: fwflex.Int32FromFramework(ctx, old.CurrentVersion),
		}

		if !new.Definition.Equal(old.Definition) {
			d := fwflex.Expand(ctx, new.Definition, &input.Definition)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
		}

		if !new.Description.Equal(old.Description) {
			input.Description = fwflex.StringFromFramework(ctx, new.Description)
		}

		outputUA, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Mainframe Modernization Application (%s)", new.ID.ValueString()), err.Error())

			return
		}

		applicationVersion := aws.ToInt32(outputUA.ApplicationVersion)
		if _, err := waitApplicationUpdated(ctx, conn, new.ID.ValueString(), applicationVersion, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Application (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		new.CurrentVersion = types.Int64Value(int64(applicationVersion))
	} else {
		new.CurrentVersion = old.CurrentVersion
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *applicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	_, err := conn.DeleteApplication(ctx, &m2.DeleteApplicationInput{
		ApplicationId: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Mainframe Modernization Application (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Application (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *applicationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func startApplication(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) { //nolint:unparam
	input := &m2.StartApplicationInput{
		ApplicationId: aws.String(id),
	}

	_, err := conn.StartApplication(ctx, input)

	if err != nil {
		return nil, err
	}

	return waitApplicationRunning(ctx, conn, id, timeout)
}

func stopApplicationIfRunning(ctx context.Context, conn *m2.Client, id string, forceStop bool, timeout time.Duration) (*m2.GetApplicationOutput, error) { //nolint:unparam
	app, err := findApplicationByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if app.Status != awstypes.ApplicationLifecycleRunning {
		return nil, nil
	}

	input := &m2.StopApplicationInput{
		ApplicationId: aws.String(id),
		ForceStop:     forceStop,
	}

	_, err = conn.StopApplication(ctx, input)

	if err != nil {
		return nil, err
	}

	return waitApplicationStopped(ctx, conn, id, timeout)
}

func findApplicationByID(ctx context.Context, conn *m2.Client, id string) (*m2.GetApplicationOutput, error) {
	input := m2.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationId == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findApplicationVersionByTwoPartKey(ctx context.Context, conn *m2.Client, id string, version int32) (*m2.GetApplicationVersionOutput, error) {
	input := m2.GetApplicationVersionInput{
		ApplicationId:      aws.String(id),
		ApplicationVersion: aws.Int32(version),
	}

	output, err := conn.GetApplicationVersion(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationVersion == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusApplication(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusApplicationVersion(ctx context.Context, conn *m2.Client, id string, version int32) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationVersionByTwoPartKey(ctx, conn, id, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitApplicationCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleCreating),
		Target:  enum.Slice(awstypes.ApplicationLifecycleCreated, awstypes.ApplicationLifecycleAvailable),
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitApplicationUpdated(ctx context.Context, conn *m2.Client, id string, version int32, timeout time.Duration) (*m2.GetApplicationVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationVersionLifecycleCreating),
		Target:  enum.Slice(awstypes.ApplicationVersionLifecycleAvailable),
		Refresh: statusApplicationVersion(ctx, conn, id, version),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationVersionOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleDeleting, awstypes.ApplicationLifecycleDeletingFromEnvironment),
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitApplicationDeletedFromEnvironment(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleDeletingFromEnvironment),
		Target:  enum.Slice(awstypes.ApplicationLifecycleAvailable),
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitApplicationStopped(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleStopping),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleStopped),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitApplicationRunning(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleStarting),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleRunning),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type applicationResourceModel struct {
	ApplicationID  types.String                                     `tfsdk:"application_id"`
	ApplicationARN types.String                                     `tfsdk:"arn"`
	CurrentVersion types.Int64                                      `tfsdk:"current_version"`
	Definition     fwtypes.ListNestedObjectValueOf[definitionModel] `tfsdk:"definition"`
	Description    types.String                                     `tfsdk:"description"`
	EngineType     fwtypes.StringEnum[awstypes.EngineType]          `tfsdk:"engine_type"`
	ID             types.String                                     `tfsdk:"id"`
	KmsKeyID       types.String                                     `tfsdk:"kms_key_id"`
	Name           types.String                                     `tfsdk:"name"`
	RoleARN        fwtypes.ARN                                      `tfsdk:"role_arn"`
	Tags           types.Map                                        `tfsdk:"tags"`
	TagsAll        types.Map                                        `tfsdk:"tags_all"`
	Timeouts       timeouts.Value                                   `tfsdk:"timeouts"`
}

func (model *applicationResourceModel) InitFromID() error {
	model.ApplicationID = model.ID

	return nil
}

func (model *applicationResourceModel) setID() {
	model.ID = model.ApplicationID
}

type definitionModel struct {
	Content    types.String `tfsdk:"content"`
	S3Location types.String `tfsdk:"s3_location"`
}

var (
	_ fwflex.Expander = definitionModel{}
)

func (m definitionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Content.IsNull():
		result = &awstypes.DefinitionMemberContent{
			Value: m.Content.ValueString(),
		}

	case !m.S3Location.IsNull():
		result = &awstypes.DefinitionMemberS3Location{
			Value: m.S3Location.ValueString(),
		}
	}

	return result, diags
}
