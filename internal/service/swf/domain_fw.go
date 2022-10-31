package swf

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/fwvalidators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
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
					resource.UseStateForUnknown(),
				},
			},
			"name_prefix": {
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					resource.UseStateForUnknown(),
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
				Validators: []tfsdk.AttributeValidator{
					fwvalidators.Int64StringBetween(0, 90),
				},
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

	conn := r.meta.SWFConn
	defaultTagsConfig := r.meta.DefaultTagsConfig
	ignoreTagsConfig := r.meta.IgnoreTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(data.Tags))

	var name, namePrefix string

	if !data.Name.IsNull() {
		name = data.Name.Value
	}
	if !data.NamePrefix.IsNull() {
		namePrefix = data.NamePrefix.Value
	}
	name = create.Name(name, namePrefix)
	input := &swf.RegisterDomainInput{
		Name:                                   aws.String(name),
		Tags:                                   Tags(tags.IgnoreAWS()),
		WorkflowExecutionRetentionPeriodInDays: aws.String(data.WorkflowExecutionRetentionPeriodInDays.Value),
	}

	if !data.Description.IsNull() {
		input.Description = aws.String(data.Description.Value)
	}

	_, err := conn.RegisterDomainWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SWF Domain (%s)", name), err.Error())

		return
	}

	data.ID = types.String{Value: name}

	output, err := FindDomainByName(ctx, conn, data.ID.Value)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SWF Domain (%s)", data.ID.Value), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = types.String{Value: aws.StringValue(output.DomainInfo.Arn)}
	if data.Name.IsNull() {
		data.Name = types.String{Value: aws.StringValue(output.DomainInfo.Name)}
	}
	if data.NamePrefix.IsNull() {
		data.NamePrefix = types.String{Value: aws.StringValue(create.NamePrefixFromName(aws.StringValue(output.DomainInfo.Name)))}
	}
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

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

	conn := r.meta.SWFConn
	defaultTagsConfig := r.meta.DefaultTagsConfig
	ignoreTagsConfig := r.meta.IgnoreTagsConfig

	output, err := FindDomainByName(ctx, conn, data.ID.Value)

	if tfresource.NotFound(err) {
		tflog.Warn(ctx, "SWF Domain not found, removing from state", map[string]interface{}{
			"id": data.ID.Value,
		})
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SWF Domain (%s)", data.ID.Value), err.Error())

		return
	}

	arn := aws.StringValue(output.DomainInfo.Arn)
	data.ARN = types.String{Value: arn}
	data.Description = types.String{Value: aws.StringValue(output.DomainInfo.Description)}
	data.Name = types.String{Value: aws.StringValue(output.DomainInfo.Name)}
	data.NamePrefix = types.String{Value: aws.StringValue(create.NamePrefixFromName(aws.StringValue(output.DomainInfo.Name)))}
	data.WorkflowExecutionRetentionPeriodInDays = types.String{Value: aws.StringValue(output.Configuration.WorkflowExecutionRetentionPeriodInDays)}

	tags, err := ListTagsWithContext(ctx, conn, arn)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing tags for SWF Domain (%s)", arn), err.Error())

		return
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	data.Tags = flex.FlattenFrameworkStringValueMap(ctx, tags.RemoveDefaultConfig(defaultTagsConfig).Map())
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceDomain) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceDomainData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.meta.SWFConn

	if !new.TagsAll.Equal(old.TagsAll) {
		if err := UpdateTagsWithContext(ctx, conn, new.ARN.Value, old.TagsAll, new.TagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SWF Domain (%s) tags", new.ID.Value), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
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

	conn := r.meta.SWFConn

	tflog.Debug(ctx, "deleting SWF Domain", map[string]interface{}{
		"id": data.ID.Value,
	})
	_, err := conn.DeprecateDomainWithContext(ctx, &swf.DeprecateDomainInput{
		Name: aws.String(data.ID.Value),
	})

	if tfawserr.ErrCodeEquals(err, swf.ErrCodeDomainDeprecatedFault, swf.ErrCodeUnknownResourceFault) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SWF Domain (%s)", data.ID.Value), err.Error())

		return
	}

	_, err = tfresource.RetryUntilNotFoundContext(ctx, 1*time.Minute, func() (interface{}, error) {
		return FindDomainByName(ctx, conn, data.ID.Value)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SWF Domain (%s) delete", data.ID.Value), err.Error())

		return
	}
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
