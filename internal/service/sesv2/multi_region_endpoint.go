// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_sesv2_multi_region_endpoint", name="Multi Region Endpoint")
func newMultiRegionEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &multiRegionEndpointResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameMultiRegionEndpoint = "Multi Region Endpoint"
)

type multiRegionEndpointResource struct {
	framework.ResourceWithModel[multiRegionEndpointResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *multiRegionEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"routes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[routeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRegion: schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"details": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[detailsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"routes_details": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routeDetailsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRegion: schema.StringAttribute{
										Required: true,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *multiRegionEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var plan multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	detailsList, d := plan.Details.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiDetails *awstypes.Details
	if len(detailsList) > 0 {
		routesDetailsList, d2 := detailsList[0].RoutesDetails.ToSlice(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d2)
		if resp.Diagnostics.HasError() {
			return
		}
		rd := make([]awstypes.RouteDetails, len(routesDetailsList))
		for i, r := range routesDetailsList {
			rd[i] = awstypes.RouteDetails{
				Region: fwflex.StringFromFramework(ctx, r.Region),
			}
		}
		apiDetails = &awstypes.Details{RoutesDetails: rd}
	}

	input := sesv2.CreateMultiRegionEndpointInput{
		EndpointName: fwflex.StringFromFramework(ctx, plan.EndpointName),
		Details:      apiDetails,
	}

	out, err := conn.CreateMultiRegionEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.EndpointName.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.EndpointName.String())
		return
	}

	plan.ID = plan.EndpointName
	plan.EndpointID = fwflex.StringToFramework(ctx, out.EndpointId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	ep, err := waitMultiRegionEndpointCreated(ctx, conn, plan.EndpointName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.EndpointName.String())
		return
	}

	plan.Status = fwtypes.StringEnumValue(ep.Status)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *multiRegionEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findMultiRegionEndpointByName(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	state.EndpointID = fwflex.StringToFramework(ctx, out.EndpointId)
	state.EndpointName = fwflex.StringToFramework(ctx, out.EndpointName)
	state.Status = fwtypes.StringEnumValue(out.Status)

	routes := make([]routeModel, len(out.Routes))
	for i, r := range out.Routes {
		routes[i] = routeModel{
			Region: fwflex.StringToFramework(ctx, r.Region),
		}
	}
	var routesDiags diag.Diagnostics
	state.Routes, routesDiags = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, routes)
	smerr.AddEnrich(ctx, &resp.Diagnostics, routesDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *multiRegionEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sesv2.DeleteMultiRegionEndpointInput{
		EndpointName: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteMultiRegionEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitMultiRegionEndpointDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

const (
	statusCreating = string(awstypes.StatusCreating)
	statusReady    = string(awstypes.StatusReady)
	statusFailed   = string(awstypes.StatusFailed)
	statusDeleting = string(awstypes.StatusDeleting)
)

func waitMultiRegionEndpointCreated(ctx context.Context, conn *sesv2.Client, name string, timeout time.Duration) (*sesv2.GetMultiRegionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreating},
		Target:                    []string{statusReady},
		Refresh:                   statusMultiRegionEndpoint(conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sesv2.GetMultiRegionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMultiRegionEndpointDeleted(ctx context.Context, conn *sesv2.Client, name string, timeout time.Duration) (*sesv2.GetMultiRegionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusReady},
		Target:  []string{},
		Refresh: statusMultiRegionEndpoint(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sesv2.GetMultiRegionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMultiRegionEndpoint(conn *sesv2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findMultiRegionEndpointByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findMultiRegionEndpointByName(ctx context.Context, conn *sesv2.Client, name string) (*sesv2.GetMultiRegionEndpointOutput, error) {
	input := sesv2.GetMultiRegionEndpointInput{
		EndpointName: aws.String(name),
	}

	out, err := conn.GetMultiRegionEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type multiRegionEndpointResourceModel struct {
	EndpointID   types.String                                  `tfsdk:"endpoint_id"`
	EndpointName types.String                                  `tfsdk:"endpoint_name"`
	Details      fwtypes.ListNestedObjectValueOf[detailsModel] `tfsdk:"details"`
	Routes       fwtypes.ListNestedObjectValueOf[routeModel]   `tfsdk:"routes"`
	Status       fwtypes.StringEnum[awstypes.Status]           `tfsdk:"status"`
	Timeouts     timeouts.Value                                `tfsdk:"timeouts"`
}

type routeModel struct {
	Region types.String `tfsdk:"region"`
}

type detailsModel struct {
	RoutesDetails fwtypes.ListNestedObjectValueOf[routeDetailsModel] `tfsdk:"routes_details"`
}

type routeDetailsModel struct {
	Region types.String `tfsdk:"region"`
}

func sweepMultiRegionEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sesv2.ListMultiRegionEndpointsInput{}
	conn := client.SESV2Client(ctx)
	var sweepResources []sweep.Sweepable

	pages := sesv2.NewListMultiRegionEndpointsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.MultiRegionEndpoints {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newMultiRegionEndpointResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.EndpointName))),
			)
		}
	}

	return sweepResources, nil
}
