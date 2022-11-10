package simpledb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	registerFrameworkResourceFactory(newResourceDomain)
}

// newResourceDomain instantiates a new Resource for the aws_simpledb_domain resource.
func newResourceDomain(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceDomain{}, nil
}

type resourceDomain struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (r *resourceDomain) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_simpledb_domain"
}

// GetSchema returns the schema for this resource.
func (r *resourceDomain) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
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

	name := data.Name.Value
	input := &simpledb.CreateDomainInput{
		DomainName: aws.String(name),
	}

	_, err := r.meta.SimpleDBConn.CreateDomainWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating SimpleDB Domain", err.Error())

		return
	}

	data.ID = types.String{Value: name}

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

	_, err := FindDomainByName(ctx, r.meta.SimpleDBConn, data.ID.Value)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(errs.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading SimpleDB Domain", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceDomain) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Noop.
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

	tflog.Debug(ctx, "deleting SimpleDB Domain", map[string]interface{}{
		"id": data.ID.Value,
	})

	_, err := r.meta.SimpleDBConn.DeleteDomainWithContext(ctx, &simpledb.DeleteDomainInput{
		DomainName: aws.String(data.ID.Value),
	})

	if err != nil {
		response.Diagnostics.AddError("deleting SimpleDB Domain", err.Error())

		return
	}
}

// ImportState is called when the provider must import the state of a resource instance.
// This method must return enough state so the Read method can properly refresh the full resource.
//
// If setting an attribute with the import identifier, it is recommended to use the ImportStatePassthroughID() call in this method.
func (r *resourceDomain) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("name"), request.ID)...)
}

type resourceDomainData struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func FindDomainByName(ctx context.Context, conn *simpledb.SimpleDB, name string) (*simpledb.DomainMetadataOutput, error) {
	input := &simpledb.DomainMetadataInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DomainMetadataWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, simpledb.ErrCodeNoSuchDomain) {
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

	return output, nil
}
