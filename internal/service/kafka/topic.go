// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kafka;kafka.DescribeTopicOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccTopicImportStateIDFunc")
func newTopicResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &topicResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

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
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configs": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
			},
			"configs_actual": schema.StringAttribute{
				// configs_actual is only for display purposes of all config on the topic, also outside 'configs'
				Computed: true,
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

	clusterARN, topicName := fwflex.StringValueFromFramework(ctx, plan.ClusterARN), fwflex.StringValueFromFramework(ctx, plan.TopicName)
	var input kafka.CreateTopicInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Configs.IsNull() {
		// Configs is base64encoded in the AWS API
		input.Configs = aws.String(inttypes.Base64Encode([]byte(fwflex.StringValueFromFramework(ctx, plan.Configs))))
	}

	_, err := conn.CreateTopic(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
		return
	}

	out, err := waitTopicCreated(ctx, conn, clusterARN, topicName, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
		return
	}

	v, diags := flattenTopicConfigsActual(ctx, out.Configs)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ConfigsActual = v
	plan.TopicARN = fwflex.StringToFramework(ctx, out.TopicArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *topicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().KafkaClient(ctx)

	var state topicResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	clusterARN, topicName := fwflex.StringValueFromFramework(ctx, state.ClusterARN), fwflex.StringValueFromFramework(ctx, state.TopicName)
	out, err := findTopicByTwoPartKey(ctx, conn, clusterARN, topicName)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, clusterARN, topicName)
		return
	}

	importing := state.TopicARN.IsNull()

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// The Configs returned from the API contains server-augmented values.
	// The resource's ConfigsActual contains all values whilst Config contains only client-configured values.
	v, diags := flattenTopicConfigsActual(ctx, out.Configs)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ConfigsActual = v

	if !state.Configs.IsNull() {
		var serverConfigs map[string]any
		if err := tfjson.DecodeFromString(fwflex.StringValueFromFramework(ctx, state.ConfigsActual), &serverConfigs); err != nil {
			resp.Diagnostics.AddError("JSON decoding server configs", err.Error())
			return
		}
		var clientConfigs map[string]any
		if err := tfjson.DecodeFromString(fwflex.StringValueFromFramework(ctx, state.Configs), &clientConfigs); err != nil {
			resp.Diagnostics.AddError("JSON decoding client configs", err.Error())
			return
		}

		for k := range clientConfigs {
			if v, ok := serverConfigs[k]; ok {
				clientConfigs[k] = v
			} else {
				delete(clientConfigs, k)
			}
		}

		v, err := tfjson.EncodeToString(clientConfigs)
		if err != nil {
			resp.Diagnostics.AddError("JSON encoding client configs", err.Error())
			return
		}

		state.Configs = jsontypes.NewNormalizedValue(v)
	} else if importing {
		state.Configs = jsontypes.Normalized{StringValue: state.ConfigsActual}
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

	clusterARN, topicName := fwflex.StringValueFromFramework(ctx, plan.ClusterARN), fwflex.StringValueFromFramework(ctx, plan.TopicName)

	// 'configs' and 'partition_count' are updated via separate API calls:
	// "You must specify either configs or partitionCount to update."
	if plan.Configs != state.Configs {
		input := kafka.UpdateTopicInput{
			ClusterArn: aws.String(clusterARN),
			TopicName:  aws.String(topicName),
		}
		if !plan.Configs.IsNull() {
			// Configs is base64encoded in the AWS API
			input.Configs = aws.String(inttypes.Base64Encode([]byte(fwflex.StringValueFromFramework(ctx, plan.Configs))))
		} else {
			input.Configs = aws.String(inttypes.Base64Encode([]byte("{}")))
		}
		_, err := conn.UpdateTopic(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
			return
		}

		out, err := waitTopicUpdated(ctx, conn, clusterARN, topicName, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
			return
		}

		v, diags := flattenTopicConfigsActual(ctx, out.Configs)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ConfigsActual = v
	}

	if plan.PartitionCount != state.PartitionCount {
		input := kafka.UpdateTopicInput{
			ClusterArn:     aws.String(clusterARN),
			PartitionCount: fwflex.Int32FromFrameworkInt64(ctx, plan.PartitionCount),
			TopicName:      aws.String(topicName),
		}
		_, err := conn.UpdateTopic(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
			return
		}

		out, err := waitTopicUpdated(ctx, conn, clusterARN, topicName, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
			return
		}

		v, diags := flattenTopicConfigsActual(ctx, out.Configs)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ConfigsActual = v
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

	clusterARN, topicName := fwflex.StringValueFromFramework(ctx, state.ClusterARN), fwflex.StringValueFromFramework(ctx, state.TopicName)
	input := kafka.DeleteTopicInput{
		ClusterArn: aws.String(clusterARN),
		TopicName:  aws.String(topicName),
	}
	_, err := conn.DeleteTopic(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
		return
	}

	if _, err := waitTopicDeleted(ctx, conn, clusterARN, topicName, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, topicName)
		return
	}
}

