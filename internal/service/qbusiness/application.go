// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_qbusiness_application", name="Application")
// @Tags(identifierAttribute="arn")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplication{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Description: "A description of the Amazon Q application.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the Amazon Q application.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "The Amazon Resource Name (ARN) of the IAM service role that provides permissions for the Amazon Q application.",
				Required:    true,
			},
			names.AttrID:                      framework.IDAttribute(),
			"identity_center_application_arn": framework.ARNAttributeComputedOnly(),
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of the IAM Identity Center instance you are either creating for—or connecting to—your Amazon Q Business application",
				Required:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"attachments_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[attachmentsConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attachments_control_mode": schema.StringAttribute{
							Required:    true,
							Description: "Status information about whether file upload functionality is activated or deactivated for your end user.",
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AttachmentsControlMode](),
							},
						},
					},
				},
			},
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyID: schema.StringAttribute{
							Required:    true,
							Description: "The identifier of the AWS KMS key that is used to encrypt your data. Amazon Q doesn't support asymmetric keys.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
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

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.CreateApplicationInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	out, err := conn.CreateApplication(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionCreating, ResNameApplication, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}

	id := aws.ToString(out.ApplicationId)
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), id)

	if _, err := waitApplicationActive(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForCreation, ResNameApplication, id, err),
			err.Error(),
		)
		return
	}

	findOut, err := findApplicationByID(ctx, conn, id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionCreating, ResNameApplication, id, err),
			err.Error(),
		)
		return
	}

	// Set unknown values
	resp.Diagnostics.Append(fwflex.Flatten(ctx, findOut, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	out, err := findApplicationByID(ctx, conn, data.ApplicationId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionReading, ResNameApplication, data.ApplicationId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.AttachmentsConfiguration.Equal(plan.AttachmentsConfiguration) ||
		!state.EncryptionConfiguration.Equal(plan.EncryptionConfiguration) ||
		!state.Description.Equal(plan.Description) ||
		!state.DisplayName.Equal(plan.DisplayName) ||
		!state.RoleArn.Equal(plan.RoleArn) ||
		!state.IdentityCenterInstanceArn.Equal(plan.IdentityCenterInstanceArn) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdateApplicationInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApplication(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionUpdating, ResNameApplication, plan.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}

		id := plan.ApplicationId.ValueString()
		if _, err := waitApplicationActive(ctx, conn, id, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForUpdate, ResNameApplication, id, err),
				err.Error(),
			)
			return
		}

		findOut, err := findApplicationByID(ctx, conn, id)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionUpdating, ResNameApplication, id, err),
				err.Error(),
			)
			return
		}

		// Set unknown values
		resp.Diagnostics.Append(fwflex.Flatten(ctx, findOut, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	id := data.ApplicationId.ValueString()
	input := &qbusiness.DeleteApplicationInput{
		ApplicationId: aws.String(id),
	}

	if _, err := conn.DeleteApplication(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionDeleting, ResNameApplication, id, err),
			err.Error(),
		)
		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForDeletion, ResNameApplication, id, err),
			err.Error(),
		)
		return
	}
}

func findApplicationByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetApplicationOutput, error) {
	input := &qbusiness.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplication(ctx, input)
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

	return output, nil
}

func statusApplication(ctx context.Context, conn *qbusiness.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
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

func waitApplicationActive(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ApplicationStatusCreating, awstypes.ApplicationStatusUpdating),
		Target:     enum.Slice(awstypes.ApplicationStatusActive),
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ApplicationStatusActive, awstypes.ApplicationStatusDeleting),
		Target:     []string{},
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

type resourceApplicationData struct {
	ApplicationId                types.String                                                  `tfsdk:"id"`
	ApplicationArn               types.String                                                  `tfsdk:"arn"`
	AttachmentsConfiguration     fwtypes.ListNestedObjectValueOf[attachmentsConfigurationData] `tfsdk:"attachments_configuration"`
	Description                  types.String                                                  `tfsdk:"description"`
	DisplayName                  types.String                                                  `tfsdk:"display_name"`
	EncryptionConfiguration      fwtypes.ListNestedObjectValueOf[encryptionConfigurationData]  `tfsdk:"encryption_configuration"`
	IdentityCenterInstanceArn    fwtypes.ARN                                                   `tfsdk:"identity_center_instance_arn"`
	IdentityCenterApplicationArn types.String                                                  `tfsdk:"identity_center_application_arn"`
	RoleArn                      fwtypes.ARN                                                   `tfsdk:"iam_service_role_arn"`
	Tags                         tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                                `tfsdk:"timeouts"`
}

type attachmentsConfigurationData struct {
	AttachmentsControlMode fwtypes.StringEnum[awstypes.AttachmentsControlMode] `tfsdk:"attachments_control_mode"`
}

type encryptionConfigurationData struct {
	KMSKeyID types.String `tfsdk:"kms_key_id"`
}
