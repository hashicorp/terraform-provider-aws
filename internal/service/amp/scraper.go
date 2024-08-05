// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_prometheus_scraper", name="Scraper")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp/types;types.ScraperDescription")
func newScraperResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &scraperResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameScraper = "Scraper"
)

type scraperResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *scraperResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_prometheus_scraper"
}

func (r *scraperResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAlias: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrRoleARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scrape_configuration": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrDestination: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scraperDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"amp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scraperAMPDestinationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
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
			names.AttrSource: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scraperSourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"eks": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scraperEKSSourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
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
									names.AttrSecurityGroupIDs: schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
											setplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrSubnetIDs: schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *scraperResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AMPClient(ctx)

	var data scraperResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We can't use AutoFlEx with the top-level resource model because the API structure uses Go interfaces.

	destinationData, diags := data.Destination.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ampDestinationData, diags := destinationData.AMP.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	destination := &awstypes.DestinationMemberAmpConfiguration{}
	resp.Diagnostics.Append(flex.Expand(ctx, ampDestinationData, &destination.Value)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceData, diags := data.Source.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	eksSourceData, diags := sourceData.EKS.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	source := &awstypes.SourceMemberEksConfiguration{}
	resp.Diagnostics.Append(flex.Expand(ctx, eksSourceData, &source.Value)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &amp.CreateScraperInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Destination: destination,
		Source:      source,
		ScrapeConfiguration: &awstypes.ScrapeConfigurationMemberConfigurationBlob{
			Value: []byte(data.ScrapeConfiguration.ValueString()),
		},
		Tags: getTagsIn(ctx),
	}

	if !data.Alias.IsNull() {
		input.Alias = aws.String(data.Alias.ValueString())
	}

	output, err := conn.CreateScraper(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionCreating, ResNameScraper, "", err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	data.ARN = flex.StringToFramework(ctx, output.Arn)
	data.ID = flex.StringToFramework(ctx, output.ScraperId)

	scraper, err := waitScraperCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionWaitingForCreation, ResNameScraper, "", err),
			err.Error(),
		)
		return
	}

	if v, ok := scraper.Source.(*awstypes.SourceMemberEksConfiguration); ok {
		resp.Diagnostics.Append(flex.Flatten(ctx, &v.Value, eksSourceData)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set values for unknowns after creation is complete.
	sourceData.EKS = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, eksSourceData)
	data.RoleARN = flex.StringToFramework(ctx, scraper.RoleArn)
	data.Source = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, sourceData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scraperResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AMPClient(ctx)

	var data scraperResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scraper, err := findScraperByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionSetting, ResNameScraper, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	// We can't use AutoFlEx with the top-level resource model because the API structure uses Go interfaces.
	data.ARN = flex.StringToFramework(ctx, scraper.Arn)
	data.Alias = flex.StringToFramework(ctx, scraper.Alias)
	if v, ok := scraper.Destination.(*awstypes.DestinationMemberAmpConfiguration); ok {
		var ampDestinationData scraperAMPDestinationModel
		resp.Diagnostics.Append(flex.Flatten(ctx, &v.Value, &ampDestinationData)...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Destination = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &scraperDestinationModel{
			AMP: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ampDestinationData),
		})
	}
	data.RoleARN = flex.StringToFramework(ctx, scraper.RoleArn)
	if v, ok := scraper.ScrapeConfiguration.(*awstypes.ScrapeConfigurationMemberConfigurationBlob); ok {
		data.ScrapeConfiguration = flex.StringValueToFramework(ctx, string(v.Value))
	}
	if v, ok := scraper.Source.(*awstypes.SourceMemberEksConfiguration); ok {
		var eksSourceData scraperEKSSourceModel
		resp.Diagnostics.Append(flex.Flatten(ctx, &v.Value, &eksSourceData)...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Source = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &scraperSourceModel{
			EKS: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &eksSourceData),
		})
	}

	setTagsOut(ctx, scraper.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scraperResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new scraperResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags only.

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *scraperResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AMPClient(ctx)

	var data scraperResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteScraper(ctx, &amp.DeleteScraperInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		ScraperId:   aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionDeleting, ResNameScraper, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitScraperDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AMP, create.ErrActionWaitingForDeletion, ResNameScraper, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *scraperResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

type scraperResourceModel struct {
	Alias               types.String                                             `tfsdk:"alias"`
	ARN                 types.String                                             `tfsdk:"arn"`
	Destination         fwtypes.ListNestedObjectValueOf[scraperDestinationModel] `tfsdk:"destination"`
	ID                  types.String                                             `tfsdk:"id"`
	RoleARN             types.String                                             `tfsdk:"role_arn"`
	ScrapeConfiguration types.String                                             `tfsdk:"scrape_configuration"`
	Source              fwtypes.ListNestedObjectValueOf[scraperSourceModel]      `tfsdk:"source"`
	Tags                types.Map                                                `tfsdk:"tags"`
	TagsAll             types.Map                                                `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                           `tfsdk:"timeouts"`
}

type scraperDestinationModel struct {
	AMP fwtypes.ListNestedObjectValueOf[scraperAMPDestinationModel] `tfsdk:"amp"`
}

type scraperAMPDestinationModel struct {
	WorkspaceARN fwtypes.ARN `tfsdk:"workspace_arn"`
}

type scraperSourceModel struct {
	EKS fwtypes.ListNestedObjectValueOf[scraperEKSSourceModel] `tfsdk:"eks"`
}

type scraperEKSSourceModel struct {
	ClusterARN       fwtypes.ARN                      `tfsdk:"cluster_arn"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
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

func statusScraper(ctx context.Context, conn *amp.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findScraperByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitScraperCreated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperStatusCodeCreating),
		Target:  enum.Slice(awstypes.ScraperStatusCodeActive),
		Refresh: statusScraper(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*awstypes.ScraperDescription); ok {
		return out, err
	}

	return nil, err
}

func waitScraperDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperStatusCodeActive, awstypes.ScraperStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusScraper(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScraperDescription); ok {
		return output, err
	}

	return nil, err
}
