// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_continuum_collector", name="Collector")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum;continuum.GetCollectorOutput")
func newCollectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &collectorResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type collectorResource struct {
	framework.ResourceWithModel[collectorResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *collectorResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"config_recorder_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"kubernetes_cluster_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Description: "ARN of the IAM role that Continuum assumes to read Kubernetes cluster context.",
			},
			"kubernetes_cluster_role_external_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "External ID that Continuum presents when assuming `kubernetes_cluster_role_arn`. The service does not return this value.",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CollectorType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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

func (r *collectorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan collectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	input := continuum.CreateCollectorInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
	}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		input.Type = plan.Type.ValueEnum()
	}
	if !plan.KubernetesClusterRoleARN.IsNull() {
		input.KubernetesClusterRole = &awstypes.KubernetesClusterRoleConfiguration{
			RoleArn:    plan.KubernetesClusterRoleARN.ValueStringPointer(),
			ExternalId: plan.KubernetesClusterRoleExternalID.ValueStringPointer(),
		}
	}

	output, err := conn.CreateCollector(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
	if output == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"))
		return
	}

	id := aws.ToString(output.CollectorId)
	plan.ID = types.StringValue(id)

	collector, err := waitCollectorActive(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	plan.flatten(collector)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state collectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	output, err := findCollectorByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	state.flatten(output)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *collectorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state collectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	if !plan.KubernetesClusterRoleARN.Equal(state.KubernetesClusterRoleARN) || !plan.KubernetesClusterRoleExternalID.Equal(state.KubernetesClusterRoleExternalID) {
		input := continuum.UpdateCollectorInput{
			CollectorId: plan.ID.ValueStringPointer(),
		}
		if plan.KubernetesClusterRoleARN.IsNull() {
			input.KubernetesClusterRole = &awstypes.KubernetesClusterRoleUpdateMemberRemove{}
		} else {
			input.KubernetesClusterRole = &awstypes.KubernetesClusterRoleUpdateMemberSet{
				Value: awstypes.KubernetesClusterRoleConfiguration{
					RoleArn:    plan.KubernetesClusterRoleARN.ValueStringPointer(),
					ExternalId: plan.KubernetesClusterRoleExternalID.ValueStringPointer(),
				},
			}
		}

		if _, err := conn.UpdateCollector(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	output, err := findCollectorByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	plan.flatten(output)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state collectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	input := continuum.DeleteCollectorInput{
		CollectorId: state.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteCollector(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	if _, err := waitCollectorDeleted(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func findCollectorByID(ctx context.Context, conn *continuum.Client, id string) (*continuum.GetCollectorOutput, error) {
	input := continuum.GetCollectorInput{
		CollectorId: aws.String(id),
	}

	return findCollector(ctx, conn, &input)
}

func findCollector(ctx context.Context, conn *continuum.Client, input *continuum.GetCollectorInput) (*continuum.GetCollectorOutput, error) {
	output, err := conn.GetCollector(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusCollector(conn *continuum.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCollectorByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitCollectorActive(ctx context.Context, conn *continuum.Client, id string, timeout time.Duration) (*continuum.GetCollectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Target:  enum.Slice(awstypes.CollectorStatusActive),
		Refresh: statusCollector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*continuum.GetCollectorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitCollectorDeleted(ctx context.Context, conn *continuum.Client, id string, timeout time.Duration) (*continuum.GetCollectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CollectorStatusPendingDeletion, awstypes.CollectorStatusActive),
		Target:  []string{},
		Refresh: statusCollector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*continuum.GetCollectorOutput); ok {
		return output, err
	}

	return nil, err
}

type collectorResourceModel struct {
	framework.WithRegionModel
	ARN                             types.String                               `tfsdk:"arn"`
	ConfigRecorderARN               types.String                               `tfsdk:"config_recorder_arn"`
	CreatedAt                       timetypes.RFC3339                          `tfsdk:"created_at"`
	ID                              types.String                               `tfsdk:"id"`
	KubernetesClusterRoleARN        fwtypes.ARN                                `tfsdk:"kubernetes_cluster_role_arn"`
	KubernetesClusterRoleExternalID types.String                               `tfsdk:"kubernetes_cluster_role_external_id"`
	Status                          types.String                               `tfsdk:"status"`
	Timeouts                        timeouts.Value                             `tfsdk:"timeouts"`
	Type                            fwtypes.StringEnum[awstypes.CollectorType] `tfsdk:"type"`
	UpdatedAt                       timetypes.RFC3339                          `tfsdk:"updated_at"`
}

func (m *collectorResourceModel) flatten(output *continuum.GetCollectorOutput) {
	m.ARN = types.StringPointerValue(output.CollectorArn)
	m.ConfigRecorderARN = types.StringPointerValue(output.ConfigRecorderArn)
	m.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	m.ID = types.StringPointerValue(output.CollectorId)
	if output.KubernetesClusterRoleArn != nil {
		m.KubernetesClusterRoleARN = fwtypes.ARNValue(aws.ToString(output.KubernetesClusterRoleArn))
	} else {
		m.KubernetesClusterRoleARN = fwtypes.ARNNull()
		m.KubernetesClusterRoleExternalID = types.StringNull()
	}
	m.Status = types.StringValue(string(output.Status))
	m.Type = fwtypes.StringEnumValue(output.Type)
	m.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)
}
