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
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_hub_content_reference", name="Hub Content Reference")
// @Tags(identifierAttribute="hub_content_arn")
// Both hub_content_arn and sagemaker_public_hub_content_arn are stripped ARNs. AWS APIs append the min_version at the end of both of them. The ListTags API expects the hub_content_arn to not contain this min_version.
// @IdentityAttribute("hub_name")
// @IdentityAttribute("hub_content_name")
// @ImportIDHandler("hubContentReferenceImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccHubContentReferenceImportStateIDFunc)
// @Testing(importStateIdAttribute="hub_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeHubContentOutput")
func newHubContentReferenceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hubContentReferenceResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameHubContentReference = "Hub Content Reference"
)

type hubContentReferenceResource struct {
	framework.ResourceWithModel[hubContentReferenceResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
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
				Description: "ARN of the hub content reference (without version suffix). The min_version is stripped off from the end of this ARN to make it usable to list tags.",
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
				Description: "Status of the hub content reference. Valid values include `Available`, `Importing`, `Deleting`, `ImportFailed`, `DeleteFailed`.",
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
				Description: "Minimum version of the hub content to reference. Use \"1.0.0\" to support all versions. Changing this value to an empty string forces replacement of the resource.",
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
				Description: "ARN of the public SageMaker JumpStart hub content to reference. The ARN must not include a version suffix.",
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
				Update: true,
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

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

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

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *hubContentReferenceResource) flatten(ctx context.Context, output *sagemaker.DescribeHubContentOutput, data *hubContentReferenceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, output, data)...)
	if diags.HasError() {
		return diags
	}

	data.HubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.HubContentArn))
	data.MinVersion = fwflex.StringToFramework(ctx, output.ReferenceMinVersion)
	data.SageMakerPublicHubContentArn = fwflex.StringToFrameworkARN(ctx, stripARNVersion(output.SageMakerPublicHubContentArn))

	return diags
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

		output, err := waitHubContentReferenceAvailable(ctx, conn, new.HubName.ValueString(), new.HubContentName.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.HubContentName.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &new))
		if response.Diagnostics.HasError() {
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

var versionSuffix = regexache.MustCompile(`/\d+\.\d+\.\d+$`)

// stripARNVersion removes the semantic version suffix from a SageMaker hub content ARN.
// If no version suffix is present, the ARN is returned unchanged.
func stripARNVersion(arn *string) *string {
	return aws.String(versionSuffix.ReplaceAllString(aws.ToString(arn), ""))
}

const hubContentReferenceImportIDSeparator = intflex.ResourceIdSeparator

func hubContentReferenceParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, hubContentReferenceImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected hub-name%[2]shub-content-name", id, hubContentReferenceImportIDSeparator)
}

var _ inttypes.ImportIDParser = hubContentReferenceImportID{}

type hubContentReferenceImportID struct{}

func (hubContentReferenceImportID) Parse(id string) (string, map[string]any, error) {
	hubName, hubContentName, err := hubContentReferenceParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"hub_name":         hubName,
		"hub_content_name": hubContentName,
	}

	return id, result, nil
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
