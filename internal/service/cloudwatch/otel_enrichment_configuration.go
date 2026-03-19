// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkResource("aws_cloudwatch_otel_enrichment_configuration", name="OTel Enrichment Configuration")
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
func newOtelEnrichmentConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &otelEnrichmentConfigurationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOtelEnrichmentConfiguration = "OTel Enrichment Configuration"
)

type otelEnrichmentConfigurationResource struct {
	framework.ResourceWithModel[otelEnrichmentConfigurationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *otelEnrichmentConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *otelEnrichmentConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var plan otelEnrichmentConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Enabled.ValueBool() {
		input := &cloudwatch.EnableOTelEnrichmentInput{}
		_, err := conn.EnableOTelEnrichment(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError("enabling OTel enrichment", err.Error())
			return
		}
	} else {
		input := &cloudwatch.DisableOTelEnrichmentInput{}
		_, err := conn.DisableOTelEnrichment(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError("disabling OTel enrichment", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *otelEnrichmentConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var state otelEnrichmentConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	input := &cloudwatch.GetOTelEnrichmentConfigurationInput{}
	out, err := conn.GetOTelEnrichmentConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("reading OTel enrichment configuration", err.Error())
		return
	}
	
	state.Enabled = types.BoolValue(*out.Enabled)
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *otelEnrichmentConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var state otelEnrichmentConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	input := &cloudwatch.DisableOTelEnrichmentInput{}
	_, err := conn.DisableOTelEnrichment(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("disabling OTel enrichment", err.Error())
		return
	}
}

type otelEnrichmentConfigurationResourceModel struct {
	framework.WithRegionModel
	Enabled  types.Bool     `tfsdk:"enabled"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
