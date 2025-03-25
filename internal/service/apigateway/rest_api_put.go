// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_api_gateway_rest_api_put", name="Rest API Put")
func newResourceRestAPIPut(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRestAPIPut{}

	r.SetDefaultCreateTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameRestAPIPut = "Rest API Put"
)

type resourceRestAPIPut struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithNoUpdate
}

func (r *resourceRestAPIPut) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"body": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fail_on_warnings": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			names.AttrParameters: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"rest_api_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTriggers: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *resourceRestAPIPut) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().APIGatewayClient(ctx)

	var plan resourceRestAPIPutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apigateway.PutRestApiInput{
		Mode: awstypes.PutModeOverwrite,
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutRestApi(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGateway, create.ErrActionCreating, ResNameRestAPIPut, plan.RestAPIID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGateway, create.ErrActionCreating, ResNameRestAPIPut, plan.RestAPIID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = waitRestAPIPutCreated(ctx, conn, plan.RestAPIID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGateway, create.ErrActionWaitingForCreation, ResNameRestAPIPut, plan.RestAPIID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRestAPIPut) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceRestAPIPutModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	out, err := findRestAPIByID(ctx, conn, state.RestAPIID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGateway, create.ErrActionSetting, ResNameRestAPIPut, state.RestAPIID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRestAPIPut) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rest_api_id"), req, resp)
}

const (
	statusNormal = "Normal"
)

func waitRestAPIPutCreated(ctx context.Context, conn *apigateway.Client, id string, timeout time.Duration) (*apigateway.GetRestApiOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusRestAPIPut(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*apigateway.GetRestApiOutput); ok {
		return out, err
	}

	return nil, err
}

func statusRestAPIPut(ctx context.Context, conn *apigateway.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findRestAPIByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// There does not seem to be a good way to figure the status of the rest api.
		// Regardless, with the waiter, this gives a little time for ops to finish.

		return out, statusNormal, nil
	}
}

type resourceRestAPIPutModel struct {
	Body           types.String        `tfsdk:"body"`
	FailOnWarnings types.Bool          `tfsdk:"fail_on_warnings"`
	Parameters     fwtypes.MapOfString `tfsdk:"parameters"`
	RestAPIID      types.String        `tfsdk:"rest_api_id"`
	Timeouts       timeouts.Value      `tfsdk:"timeouts"`
	Triggers       fwtypes.MapOfString `tfsdk:"triggers"`
}
