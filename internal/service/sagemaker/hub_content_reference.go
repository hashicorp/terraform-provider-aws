// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_hub_content_reference", name="Hub Content Reference")
// @Tags(identifierAttribute="hub_content_arn")
// @Testing(importStateIdAttribute="hub_name,hub_content_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeHubContentOutput")
func newHubContentReferenceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hubContentReferenceResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameHubContentReference     = "Hub Content Reference"
	hubContentReferenceIDPartCount = 2
)

type hubContentReferenceResource struct {
	framework.ResourceWithModel[hubContentReferenceResourceModel]
	framework.WithTimeouts
}

func (r *hubContentReferenceResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hub_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				Description: "ARN of the private SageMaker Hub that contains the content reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hub_content_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				Description: "ARN of the hub content reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hub_content_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the hub content reference.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`),
						"valid characters are a-z, A-Z, 0-9, and - (hyphen)",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hub_content_status": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.HubContentStatus](),
				Computed:    true,
				Description: "Status of the hub content reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hub_content_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the hub content reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hub_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the private SageMaker Hub to add the content reference to.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`),
						"valid characters are a-z, A-Z, 0-9, and - (hyphen)",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"min_version": schema.StringAttribute{
				Optional:    true,
				Description: "Minimum version of the hub content to reference. Use \"1.0.0\" to support all versions.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							// Removing min_version needs a replacement
							resp.RequiresReplace = !req.StateValue.IsNull() && req.StateValue.ValueString() != "" &&
								(req.PlanValue.IsNull() || req.PlanValue.ValueString() == "")
						},
						"Removing min_version requires replacement of the resource.",
						"Removing min_version requires replacement of the resource.",
					),
				},
			},
			"sagemaker_public_hub_content_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the public JumpStart hub content to reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *hubContentReferenceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data hubContentReferenceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	var input sagemaker.CreateHubContentReferenceInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateHubContentReference(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.HubContentName.ValueString())
		return
	}

	output, err := waitHubContentReferenceAvailable(ctx, conn, data.HubName.ValueString(), data.HubContentName.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.HubContentName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.HubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.HubContentArn))
	data.MinVersion = fwflex.StringToFramework(ctx, output.ReferenceMinVersion)
	data.SageMakerPublicHubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.SageMakerPublicHubContentArn))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *hubContentReferenceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data hubContentReferenceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	output, err := findHubContentByName(ctx, conn, data.HubName.ValueString(), data.HubContentName.ValueString(), awstypes.HubContentTypeModelReference)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.HubContentName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.HubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.HubContentArn))
	data.MinVersion = fwflex.StringToFramework(ctx, output.ReferenceMinVersion)
	data.SageMakerPublicHubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.SageMakerPublicHubContentArn))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *hubContentReferenceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old hubContentReferenceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	if !new.MinVersion.Equal(old.MinVersion) {
		conn := r.Meta().SageMakerClient(ctx)

		var input sagemaker.UpdateHubContentReferenceInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}
		input.HubContentType = awstypes.HubContentTypeModelReference

		_, err := conn.UpdateHubContentReference(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.HubContentName.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, new))
}

func (r *hubContentReferenceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data hubContentReferenceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	hubName := data.HubName.ValueString()
	hubContentName := data.HubContentName.ValueString()

	_, err := conn.DeleteHubContentReference(ctx, &sagemaker.DeleteHubContentReferenceInput{
		HubName:        aws.String(hubName),
		HubContentName: aws.String(hubContentName),
		HubContentType: awstypes.HubContentTypeModelReference,
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, hubContentName)
		return
	}

	if _, err := waitHubContentReferenceDeleted(ctx, conn, hubName, hubContentName, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, hubContentName)
	}
}

func (r *hubContentReferenceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := flex.ExpandResourceId(request.ID, hubContentReferenceIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError(
			"importing SageMaker Hub Content Reference",
			fmt.Sprintf("invalid import ID %q, expected hub_name%shub_content_name: %s", request.ID, flex.ResourceIdSeparator, err),
		)
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("hub_name"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("hub_content_name"), parts[1])...)
}

// stripARNVersion removes the version suffix from a SageMaker ARN
func stripARNVersion(arn *string) *string {
	s := aws.ToString(arn)
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[:i]
	}
	return aws.String(s)
}

func findHubContentByName(ctx context.Context, conn *sagemaker.Client, hubName, hubContentName string, contentType awstypes.HubContentType) (*sagemaker.DescribeHubContentOutput, error) {
	input := &sagemaker.DescribeHubContentInput{
		HubName:        aws.String(hubName),
		HubContentName: aws.String(hubContentName),
		HubContentType: contentType,
	}

	output, err := conn.DescribeHubContent(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

type hubContentReferenceResourceModel struct {
	framework.WithRegionModel
	HubArn                       fwtypes.ARN                                   `tfsdk:"hub_arn"`
	HubContentArn                fwtypes.ARN                                   `tfsdk:"hub_content_arn"`
	HubContentName               types.String                                  `tfsdk:"hub_content_name"`
	HubContentStatus             fwtypes.StringEnum[awstypes.HubContentStatus] `tfsdk:"hub_content_status"`
	HubContentVersion            types.String                                  `tfsdk:"hub_content_version"`
	HubName                      types.String                                  `tfsdk:"hub_name"`
	MinVersion                   types.String                                  `tfsdk:"min_version"`
	SageMakerPublicHubContentArn fwtypes.ARN                                   `tfsdk:"sagemaker_public_hub_content_arn"`
	Tags                         tftags.Map                                    `tfsdk:"tags"`
	TagsAll                      tftags.Map                                    `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                `tfsdk:"timeouts"`
}
