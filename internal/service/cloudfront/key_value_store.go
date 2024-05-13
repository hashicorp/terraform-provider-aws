// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Key Value Store")
func newKeyValueStoreResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &keyValueStoreResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type keyValueStoreResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *keyValueStoreResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_key_value_store"
}

func (r *keyValueStoreResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrComment: schema.StringAttribute{
				Optional: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-_]{1,64}$`),
						"must contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *keyValueStoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data keyValueStoreResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.CreateKeyValueStoreInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	name := aws.ToString(input.Name)
	_, err := conn.CreateKeyValueStore(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Key Value Store (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	outputDKVS, err := waitKeyValueStoreCreated(ctx, conn, name, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront Key Value Store (%s) create", name), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputDKVS.KeyValueStore, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, outputDKVS.ETag)
	data.setID() // API response has a field named 'Id' which isn't the resource's ID.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyValueStoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data keyValueStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findKeyValueStoreByName(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Key Value Store (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.KeyValueStore, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.setID() // API response has a field named 'Id' which isn't the resource's ID.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyValueStoreResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new keyValueStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	kvsARN := old.ARN.ValueString()

	// Updating changes the etag of the key value store.
	// Use a mutex serialize actions
	mutexKey := kvsARN
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	input := &cloudfront.UpdateKeyValueStoreInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.IfMatch = fwflex.StringFromFramework(ctx, old.ETag)

	output, err := conn.UpdateKeyValueStore(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating CloudFront Key Value Store (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.KeyValueStore, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	new.ETag = fwflex.StringToFramework(ctx, output.ETag)
	new.setID() // API response has a field named 'Id' which isn't the resource's ID.

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *keyValueStoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data keyValueStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	kvsARN := data.ARN.ValueString()

	// Use a mutex serialize actions
	mutexKey := kvsARN
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	input := &cloudfront.DeleteKeyValueStoreInput{
		IfMatch: fwflex.StringFromFramework(ctx, data.ETag),
		Name:    fwflex.StringFromFramework(ctx, data.ID),
	}

	_, err := conn.DeleteKeyValueStore(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Key Value Store (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findKeyValueStoreByName(ctx context.Context, conn *cloudfront.Client, name string) (*cloudfront.DescribeKeyValueStoreOutput, error) {
	input := &cloudfront.DescribeKeyValueStoreInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeKeyValueStore(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyValueStore == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusKeyValueStore(ctx context.Context, conn *cloudfront.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findKeyValueStoreByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.KeyValueStore.Status), nil
	}
}

func waitKeyValueStoreCreated(ctx context.Context, conn *cloudfront.Client, name string, timeout time.Duration) (*cloudfront.DescribeKeyValueStoreOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{keyValueStoreStatusProvisioning},
		Target:  []string{keyValueStoreStatusReady},
		Refresh: statusKeyValueStore(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.DescribeKeyValueStoreOutput); ok {
		return output, err
	}

	return nil, err
}

type keyValueStoreResourceModel struct {
	ARN              types.String      `tfsdk:"arn"`
	Comment          types.String      `tfsdk:"comment"`
	ETag             types.String      `tfsdk:"etag"`
	ID               types.String      `tfsdk:"id"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`
	Name             types.String      `tfsdk:"name"`
	Timeouts         timeouts.Value    `tfsdk:"timeouts"`
}

func (data *keyValueStoreResourceModel) InitFromID() error {
	data.Name = data.ID

	return nil
}

func (data *keyValueStoreResourceModel) setID() {
	data.ID = data.Name
}
