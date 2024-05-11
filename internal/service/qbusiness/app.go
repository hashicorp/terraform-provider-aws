// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @FrameworkResource("aws_qbusiness_app", name="Application")
// @Tags(identifierAttribute="arn")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplication{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
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

func (r *resourceApplication) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_app"
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
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
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of the IAM Identity Center instance you are either creating for—or connecting to—your Amazon Q Business application",
				Required:    true,
			},
			"identity_center_application_arn": framework.ARNAttributeComputedOnly(),
			names.AttrTags:                    tftags.TagsAttribute(),
			names.AttrTagsAll:                 tftags.TagsAttributeComputedOnly(),
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
		resp.Diagnostics.AddError("failed to create Q Business application", err.Error())
		return
	}

	appId := aws.ToString(out.ApplicationId)

	if _, err := waitApplicationCreated(ctx, conn, appId, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business application creation", err.Error())
		return
	}

	app, err := FindAppByID(ctx, conn, appId)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve created Q Business application (%s)", appId), err.Error())
		return
	}

	data.ApplicationId = fwflex.StringToFramework(ctx, app.ApplicationId)
	data.ApplicationArn = fwflex.StringToFramework(ctx, app.ApplicationArn)
	data.IdentityCenterApplicationArn = fwflex.StringToFramework(ctx, app.IdentityCenterApplicationArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	appId := data.ApplicationId.ValueString()
	input := &qbusiness.DeleteApplicationInput{
		ApplicationId: aws.String(appId),
	}

	_, err := conn.DeleteApplication(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete Q Business application (%s)", data.ApplicationId.ValueString()), err.Error())
		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, appId, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business application deletion", err.Error())
		return
	}
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	out, err := FindAppByID(ctx, conn, data.ApplicationId.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business application (%s)", data.ApplicationId.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.IdentityCenterInstanceArn = fwflex.StringToFrameworkARN(ctx, convertARN(out.IdentityCenterApplicationArn))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.AttachmentsConfiguration.Equal(new.AttachmentsConfiguration) ||
		!old.EncryptionConfiguration.Equal(new.EncryptionConfiguration) ||
		!old.Description.Equal(new.Description) ||
		!old.DisplayName.Equal(new.DisplayName) ||
		!old.RoleArn.Equal(new.RoleArn) ||
		!old.IdentityCenterInstanceArn.Equal(new.IdentityCenterInstanceArn) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdateApplicationInput{}

		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("failed to update Q Business application (%s)", old.ApplicationId.ValueString()), err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

// Converts the ARN of the Identity Center Application to the ARN of the Identity Center Instance
func convertARN(arn *string) *string {
	parts := strings.Split(*arn, ":")
	subParts := strings.Split(parts[5], "/")
	newArn := fmt.Sprintf("%s:%s:%s:::instance/%s", parts[0], parts[1], parts[2], subParts[1])
	return &newArn
}

func (r *resourceApplication) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceApplicationData struct {
	ApplicationId                types.String                                                  `tfsdk:"id"`
	ApplicationArn               types.String                                                  `tfsdk:"arn"`
	Description                  types.String                                                  `tfsdk:"description"`
	DisplayName                  types.String                                                  `tfsdk:"display_name"`
	RoleArn                      fwtypes.ARN                                                   `tfsdk:"iam_service_role_arn"`
	IdentityCenterInstanceArn    fwtypes.ARN                                                   `tfsdk:"identity_center_instance_arn"`
	IdentityCenterApplicationArn types.String                                                  `tfsdk:"identity_center_application_arn"`
	Tags                         types.Map                                                     `tfsdk:"tags"`
	TagsAll                      types.Map                                                     `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                                `tfsdk:"timeouts"`
	AttachmentsConfiguration     fwtypes.ListNestedObjectValueOf[attachmentsConfigurationData] `tfsdk:"attachments_configuration"`
	EncryptionConfiguration      fwtypes.ListNestedObjectValueOf[encryptionConfigurationData]  `tfsdk:"encryption_configuration"`
}

type attachmentsConfigurationData struct {
	AttachmentsControlMode fwtypes.StringEnum[awstypes.AttachmentsControlMode] `tfsdk:"attachments_control_mode"`
}

type encryptionConfigurationData struct {
	KMSKeyID types.String `tfsdk:"kms_key_id"`
}

func FindAppByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetApplicationOutput, error) {
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
