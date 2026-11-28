// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecs

import (
	"context"
	"iter"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_gateway_cluster_association", name="Gateway Cluster Association")
// @IdentityAttribute("gateway_arn")
// @IdentityAttribute("cluster_arn")
// @ImportIDHandler("gatewayClusterAssociationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="gateway_arn;cluster_arn", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(preCheck="testAccPreCheckGateways")
func newGatewayClusterAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayClusterAssociationResource{}

	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type gatewayClusterAssociationResource struct {
	framework.ResourceWithModel[gatewayClusterAssociationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *gatewayClusterAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"gateway_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *gatewayClusterAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan gatewayClusterAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	gatewayARN, clusterARN := fwflex.StringValueFromFramework(ctx, plan.GatewayARN), fwflex.StringValueFromFramework(ctx, plan.ClusterARN)
	id := gatewayClusterAssociationID(gatewayARN, clusterARN)
	var input ecs.AssociateGatewayWithClusterInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNameSuffix("ARN")))
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := conn.AssociateGatewayWithCluster(ctx, &input); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.GatewayClusterAssociationSummary, error) {
		return findGatewayClusterAssociationByTwoPartKey(ctx, conn, gatewayARN, clusterARN)
	}); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
	}
}

func (r *gatewayClusterAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayClusterAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	gatewayARN, clusterARN := fwflex.StringValueFromFramework(ctx, data.GatewayARN), fwflex.StringValueFromFramework(ctx, data.ClusterARN)
	association, err := findGatewayClusterAssociationByTwoPartKey(ctx, conn, gatewayARN, clusterARN)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayClusterAssociationID(gatewayARN, clusterARN))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, association, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayClusterAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayClusterAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	gatewayARN, clusterARN := fwflex.StringValueFromFramework(ctx, data.GatewayARN), fwflex.StringValueFromFramework(ctx, data.ClusterARN)
	id := gatewayClusterAssociationID(gatewayARN, clusterARN)
	association, err := findGatewayClusterAssociationByTwoPartKey(ctx, conn, gatewayARN, clusterARN)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if association.Status == awstypes.GatewayClusterAssociationStatusActive {
		var input ecs.DisassociateGatewayFromClusterInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNameSuffix("ARN")))
		if response.Diagnostics.HasError() {
			return
		}

		if _, err := conn.DisassociateGatewayFromCluster(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	if _, err := waitGatewayClusterAssociationDeleted(ctx, conn, gatewayARN, clusterARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
	}
}

func gatewayClusterAssociationID(gatewayARN, clusterARN string) string {
	return strings.Join([]string{gatewayARN, clusterARN}, intflex.ResourceIdSeparator)
}

var (
	_ inttypes.ImportIDParser = gatewayClusterAssociationImportID{}
)

type gatewayClusterAssociationImportID struct{}

func (gatewayClusterAssociationImportID) Parse(id string) (string, map[string]any, error) {
	const (
		gatewayClusterAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, gatewayClusterAssociationIDParts, false)
	if err != nil {
		return "", nil, smarterr.NewError(err)
	}

	return id, map[string]any{
		"gateway_arn": parts[0],
		"cluster_arn": parts[1],
	}, nil
}

func findGatewayClusterAssociationByTwoPartKey(ctx context.Context, conn *ecs.Client, gatewayARN, clusterARN string) (*awstypes.GatewayClusterAssociationSummary, error) {
	input := ecs.ListGatewayClusterAssociationsInput{
		Cluster: aws.String(clusterARN),
		Gateway: aws.String(gatewayARN),
	}

	output, err := tfslices.CollectAndConcatWithError(listGatewayClusterAssociationPages(ctx, conn, &input), tfslices.WithFilter(func(v awstypes.GatewayClusterAssociationSummary) bool {
		return aws.ToString(v.GatewayArn) == gatewayARN && aws.ToString(v.ClusterArn) == clusterARN && v.Status != awstypes.GatewayClusterAssociationStatusInactive
	}))
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func listGatewayClusterAssociationPages(ctx context.Context, conn *ecs.Client, input *ecs.ListGatewayClusterAssociationsInput) iter.Seq2[[]awstypes.GatewayClusterAssociationSummary, error] {
	return func(yield func([]awstypes.GatewayClusterAssociationSummary, error) bool) {
		pages := ecs.NewListGatewayClusterAssociationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.GatewayClusterAssociations, nil) {
				return
			}
		}
	}
}

func statusGatewayClusterAssociation(conn *ecs.Client, gatewayARN, clusterARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGatewayClusterAssociationByTwoPartKey(ctx, conn, gatewayARN, clusterARN)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitGatewayClusterAssociationDeleted(ctx context.Context, conn *ecs.Client, gatewayARN, clusterARN string, timeout time.Duration) (*awstypes.GatewayClusterAssociationSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayClusterAssociationStatusActive, awstypes.GatewayClusterAssociationStatusDisassociating),
		Target:  []string{},
		Refresh: statusGatewayClusterAssociation(conn, gatewayARN, clusterARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.GatewayClusterAssociationSummary); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type gatewayClusterAssociationResourceModel struct {
	framework.WithRegionModel
	ClusterARN fwtypes.ARN    `tfsdk:"cluster_arn"`
	GatewayARN fwtypes.ARN    `tfsdk:"gateway_arn"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