func waitTopicCreated(ctx context.Context, conn *kafka.Client, clusterARN, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Target:                    enum.Slice(awstypes.TopicStateActive),
		Refresh:                   statusTopic(conn, clusterARN, topicName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTopicUpdated(ctx context.Context, conn *kafka.Client, clusterARN, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TopicStateUpdating),
		Target:                    enum.Slice(awstypes.TopicStateActive),
		Refresh:                   statusTopic(conn, clusterARN, topicName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTopicDeleted(ctx context.Context, conn *kafka.Client, clusterARN, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TopicStateDeleting, awstypes.TopicStateActive),
		Target:  []string{},
		Refresh: statusTopic(conn, clusterARN, topicName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return out, err
	}

	return nil, err
}

func statusTopic(conn *kafka.Client, clusterARN, topicName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTopicByTwoPartKey(ctx, conn, clusterARN, topicName)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findTopicByTwoPartKey(ctx context.Context, conn *kafka.Client, clusterARN, topicName string) (*kafka.DescribeTopicOutput, error) {
	input := kafka.DescribeTopicInput{
		ClusterArn: aws.String(clusterARN),
		TopicName:  aws.String(topicName),
	}

	return findTopic(ctx, conn, &input)
}

func findTopic(ctx context.Context, conn *kafka.Client, input *kafka.DescribeTopicInput) (*kafka.DescribeTopicOutput, error) {
	output, err := conn.DescribeTopic(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TopicName == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func flattenTopicConfigsActual(ctx context.Context, configs *string) (types.String, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// Configs is base64encoded in the AWS API
	if configs != nil {
		v, err := inttypes.Base64Decode(aws.ToString(configs))
		if err != nil {
			diags.AddError("base64 decoding configs", err.Error())
			return types.StringNull(), diags
		}

		return fwflex.StringValueToFramework(ctx, v), diags
	}

	return types.StringNull(), diags
}

type topicResourceModel struct {
	framework.WithRegionModel
	ClusterARN        fwtypes.ARN          `tfsdk:"cluster_arn"`
	Configs           jsontypes.Normalized `tfsdk:"configs" autoflex:"-"`
	ConfigsActual     types.String         `tfsdk:"configs_actual" autoflex:"-"`
	PartitionCount    types.Int64          `tfsdk:"partition_count"`
	ReplicationFactor types.Int64          `tfsdk:"replication_factor"`
	Timeouts          timeouts.Value       `tfsdk:"timeouts"`
	TopicARN          types.String         `tfsdk:"arn"`
	TopicName         types.String         `tfsdk:"name"`
}

var (
	_ inttypes.ImportIDParser = topicImportID{}
)

type topicImportID struct{}

func (topicImportID) Parse(id string) (string, map[string]any, error) {
	const (
		topicIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, topicIDParts, true)

	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"cluster_arn":  parts[0],
		names.AttrName: parts[1],
	}

	return id, result, nil
}
