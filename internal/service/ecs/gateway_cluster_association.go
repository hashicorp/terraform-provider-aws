// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_gateway_cluster_association", name="Gateway Cluster Association")
// @IdentityAttribute("cluster")
// @IdentityAttribute("gateway")
// @ImportIDHandler("gatewayClusterAssociationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccGatewayClusterAssociationImportStateIDFunc")
// @Testing(importStateIdAttribute="cluster")
func newGatewayClusterAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayClusterAssociationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type gatewayClusterAssociationResource struct {
	framework.ResourceWithModel[gatewayClusterAssociationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
	framework.WithNoUpdate
}

func (r *gatewayClusterAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"gateway_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *gatewayClusterAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan gatewayClusterAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	cluster, gateway := plan.Cluster.ValueString(), plan.Gateway.ValueString()
	input := ecs.AssociateGatewayWithClusterInput{
		Cluster: aws.String(cluster),
		Gateway: aws.String(gateway),
	}
	_, err := conn.AssociateGatewayWithCluster(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Gateway Cluster Association (%s,%s)", cluster, gateway), err.Error())
		return
	}

	association, err := waitGatewayClusterAssociationActive(ctx, conn, cluster, gateway, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Gateway Cluster Association (%s,%s) create", cluster, gateway), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, association, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *gatewayClusterAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state gatewayClusterAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	cluster, gateway := state.Cluster.ValueString(), state.Gateway.ValueString()
	association, err := findGatewayClusterAssociationByTwoPartKey(ctx, conn, cluster, gateway)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Gateway Cluster Association (%s,%s)", cluster, gateway), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, association, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *gatewayClusterAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state gatewayClusterAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	cluster, gateway := state.Cluster.ValueString(), state.Gateway.ValueString()
	input := ecs.DisassociateGatewayFromClusterInput{
		Cluster: aws.String(cluster),
		Gateway: aws.String(gateway),
	}
	_, err := conn.DisassociateGatewayFromCluster(ctx, &input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.GatewayNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Gateway Cluster Association (%s,%s)", cluster, gateway), err.Error())
		return
	}

	if _, err := waitGatewayClusterAssociationDeleted(ctx, conn, cluster, gateway, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Gateway Cluster Association (%s,%s) delete", cluster, gateway), err.Error())
		return
	}
}

func findGatewayClusterAssociationByTwoPartKey(ctx context.Context, conn *ecs.Client, cluster, gateway string) (*awstypes.GatewayClusterAssociationSummary, error) {
	input := ecs.ListGatewayClusterAssociationsInput{
		Cluster: aws.String(cluster),
		Gateway: aws.String(gateway),
	}
	output, err := findGatewayClusterAssociations(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	association, err := tfresource.AssertSingleValueResult(output)

	if err != nil {
		return nil, err
	}

	if status := association.Status; status == awstypes.GatewayClusterAssociationStatusInactive {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return association, nil
}

func findGatewayClusterAssociations(ctx context.Context, conn *ecs.Client, input *ecs.ListGatewayClusterAssociationsInput) ([]awstypes.GatewayClusterAssociationSummary, error) {
	var output []awstypes.GatewayClusterAssociationSummary

	pages := ecs.NewListGatewayClusterAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.GatewayNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.GatewayClusterAssociations...)
	}

	return output, nil
}

func statusGatewayClusterAssociation(conn *ecs.Client, cluster, gateway string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGatewayClusterAssociationByTwoPartKey(ctx, conn, cluster, gateway)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGatewayClusterAssociationActive(ctx context.Context, conn *ecs.Client, cluster, gateway string, timeout time.Duration) (*awstypes.GatewayClusterAssociationSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{},
		Target:         enum.Slice(awstypes.GatewayClusterAssociationStatusActive),
		Refresh:        statusGatewayClusterAssociation(conn, cluster, gateway),
		Timeout:        timeout,
		NotFoundChecks: 20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GatewayClusterAssociationSummary); ok {
		return output, err
	}

	return nil, err
}

func waitGatewayClusterAssociationDeleted(ctx context.Context, conn *ecs.Client, cluster, gateway string, timeout time.Duration) (*awstypes.GatewayClusterAssociationSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayClusterAssociationStatusActive, awstypes.GatewayClusterAssociationStatusDisassociating),
		Target:  []string{},
		Refresh: statusGatewayClusterAssociation(conn, cluster, gateway),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GatewayClusterAssociationSummary); ok {
		return output, err
	}

	return nil, err
}

type gatewayClusterAssociationResourceModel struct {
	framework.WithRegionModel
	Cluster    types.String   `tfsdk:"cluster"`
	ClusterARN fwtypes.ARN    `tfsdk:"cluster_arn"`
	Gateway    types.String   `tfsdk:"gateway"`
	GatewayARN fwtypes.ARN    `tfsdk:"gateway_arn"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

const gatewayClusterAssociationIDSeparator = ","

var (
	_ inttypes.ImportIDParser = gatewayClusterAssociationImportID{}
)

type gatewayClusterAssociationImportID struct{}

func (gatewayClusterAssociationImportID) Parse(id string) (string, map[string]any, error) {
	cluster, gateway, found := strings.Cut(id, gatewayClusterAssociationIDSeparator)
	if !found || cluster == "" || gateway == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <cluster>%s<gateway>", id, gatewayClusterAssociationIDSeparator)
	}

	result := map[string]any{
		"cluster": cluster,
		"gateway": gateway,
	}

	return id, result, nil
}
