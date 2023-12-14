// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Scraper")
// @Tags(identifierAttribute="arn")
func newResourceScraper(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceScraper{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameScraper = "Scraper"
)

type resourceScraper struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceScraper) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_prometheus_scraper"
}

func (r *resourceScraper) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alias": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"scrape_configuration": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags: schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"destination": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"amp": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"workspace_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"source": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"eks": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"cluster_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"security_group_ids": schema.SetAttribute{
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
											setplanmodifier.UseStateForUnknown(),
										},
									},
									"subnet_ids": schema.SetAttribute{
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Delete: true,
	})

	response.Schema = s
}

func (r *resourceScraper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AMPClient(ctx)

	var plan resourceScraperModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &amp.CreateScraperInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Source:      expandSource(ctx, plan.Source, resp.Diagnostics),
		Destination: expandDestination(ctx, plan.Destination, resp.Diagnostics),
		ScrapeConfiguration: &awstypes.ScrapeConfigurationMemberConfigurationBlob{
			Value: []byte(plan.ScrapeConfiguration.ValueString()),
		},
		Tags: getTagsInV2(ctx),
	}

	if !plan.Alias.IsNull() {
		in.Alias = aws.String(plan.Alias.ValueString())
	}

	out, err := conn.CreateScraper(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionCreating, ResNameScraper, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ScraperId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionCreating, ResNameScraper, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = flex.StringToFramework(ctx, out.ScraperId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitScraperCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionWaitingForCreation, ResNameScraper, "", err),
			err.Error(),
		)
		return
	}

	readOut, _ := findScraperByID(ctx, conn, *out.ScraperId)
	plan.refreshFromOutput(ctx, readOut, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceScraper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceScraperModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	out, err := findScraperByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionSetting, ResNameScraper, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	state.refreshFromOutput(ctx, out, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceScraper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceScraperModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceScraper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AMPClient(ctx)

	var state resourceScraperModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &amp.DeleteScraperInput{
		ScraperId:   aws.String(state.ID.ValueString()),
		ClientToken: aws.String(sdkid.UniqueId()),
	}

	_, err := conn.DeleteScraper(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionDeleting, ResNameScraper, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitScraperDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionWaitingForDeletion, ResNameScraper, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceScraper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (state *resourceScraper) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	state.SetTagsAll(ctx, request, response)
}

func findScraperByID(ctx context.Context, conn *amp.Client, id string) (*awstypes.ScraperDescription, error) {
	input := &amp.DescribeScraperInput{
		ScraperId: aws.String(id),
	}

	output, err := conn.DescribeScraper(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Scraper == nil || output.Scraper.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Scraper, nil
}

func flattenDestination(ctx context.Context, apiObject awstypes.Destination, diags diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: destinationModelAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType)
	}

	ampDestination, ok := apiObject.(*awstypes.DestinationMemberAmpConfiguration)
	if !ok {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{
		"amp": flattenDestinationAMPConfig(ctx, ampDestination.Value, diags),
	}
	objVal, d := types.ObjectValue(destinationModelAttrTypes, attrs)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal
}

func flattenDestinationAMPConfig(ctx context.Context, apiObject awstypes.AmpConfiguration, diags diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: ampDestinationModelAttrTypes}

	attrs := map[string]attr.Value{
		"workspace_arn": fwtypes.ARNValue(*apiObject.WorkspaceArn),
	}
	objVal, d := types.ObjectValue(ampDestinationModelAttrTypes, attrs)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal
}

func flattenSource(ctx context.Context, apiObject awstypes.Source, diags diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: sourceModelAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType)
	}

	eksSource, ok := apiObject.(*awstypes.SourceMemberEksConfiguration)
	if !ok {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{
		"eks": flattenSourceEKSConfig(ctx, eksSource.Value, diags),
	}
	objVal, d := types.ObjectValue(sourceModelAttrTypes, attrs)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal
}

func flattenSourceEKSConfig(ctx context.Context, apiObject awstypes.EksConfiguration, diags diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: eksSourceModelAttrTypes}

	attrs := map[string]attr.Value{
		"cluster_arn":        fwtypes.ARNValue(*apiObject.ClusterArn),
		"subnet_ids":         flex.FlattenFrameworkStringValueSet(ctx, apiObject.SubnetIds),
		"security_group_ids": flex.FlattenFrameworkStringValueSet(ctx, apiObject.SecurityGroupIds),
	}
	objVal, d := types.ObjectValue(eksSourceModelAttrTypes, attrs)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal
}

