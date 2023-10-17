// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	// e.g. example--usw2-az2--x-s3
	directoryBucketNameRegex = regexache.MustCompile(`^([0-9a-z.-]+)--([a-z]+\d+-az\d+)--x-s3$`)
)

// @FrameworkResource(name="Directory Bucket")
func newResourceDirectoryBucket(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDirectoryBucket{}

	return r, nil
}

type resourceDirectoryBucket struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceDirectoryBucket) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3_directory_bucket"
}

func (r *resourceDirectoryBucket) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"bucket": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(directoryBucketNameRegex, `must be in the format [bucket_name]--[azid]--x-s3. Use the aws_s3_bucket resource to manage general purpose buckets`),
				},
			},
			"data_redundancy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  enum.FrameworkDefault(awstypes.DataRedundancySingleAvailabilityZone),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.DataRedundancy](),
				},
			},
			"force_destroy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrID: framework.IDAttribute(),
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  enum.FrameworkDefault(awstypes.BucketTypeDirectory),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.BucketType](),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"location": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceDirectoryBucketLocationData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  enum.FrameworkDefault(awstypes.LocationTypeAvailabilityZone),
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.LocationType](),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
			},
		},
	}
}

func (r *resourceDirectoryBucket) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	locationData, diags := data.Location.ToPtr(ctx)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	input := &s3.CreateBucketInput{
		Bucket: flex.StringFromFramework(ctx, data.Bucket),
		CreateBucketConfiguration: &awstypes.CreateBucketConfiguration{
			Bucket: &awstypes.BucketInfo{
				DataRedundancy: awstypes.DataRedundancy(data.DataRedundancy.ValueString()),
				Type:           awstypes.BucketType(data.Type.ValueString()),
			},
			Location: &awstypes.LocationInfo{
				Name: flex.StringFromFramework(ctx, locationData.Name),
				Type: awstypes.LocationType(locationData.Type.ValueString()),
			},
		},
	}

	_, err := conn.CreateBucket(ctx, input, useRegionalEndpointInUSEast1)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Directory Bucket (%s)", data.Bucket.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = types.StringValue(r.arn(data.Bucket.ValueString()))
	data.ID = data.Bucket

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceDirectoryBucket) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	err := findBucket(ctx, conn, data.ID.ValueString(), useRegionalEndpointInUSEast1)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Directory Bucket (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.ARN = types.StringValue(r.arn(data.ID.ValueString()))
	data.Bucket = data.ID
	// No API to return bucket type, location etc.
	data.DataRedundancy = flex.StringValueToFramework(ctx, awstypes.DataRedundancySingleAvailabilityZone)
	if matches := directoryBucketNameRegex.FindStringSubmatch(data.ID.ValueString()); len(matches) == 3 {
		data.Location = fwtypes.NewListNestedObjectValueOfPtr(ctx, &resourceDirectoryBucketLocationData{
			Name: flex.StringValueToFramework(ctx, matches[2]),
			Type: flex.StringValueToFramework(ctx, awstypes.LocationTypeAvailabilityZone),
		})
	}
	data.Type = flex.StringValueToFramework(ctx, awstypes.BucketTypeDirectory)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceDirectoryBucket) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceDirectoryBucket) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	_, err := conn.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: flex.StringFromFramework(ctx, data.ID),
	}, useRegionalEndpointInUSEast1)

	if tfawserr.ErrCodeEquals(err, errCodeBucketNotEmpty) {
		if data.ForceDestroy.ValueBool() {
			// Empty the bucket and try again.
			_, err = emptyBucket(ctx, conn, data.ID.ValueString(), false)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("emptying S3 Directory Bucket (%s)", data.ID.ValueString()), err.Error())

				return
			}

			_, err = conn.DeleteBucket(ctx, &s3.DeleteBucketInput{
				Bucket: flex.StringFromFramework(ctx, data.ID),
			})
		}
	}

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Directory Bucket (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

// arn returns the ARN of the specified bucket.
func (r *resourceDirectoryBucket) arn(bucket string) string {
	return r.RegionalARN("s3express", fmt.Sprintf("bucket/%s", bucket))
}

type resourceDirectoryBucketData struct {
	ARN            types.String                                                         `tfsdk:"arn"`
	Bucket         types.String                                                         `tfsdk:"bucket"`
	DataRedundancy types.String                                                         `tfsdk:"data_redundancy"`
	ForceDestroy   types.Bool                                                           `tfsdk:"force_destroy"`
	Location       fwtypes.ListNestedObjectValueOf[resourceDirectoryBucketLocationData] `tfsdk:"location"`
	ID             types.String                                                         `tfsdk:"id"`
	Type           types.String                                                         `tfsdk:"type"`
}

type resourceDirectoryBucketLocationData struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}
