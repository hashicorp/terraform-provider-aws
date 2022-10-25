package swf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func init() {
	registerFrameworkResourceFactory(newResourceDomain)
}

// newResourceDomain instantiates a new Resource for the aws_swf_domain resource.
func newResourceDomain(context.Context) (intf.ResourceWithConfigureAndImportState, error) {
	return &resourceDomain{}, nil
}

type resourceDomain struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (r *resourceDomain) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_swf_domain"
}

// GetSchema returns the schema for this resource.
func (r *resourceDomain) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"arn": {
				Type:     types.StringType,
				Computed: true,
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"name_prefix": {
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputed(),
			"workflow_execution_retention_period_in_days": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
				// TODO Validate,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *resourceDomain) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceDomain) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceDomainData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.String{Value: "TODO"}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *resourceDomain) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceDomainData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceDomain) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data resourceDomainData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically call DeleteResponse.State.RemoveResource(),
// so it can be omitted from provider logic.
func (r *resourceDomain) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceDomainData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting TODO", map[string]interface{}{
		"id": data.ID.Value,
	})
}

// ImportState is called when the provider must import the state of a resource instance.
// This method must return enough state so the Read method can properly refresh the full resource.
//
// If setting an attribute with the import identifier, it is recommended to use the ImportStatePassthroughID() call in this method.
func (r *resourceDomain) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type resourceDomainData struct {
	ARN                                    types.String `tfsdk:"arn"`
	Description                            types.String `tfsdk:"description"`
	ID                                     types.String `tfsdk:"id"`
	Name                                   types.String `tfsdk:"name"`
	NamePrefix                             types.String `tfsdk:"name_prefix"`
	Tags                                   types.Map    `tfsdk:"tags"`
	TagsAll                                types.Map    `tfsdk:"tags_all"`
	WorkflowExecutionRetentionPeriodInDays types.String `tfsdk:"workflow_execution_retention_period_in_days"`
}
