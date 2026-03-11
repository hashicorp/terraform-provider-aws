// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package amp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
	r.SetDefaultUpdateTimeout(2 * time.Minute)

	return r, nil
}

type scraperResource struct {
	framework.ResourceWithModel[scraperResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *scraperResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAlias: schema.StringAttribute{
				Optional: true,
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
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrDestination: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[destinationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"amp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ampConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"workspace_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"role_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[roleConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_role_arn": schema.StringAttribute{
							Optional:   true,
							CustomType: fwtypes.ARNType,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("target_role_arn"),
								),
							},
						},
						"target_role_arn": schema.StringAttribute{
							Optional:   true,
							CustomType: fwtypes.ARNType,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("source_role_arn"),
								),
							},
						},
					},
				},
			},
			names.AttrSource: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceModel](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[eksConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
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
				Update: true,
			}),
		},
	}
}

func (r *scraperResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data scraperResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	var input amp.CreateScraperInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.ScrapeConfiguration = &awstypes.ScrapeConfigurationMemberConfigurationBlob{
		Value: []byte(data.ScrapeConfiguration.ValueString()),
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateScraper(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Prometheus Scraper", err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	data.ID = fwflex.StringToFramework(ctx, output.ScraperId)

	scraper, err := waitScraperCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	data.RoleARN = fwflex.StringToFramework(ctx, scraper.RoleArn)
	response.Diagnostics.Append(fwflex.Flatten(ctx, scraper.Source, &data.Source)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *scraperResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data scraperResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	scraper, err := findScraperByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Prometheus Scraper (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, scraper, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if v, ok := scraper.ScrapeConfiguration.(*awstypes.ScrapeConfigurationMemberConfigurationBlob); ok {
		data.ScrapeConfiguration = fwflex.StringValueToFramework(ctx, string(v.Value))
	}

	setTagsOut(ctx, scraper.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *scraperResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old scraperResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	if !new.Alias.Equal(old.Alias) ||
		!new.Destination.Equal(old.Destination) ||
		!new.RoleConfiguration.Equal(old.RoleConfiguration) ||
		!new.ScrapeConfiguration.Equal(old.ScrapeConfiguration) {
		var input amp.UpdateScraperInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())
		input.ScrapeConfiguration = &awstypes.ScrapeConfigurationMemberConfigurationBlob{
			Value: []byte(new.ScrapeConfiguration.ValueString()),
		}
		input.ScraperId = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateScraper(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Prometheus Scraper (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitScraperUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *scraperResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data scraperResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	input := amp.DeleteScraperInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		ScraperId:   fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteScraper(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Prometheus Scraper (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitScraperDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

type scraperResourceModel struct {
	framework.WithRegionModel
	Alias               types.String                                            `tfsdk:"alias"`
	ARN                 types.String                                            `tfsdk:"arn"`
	Destination         fwtypes.ListNestedObjectValueOf[destinationModel]       `tfsdk:"destination"`
	ID                  types.String                                            `tfsdk:"id"`
	RoleARN             types.String                                            `tfsdk:"role_arn"`
	RoleConfiguration   fwtypes.ListNestedObjectValueOf[roleConfigurationModel] `tfsdk:"role_configuration"`
	ScrapeConfiguration types.String                                            `tfsdk:"scrape_configuration" autoflex:"-"`
	Source              fwtypes.ListNestedObjectValueOf[sourceModel]            `tfsdk:"source"`
	Tags                tftags.Map                                              `tfsdk:"tags"`
	TagsAll             tftags.Map                                              `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                          `tfsdk:"timeouts"`
}

type destinationModel struct {
	AMP fwtypes.ListNestedObjectValueOf[ampConfigurationModel] `tfsdk:"amp"`
}

var (
	_ fwflex.Expander  = destinationModel{}
	_ fwflex.Flattener = &destinationModel{}
)

func (m destinationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v awstypes.Destination

	switch {
	case !m.AMP.IsNull():
		data, d := m.AMP.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.DestinationMemberAmpConfiguration
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	}

	return v, diags
}

func (m *destinationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.DestinationMemberAmpConfiguration:
		var data ampConfigurationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.AMP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

type ampConfigurationModel struct {
	WorkspaceARN fwtypes.ARN `tfsdk:"workspace_arn"`
}

type sourceModel struct {
	EKS fwtypes.ListNestedObjectValueOf[eksConfigurationModel] `tfsdk:"eks"`
}

var (
	_ fwflex.Expander  = sourceModel{}
	_ fwflex.Flattener = &sourceModel{}
)

func (m sourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v awstypes.Source

	switch {
	case !m.EKS.IsNull():
		data, d := m.EKS.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.SourceMemberEksConfiguration
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	}

	return v, diags
}

func (m *sourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.SourceMemberEksConfiguration:
		var data eksConfigurationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.EKS = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

type eksConfigurationModel struct {
	ClusterARN       fwtypes.ARN         `tfsdk:"cluster_arn"`
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetOfString `tfsdk:"subnet_ids"`
}

type roleConfigurationModel struct {
	SourceRoleARN fwtypes.ARN `tfsdk:"source_role_arn"`
	TargetRoleARN fwtypes.ARN `tfsdk:"target_role_arn"`
}

func findScraperByID(ctx context.Context, conn *amp.Client, id string) (*awstypes.ScraperDescription, error) {
	input := amp.DescribeScraperInput{
		ScraperId: aws.String(id),
	}

	return findScraper(ctx, conn, &input)
}

func findScraper(ctx context.Context, conn *amp.Client, input *amp.DescribeScraperInput) (*awstypes.ScraperDescription, error) {
	output, err := conn.DescribeScraper(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Scraper == nil || output.Scraper.Status == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Scraper, nil
}

func statusScraper(conn *amp.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findScraperByID(ctx, conn, id)

		if retry.NotFound(err) {
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
		Refresh: statusScraper(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScraperDescription); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitScraperUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperStatusCodeUpdating),
		Target:  enum.Slice(awstypes.ScraperStatusCodeActive),
		Refresh: statusScraper(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScraperDescription); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitScraperDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperStatusCodeActive, awstypes.ScraperStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusScraper(conn, id),
		Timeout: timeout,
		Delay:   8 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScraperDescription); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}
