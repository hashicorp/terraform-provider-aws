// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshiftserverless_snapshot_copy_configuration", name="Snapshot Copy Configuration")
func newResourceSnapshotCopyConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSnapshotCopyConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSnapshotCopyConfiguration = "Snapshot Copy Configuration"
)

type resourceSnapshotCopyConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceSnapshotCopyConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_kms_key_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"namespace_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_retention_period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
		},
	}
}

func (r *resourceSnapshotCopyConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan snapshotCopyConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input redshiftserverless.CreateSnapshotCopyConfigurationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateSnapshotCopyConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionCreating, ResNameSnapshotCopyConfiguration, plan.NamespaceName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.SnapshotCopyConfiguration == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionCreating, ResNameSnapshotCopyConfiguration, plan.NamespaceName.String(), nil),
			errors.New("empty out").Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.SnapshotCopyConfiguration, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSnapshotCopyConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state snapshotCopyConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSnapshotCopyConfigurationByID(ctx, conn, state.NamespaceName.ValueString(), state.SnapshotCopyConfigurationId.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionSetting, ResNameSnapshotCopyConfiguration, state.SnapshotCopyConfigurationId.String(), err),
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

func (r *resourceSnapshotCopyConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan, state snapshotCopyConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input redshiftserverless.UpdateSnapshotCopyConfigurationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateSnapshotCopyConfiguration(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionUpdating, ResNameSnapshotCopyConfiguration, plan.SnapshotCopyConfigurationId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.SnapshotCopyConfiguration == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionUpdating, ResNameSnapshotCopyConfiguration, plan.SnapshotCopyConfigurationId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSnapshotCopyConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state snapshotCopyConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := redshiftserverless.DeleteSnapshotCopyConfigurationInput{
		SnapshotCopyConfigurationId: state.SnapshotCopyConfigurationId.ValueStringPointer(),
	}
	_, err := conn.DeleteSnapshotCopyConfiguration(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionDeleting, ResNameSnapshotCopyConfiguration, state.SnapshotCopyConfigurationId.String(), err),
			err.Error(),
		)
		return
	}
}

func findSnapshotCopyConfigurationByID(ctx context.Context, conn *redshiftserverless.Client, namespaceName string, id string) (*awstypes.SnapshotCopyConfiguration, error) {
	input := &redshiftserverless.ListSnapshotCopyConfigurationsInput{
		NamespaceName: aws.String(namespaceName),
	}

	var configuration *awstypes.SnapshotCopyConfiguration

	// iterate over snapshot copy configuration paginator
	p := redshiftserverless.NewListSnapshotCopyConfigurationsPaginator(conn, input)
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range out.SnapshotCopyConfigurations {
			if aws.ToString(item.SnapshotCopyConfigurationId) == id {
				cpy := item
				configuration = &cpy
				break
			}
		}

		if configuration != nil {
			break
		}
	}

	if configuration == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}
	return configuration, nil
}

type snapshotCopyConfigurationResourceModel struct {
	DestinationKmsKeyId          types.String `tfsdk:"destination_kms_key_id"`
	DestinationRegion            types.String `tfsdk:"destination_region"`
	NamespaceName                types.String `tfsdk:"namespace_name"`
	SnapshotCopyConfigurationArn fwtypes.ARN  `tfsdk:"arn"`
	SnapshotCopyConfigurationId  types.String `tfsdk:"id"`
	SnapshotRetentionPeriod      types.Int64  `tfsdk:"snapshot_retention_period"`
}
