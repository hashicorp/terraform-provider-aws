package resourceexplorer2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	registerFrameworkResourceFactory(newResourceIndex)
}

// newResourceSecurityGroupIngressRule instantiates a new Resource for the aws_resourceexplorer2_index resource.
func newResourceIndex(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceIndex{}, nil
}

type resourceIndex struct {
	framework.ResourceWithConfigure
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (r *resourceIndex) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_resourceexplorer2_index"
}

// GetSchema returns the schema for this resource.
func (r *resourceIndex) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"arn": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"id":       framework.IDAttribute(),
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
			"type": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					enum.FrameworkValidate[awstypes.IndexType](),
				},
			},
		},
	}

	return schema, nil
}

// TODO Timeouts

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceIndex) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceIndexData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(data.Tags))

	input := &resourceexplorer2.CreateIndexInput{
		ClientToken: aws.String(sdkresource.UniqueId()),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateIndex(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Resource Explorer Index", err.Error())

		return
	}

	arn := aws.ToString(output.Arn)
	data.ID = types.StringValue(arn)

	if _, err := waitIndexCreated(ctx, conn, -1); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Resource Explorer Index (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	if data.Type.ValueString() == string(awstypes.IndexTypeAggregator) {
		// TODO AGGREGATOR type
	}

	// Set values for unknowns.
	data.ARN = types.StringValue(arn)
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *resourceIndex) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceIndexData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig

	output, err := FindIndex(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Resource Explorer Index (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = flex.StringToFramework(ctx, output.Arn)
	data.Type = types.StringValue(string(output.Type))

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// AWS APIs often return empty lists of tags when none have been configured.
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 {
		data.Tags = tftags.Null
	} else {
		data.Tags = flex.FlattenFrameworkStringValueMap(ctx, tags)
	}
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceIndex) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically call DeleteResponse.State.RemoveResource(),
// so it can be omitted from provider logic.
func (r *resourceIndex) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceIndexData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client

	tflog.Debug(ctx, "deleting Resource Explorer Index", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	_, err := conn.DeleteIndex(ctx, &resourceexplorer2.DeleteIndexInput{
		Arn: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Resource Explorer Index (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitIndexDeleted(ctx, conn, -1); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Resource Explorer Index (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

// ImportState is called when the provider must import the state of a resource instance.
// This method must return enough state so the Read method can properly refresh the full resource.
//
// If setting an attribute with the import identifier, it is recommended to use the ImportStatePassthroughID() call in this method.
func (r *resourceIndex) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type resourceIndexData struct {
	ARN     types.String `tfsdk:"arn"`
	ID      types.String `tfsdk:"id"`
	Tags    types.Map    `tfsdk:"tags"`
	TagsAll types.Map    `tfsdk:"tags_all"`
	Type    types.String `tfsdk:"type"`
}

func FindIndex(ctx context.Context, conn *resourceexplorer2.Client) (*resourceexplorer2.GetIndexOutput, error) {
	input := &resourceexplorer2.GetIndexInput{}

	output, err := conn.GetIndex(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &sdkresource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if state := output.State; state == awstypes.IndexStateDeleted {
		return nil, &sdkresource.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusIndex(ctx context.Context, conn *resourceexplorer2.Client) sdkresource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIndex(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitIndexCreated(ctx context.Context, conn *resourceexplorer2.Client, timeout time.Duration) (*resourceexplorer2.GetIndexOutput, error) {
	stateConf := &sdkresource.StateChangeConf{
		Pending: enum.Slice(awstypes.IndexStateCreating),
		Target:  enum.Slice(awstypes.IndexStateActive),
		Refresh: statusIndex(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*resourceexplorer2.GetIndexOutput); ok {
		return output, err
	}

	return nil, err
}

func waitIndexUpdated(ctx context.Context, conn *resourceexplorer2.Client, timeout time.Duration) (*resourceexplorer2.GetIndexOutput, error) {
	stateConf := &sdkresource.StateChangeConf{
		Pending: enum.Slice(awstypes.IndexStateUpdating),
		Target:  enum.Slice(awstypes.IndexStateActive),
		Refresh: statusIndex(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*resourceexplorer2.GetIndexOutput); ok {
		return output, err
	}

	return nil, err
}

func waitIndexDeleted(ctx context.Context, conn *resourceexplorer2.Client, timeout time.Duration) (*resourceexplorer2.GetIndexOutput, error) {
	stateConf := &sdkresource.StateChangeConf{
		Pending: enum.Slice(awstypes.IndexStateDeleting),
		Target:  []string{},
		Refresh: statusIndex(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*resourceexplorer2.GetIndexOutput); ok {
		return output, err
	}

	return nil, err
}
