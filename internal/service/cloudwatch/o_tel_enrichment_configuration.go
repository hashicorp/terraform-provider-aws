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
// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_cloudwatch_otel_enrichment_configuration", name="OTel Enrichment Configuration")
// @SingletonIdentity

// TIP: ==== RESOURCE IDENTITY ====
// Identify which attributes can be used to uniquely identify the resource.
// 
// * If the AWS APIs for the resource take the ARN as an identifier, use
// ARN Identity.
// * If the resource is a singleton (i.e., there is only one instance per region, or account for global resource types), use Singleton Identity.
// * Otherwise, use Parameterized Identity with one or more identity attributes.
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch;cloudwatch.DescribeOTelEnrichmentConfigurationResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
func newOTelEnrichmentConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &oTelEnrichmentConfigurationResource{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOTelEnrichmentConfiguration = "O Tel Enrichment Configuration"
)

type oTelEnrichmentConfigurationResource struct {
	framework.ResourceWithModel[oTelEnrichmentConfigurationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}


// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
// * If a user can provide a value ("configure a value") for an
//   attribute (e.g., instances = 5), we call the attribute an
//   "argument."
// * You change the way users interact with attributes using:
//     - Required
//     - Optional
//     - Computed
// * There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
// 2. Optional only - the user can configure or omit a value; do not
//    use Default or DefaultFunc
// Optional: true,
//
// 3. Computed only - the provider can provide a value but the user
//    cannot, i.e., read-only
// Computed: true,
//
// 4. Optional AND Computed - the provider or user can provide a value;
//    use this combination if you are using Default
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *oTelEnrichmentConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *oTelEnrichmentConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var plan oTelEnrichmentConfigurationResourceModel
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

func (r *oTelEnrichmentConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var state oTelEnrichmentConfigurationResourceModel
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

func (r *oTelEnrichmentConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)
	
	var state oTelEnrichmentConfigurationResourceModel
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

type oTelEnrichmentConfigurationResourceModel struct {
	framework.WithRegionModel
	Enabled  types.Bool     `tfsdk:"enabled"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
