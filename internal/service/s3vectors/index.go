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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3vectors_index", name="Index")
// @ArnIdentity("index_arn")
// @Tags(identifierAttribute="index_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3vectors/types;awstypes;awstypes.Index")
// @Testing(hasNoPreExistingResource=true)
func newIndexResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &indexResource{}

	return r, nil
}

type indexResource struct {
	framework.ResourceWithModel[indexResourceModel]
	framework.WithImportByIdentity
}

func (r *indexResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dimension": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"distance_metric": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DistanceMetric](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEncryptionConfiguration: framework.ResourceOptionalComputedListOfObjectsAttribute[indexEncryptionConfigurationModel](ctx, 1, nil,
				listplanmodifier.UseStateForUnknown(),
				listplanmodifier.RequiresReplace(),
			),
			"index_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vector_bucket_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"metadata_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[indexMetadataConfigurationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"non_filterable_metadata_keys": schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 10),
							},
						},
					},
				},
			},
		},
	}
}

func (r *indexResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data indexResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	vectorBucketName, indexName := fwflex.StringValueFromFramework(ctx, data.VectorBucketName), fwflex.StringValueFromFramework(ctx, data.IndexName)
	var input s3vectors.CreateIndexInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateIndex(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Vectors Index (%s)", indexName), err.Error())

		return
	}

	output, err := findIndexByTwoPartKey(ctx, conn, vectorBucketName, indexName)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Index (%s)", indexName), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *indexResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data indexResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.IndexARN)
	output, err := findIndexByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Index (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *indexResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data indexResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.IndexARN)
	err := deleteIndex(ctx, conn, arn)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Vectors Index (%s)", arn), err.Error())

		return
	}
}

func deleteIndex(ctx context.Context, conn *s3vectors.Client, arn string) error {
	input := s3vectors.DeleteIndexInput{
		IndexArn: aws.String(arn),
	}
	_, err := conn.DeleteIndex(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		err = nil
	}

	return err
}

func findIndexByARN(ctx context.Context, conn *s3vectors.Client, arn string) (*awstypes.Index, error) {
	input := s3vectors.GetIndexInput{
		IndexArn: aws.String(arn),
	}

	return findIndex(ctx, conn, &input)
}

func findIndexByTwoPartKey(ctx context.Context, conn *s3vectors.Client, vectorBucketName, indexName string) (*awstypes.Index, error) {
	input := s3vectors.GetIndexInput{
		IndexName:        aws.String(indexName),
		VectorBucketName: aws.String(vectorBucketName),
	}

	return findIndex(ctx, conn, &input)
}

func findIndex(ctx context.Context, conn *s3vectors.Client, input *s3vectors.GetIndexInput) (*awstypes.Index, error) {
	output, err := conn.GetIndex(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Index == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Index, nil
}

type indexResourceModel struct {
	framework.WithRegionModel
	CreationTime            timetypes.RFC3339                                                  `tfsdk:"creation_time"`
	DataType                fwtypes.StringEnum[awstypes.DataType]                              `tfsdk:"data_type"`
	Dimension               types.Int32                                                        `tfsdk:"dimension"`
	DistanceMetric          fwtypes.StringEnum[awstypes.DistanceMetric]                        `tfsdk:"distance_metric"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[indexEncryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	IndexARN                types.String                                                       `tfsdk:"index_arn"`
	IndexName               types.String                                                       `tfsdk:"index_name"`
	MetadataConfiguration   fwtypes.ListNestedObjectValueOf[indexMetadataConfigurationModel]   `tfsdk:"metadata_configuration"`
	Tags                    tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                         `tfsdk:"tags_all"`
	VectorBucketName        types.String                                                       `tfsdk:"vector_bucket_name"`
}

type indexEncryptionConfigurationModel struct {
	KMSKeyARN fwtypes.ARN                          `tfsdk:"kms_key_arn"`
	SseType   fwtypes.StringEnum[awstypes.SseType] `tfsdk:"sse_type"`
}

type indexMetadataConfigurationModel struct {
	NonFilterableMetadataKeys fwtypes.SetOfString `tfsdk:"non_filterable_metadata_keys"`
}
