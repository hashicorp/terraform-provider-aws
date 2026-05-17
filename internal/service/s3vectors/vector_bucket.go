// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3vectors

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3vectors_vector_bucket", name="Vector Bucket")
// @ArnIdentity("vector_bucket_arn")
// @Tags(identifierAttribute="vector_bucket_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3vectors/types;awstypes;awstypes.VectorBucket")
// @Testing(importIgnore="force_destroy")
// @Testing(hasNoPreExistingResource=true)
func newVectorBucketResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &vectorBucketResource{}

	return r, nil
}

type vectorBucketResource struct {
	framework.ResourceWithModel[vectorBucketResourceModel]
	framework.WithImportByIdentity
}

func (r *vectorBucketResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEncryptionConfiguration: framework.ResourceOptionalComputedListOfObjectsAttribute[encryptionConfigurationModel](ctx, 1, nil, listplanmodifier.RequiresReplaceIfConfigured(), listplanmodifier.UseStateForUnknown()),
			names.AttrForceDestroy: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vector_bucket_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vector_bucket_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersHyphens,
					tfstringvalidator.StartsWithLetterOrNumber,
					tfstringvalidator.EndsWithLetterOrNumber,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *vectorBucketResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.VectorBucketName)
	var input s3vectors.CreateVectorBucketInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateVectorBucket(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Vectors Vector Bucket (%s)", name), err.Error())

		return
	}

	output, err := findVectorBucketByName(ctx, conn, name)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Vector Bucket (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vectorBucketResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	output, err := findVectorBucketByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Vector Bucket (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vectorBucketResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	if data.ForceDestroy.ValueBool() {
		input := s3vectors.ListIndexesInput{
			VectorBucketArn: aws.String(arn),
		}

		pages := s3vectors.NewListIndexesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("listing S3 Vectors Vector Bucket (%s) indexes", arn), err.Error())

				return
			}

			for _, v := range page.Indexes {
				arn := aws.ToString(v.IndexArn)
				err := deleteIndex(ctx, conn, arn)

				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Vectors index (%s)", arn), err.Error())

					return
				}
			}
		}
	}

	input := s3vectors.DeleteVectorBucketInput{
		VectorBucketArn: aws.String(arn),
	}
	_, err := conn.DeleteVectorBucket(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Vectors Vector Bucket (%s)", arn), err.Error())

		return
	}
}

func findVectorBucketByARN(ctx context.Context, conn *s3vectors.Client, arn string) (*awstypes.VectorBucket, error) {
	input := s3vectors.GetVectorBucketInput{
		VectorBucketArn: aws.String(arn),
	}

	return findVectorBucket(ctx, conn, &input)
}

func findVectorBucketByName(ctx context.Context, conn *s3vectors.Client, name string) (*awstypes.VectorBucket, error) {
	input := s3vectors.GetVectorBucketInput{
		VectorBucketName: aws.String(name),
	}

	return findVectorBucket(ctx, conn, &input)
}

func findVectorBucket(ctx context.Context, conn *s3vectors.Client, input *s3vectors.GetVectorBucketInput) (*awstypes.VectorBucket, error) {
	output, err := conn.GetVectorBucket(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VectorBucket == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.VectorBucket, nil
}

type vectorBucketResourceModel struct {
	framework.WithRegionModel
	CreationTime            timetypes.RFC3339                                             `tfsdk:"creation_time"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	ForceDestroy            types.Bool                                                    `tfsdk:"force_destroy"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	VectorBucketARN         types.String                                                  `tfsdk:"vector_bucket_arn"`
	VectorBucketName        types.String                                                  `tfsdk:"vector_bucket_name"`
}

type encryptionConfigurationModel struct {
	KMSKeyARN fwtypes.ARN                          `tfsdk:"kms_key_arn"`
	SSEType   fwtypes.StringEnum[awstypes.SseType] `tfsdk:"sse_type"`
}
