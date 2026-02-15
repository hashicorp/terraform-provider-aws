// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_msk_topic", name="Topic")
// @IdentityAttribute("name")
// @IdentityAttribute("cluster_arn")
// @ImportIDHandler(topicImportID)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kafka;kafka.DescribeTopicResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newTopicResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &topicResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTopic = "Topic"
)

type topicResource struct {
	framework.ResourceWithModel[topicResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *topicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cluster_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"partition_count": schema.Int64Attribute{
				Required: true,
			},
			"replication_factor": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"configs": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
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

func (r *topicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var plan topicResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input kafka.CreateTopicInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Topic")))
	if resp.Diagnostics.HasError() {
		return
	}

	outPartial, err := conn.CreateTopic(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if outPartial == nil || outPartial.TopicName == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outPartial, &plan, flex.WithFieldNamePrefix("Topic")))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outFull, err := waitTopicCreated(ctx, conn, plan.Name.ValueString(), plan.ClusterARN.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outFull, &plan, flex.WithFieldNamePrefix("Topic")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *topicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state topicResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTopicByTwoPartKey(ctx, conn, state.Name.ValueString(), state.ClusterARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String(), state.ClusterARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Topic")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *topicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var plan, state topicResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input kafka.UpdateTopicInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Topic")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateTopic(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil || out.TopicName == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("Topic")))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitTopicUpdated(ctx, conn, plan.Name.ValueString(), plan.ClusterARN.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *topicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state topicResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := kafka.DeleteTopicInput{
		TopicName:  state.Name.ValueStringPointer(),
		ClusterArn: state.ClusterARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTopic(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTopicDeleted(ctx, conn, state.Name.ValueString(), state.ClusterARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func waitTopicCreated(ctx context.Context, conn *kafka.Client, name string, clusterARN string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.TopicStateActive),
		Refresh:                   statusTopic(conn, name, clusterARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTopicUpdated(ctx context.Context, conn *kafka.Client, name string, clusterARN string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TopicStateUpdating),
		Target:                    enum.Slice(awstypes.TopicStateActive),
		Refresh:                   statusTopic(conn, name, clusterARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTopicDeleted(ctx context.Context, conn *kafka.Client, name string, clusterARN string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TopicStateDeleting, awstypes.TopicStateActive),
		Target:  []string{},
		Refresh: statusTopic(conn, name, clusterARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusTopic(conn *kafka.Client, name string, clusterARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTopicByTwoPartKey(ctx, conn, name, clusterARN)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findTopicByTwoPartKey(ctx context.Context, conn *kafka.Client, name string, clusterARN string) (*kafka.DescribeTopicOutput, error) {
	input := kafka.DescribeTopicInput{
		TopicName:  aws.String(name),
		ClusterArn: aws.String(clusterARN),
	}

	out, err := conn.DescribeTopic(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TopicName == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type topicResourceModel struct {
	framework.WithRegionModel
	ARN               types.String   `tfsdk:"arn"`
	ClusterARN        types.String   `tfsdk:"cluster_arn"`
	Name              types.String   `tfsdk:"name"`
	PartitionCount    types.Int64    `tfsdk:"partition_count"`
	ReplicationFactor types.Int64    `tfsdk:"replication_factor"`
	Configs           types.String   `tfsdk:"configs"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ inttypes.ImportIDParser = topicImportID{}
)

type topicImportID struct{}

func (topicImportID) Parse(id string) (string, map[string]string, error) {
	name, clusterARN, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <topic name>"+intflex.ResourceIdSeparator+"<cluster ARN>", id)
	}

	result := map[string]string{
		names.AttrName: name,
		"cluster_arn":  clusterARN,
	}

	return id, result, nil
}