func expandDestination(ctx context.Context, dst types.List, diags diag.Diagnostics) awstypes.Destination {

	var tfList []destinationModel
	diags.Append(dst.ElementsAs(ctx, &tfList, false)...)

	if len(tfList) == 0 {
		return nil
	}
	tfObj := tfList[0]

	var ampDestination []ampDestinationModel
	diags.Append(tfObj.AMP.ElementsAs(ctx, &ampDestination, false)...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.DestinationMemberAmpConfiguration{Value: expandAMPDestination(ampDestination)}
}

func expandAMPDestination(tfList []ampDestinationModel) awstypes.AmpConfiguration {
	if len(tfList) == 0 {
		return awstypes.AmpConfiguration{}
	}

	tfObj := tfList[0]
	ampConfig := awstypes.AmpConfiguration{
		WorkspaceArn: tfObj.AWSPrometheusWorkspaceARN.ValueStringPointer(),
	}

	return ampConfig
}
func expandSource(ctx context.Context, src types.List, diags diag.Diagnostics) awstypes.Source {

	var tfList []sourceModel
	diags.Append(src.ElementsAs(ctx, &tfList, false)...)

	if len(tfList) == 0 {
		return nil
	}
	tfObj := tfList[0]

	var eksSource []eksSourceModel
	diags.Append(tfObj.EKS.ElementsAs(ctx, &eksSource, false)...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.SourceMemberEksConfiguration{Value: expandEKSSource(ctx, eksSource)}
}

func expandEKSSource(ctx context.Context, tfList []eksSourceModel) awstypes.EksConfiguration {

	if len(tfList) == 0 {
		return awstypes.EksConfiguration{}
	}

	tfObj := tfList[0]
	eksSource := awstypes.EksConfiguration{
		ClusterArn: tfObj.EKSClusterARN.ValueStringPointer(),
		SubnetIds:  flex.ExpandFrameworkStringValueSet(ctx, tfObj.SubnetIds),
	}

	if !tfObj.SecurityGroupIds.IsNull() {
		eksSource.SecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, tfObj.SecurityGroupIds)
	}

	return eksSource
}

type resourceScraperModel struct {
	Alias               types.String   `tfsdk:"alias"`
	ARN                 types.String   `tfsdk:"arn"`
	Destination         types.List     `tfsdk:"destination"`
	ID                  types.String   `tfsdk:"id"`
	ScrapeConfiguration types.String   `tfsdk:"scrape_configuration"`
	Source              types.List     `tfsdk:"source"`
	Tags                types.Map      `tfsdk:"tags"`
	TagsAll             types.Map      `tfsdk:"tags_all"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

type destinationModel struct {
	AMP types.List `tfsdk:"amp"`
}

type ampDestinationModel struct {
	AWSPrometheusWorkspaceARN fwtypes.ARN `tfsdk:"workspace_arn"`
}

type sourceModel struct {
	EKS types.List `tfsdk:"eks"`
}

type eksSourceModel struct {
	EKSClusterARN    fwtypes.ARN `tfsdk:"cluster_arn"`
	SubnetIds        types.Set   `tfsdk:"subnet_ids"`
	SecurityGroupIds types.Set   `tfsdk:"security_group_ids"`
}

var destinationModelAttrTypes = map[string]attr.Type{
	"amp": types.ListType{ElemType: types.ObjectType{AttrTypes: ampDestinationModelAttrTypes}},
}
var ampDestinationModelAttrTypes = map[string]attr.Type{
	"workspace_arn": fwtypes.ARNType,
}

var sourceModelAttrTypes = map[string]attr.Type{
	"eks": types.ListType{ElemType: types.ObjectType{AttrTypes: eksSourceModelAttrTypes}},
}
var eksSourceModelAttrTypes = map[string]attr.Type{
	"cluster_arn":        fwtypes.ARNType,
	"subnet_ids":         types.SetType{ElemType: types.StringType},
	"security_group_ids": types.SetType{ElemType: types.StringType},
}

func (state *resourceScraperModel) refreshFromOutput(ctx context.Context, out *awstypes.ScraperDescription, diags diag.Diagnostics) {

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.ScraperId)
	state.Alias = flex.StringToFramework(ctx, out.Alias)
	if scrapeCfg, ok := out.ScrapeConfiguration.(*awstypes.ScrapeConfigurationMemberConfigurationBlob); ok {
		state.ScrapeConfiguration = flex.StringValueToFramework(ctx, string(scrapeCfg.Value))
	}

	setTagsOutV2(ctx, out.Tags)
	state.Destination = flattenDestination(ctx, out.Destination, diags)
	state.Source = flattenSource(ctx, out.Source, diags)
}
