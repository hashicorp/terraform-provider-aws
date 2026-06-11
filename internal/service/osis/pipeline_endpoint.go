// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_osis_pipeline_endpoint", name="Pipeline Endpoint")
func newPipelineEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &pipelineEndpointResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)
	return r, nil
}

type pipelineEndpointResource struct {
	framework.ResourceWithModel[pipelineEndpointResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *pipelineEndpointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"pipeline_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"vpc_options": schema.ListNestedBlock{
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[pipelineEndpointVPCOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
							},
						},
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
							},
						},
					},
				}},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *pipelineEndpointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data pipelineEndpointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := osis.CreatePipelineEndpointInput{}

	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreatePipelineEndpoint(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating OpenSearch Ingestion Pipeline Endpoint", err.Error())
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, aws.ToString(output.EndpointId))

	endpoint, err := waitPipelineEndpointCreated(ctx, conn, aws.ToString(output.EndpointId), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(output.EndpointId))...)
		response.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Ingestion Pipeline Endpoint (%s) create", aws.ToString(output.EndpointId)), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, endpoint, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Status = fwflex.StringValueToFramework(ctx, string(endpoint.Status))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineEndpointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data pipelineEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	endpointID := data.ID.ValueString()
	endpoint, err := findPipelineEndpointByID(ctx, conn, endpointID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Ingestion Pipeline Endpoint (%s)", endpointID), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, endpoint, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineEndpointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data pipelineEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	endpointID := data.ID.ValueString()
	input := &osis.DeletePipelineEndpointInput{
		EndpointId: aws.String(endpointID),
	}

	_, err := conn.DeletePipelineEndpoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Ingestion Pipeline Endpoint (%s)", endpointID), err.Error())
		return
	}

	if _, err := waitPipelineEndpointDeleted(ctx, conn, endpointID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Ingestion Pipeline Endpoint (%s) delete", endpointID), err.Error())
		return
	}
}

// findPipelineEndpointByID retrieves a pipeline endpoint by its ID.
func findPipelineEndpointByID(ctx context.Context, conn *osis.Client, endpointID string) (*awstypes.PipelineEndpoint, error) {
	input := &osis.ListPipelineEndpointsInput{
		MaxResults: aws.Int32(100),
	}

	var endpoint *awstypes.PipelineEndpoint
	pages := osis.NewListPipelineEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, e := range page.PipelineEndpoints {
			if aws.ToString(e.EndpointId) == endpointID {
				endpoint = &e
				return endpoint, nil
			}
		}
	}

	if endpoint == nil {
		return nil, &sdkretry.NotFoundError{
			LastRequest: input,
		}
	}

	return endpoint, nil
}

func statusPipelineEndpoint(ctx context.Context, conn *osis.Client, endpointID string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPipelineEndpointByID(ctx, conn, endpointID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPipelineEndpointCreated(ctx context.Context, conn *osis.Client, endpointID string, timeout time.Duration) (*awstypes.PipelineEndpoint, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineEndpointStatusCreating),
		Target:     enum.Slice(awstypes.PipelineEndpointStatusActive),
		Refresh:    statusPipelineEndpoint(ctx, conn, endpointID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PipelineEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitPipelineEndpointDeleted(ctx context.Context, conn *osis.Client, endpointID string, timeout time.Duration) (*awstypes.PipelineEndpoint, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineEndpointStatusDeleting),
		Target:     []string{},
		Refresh:    statusPipelineEndpoint(ctx, conn, endpointID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PipelineEndpoint); ok {
		return output, err
	}

	return nil, err
}

type pipelineEndpointResourceModel struct {
	framework.WithRegionModel
	ID          types.String                                                     `tfsdk:"id"`
	PipelineARN fwtypes.ARN                                                      `tfsdk:"pipeline_arn"`
	Status      types.String                                                     `tfsdk:"status"`
	VPCID       types.String                                                     `tfsdk:"vpc_id"`
	VPCOptions  fwtypes.ListNestedObjectValueOf[pipelineEndpointVPCOptionsModel] `tfsdk:"vpc_options"`
	Timeouts    timeouts.Value                                                   `tfsdk:"timeouts"`
}

type pipelineEndpointVPCOptionsModel struct {
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
}
