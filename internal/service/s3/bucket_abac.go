// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3_bucket_abac", name="Bucket ABAC")
func newResourceBucketABAC(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceBucketABAC{}, nil
}

const (
	ResNameBucketABAC = "Bucket ABAC"
)

type resourceBucketABAC struct {
	framework.ResourceWithModel[resourceBucketABACModel]
}

func (r *resourceBucketABAC) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			names.AttrExpectedBucketOwner: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				DeprecationMessage: "This attribute will be removed in a future verion of the provider.",
			},
		},
		Blocks: map[string]schema.Block{
			"abac_status": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[abacStatusModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrStatus: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceBucketABAC) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceBucketABACModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := plan.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	var input s3.PutBucketAbacInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutBucketAbac(ctx, &input)
	}, errCodeNoSuchBucket)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, bucket)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceBucketABAC) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceBucketABACModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := state.Bucket.ValueString()
	expectedBucketOwner := state.ExpectedBucketOwner.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	out, err := findBucketABAC(ctx, conn, bucket, expectedBucketOwner)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, bucket)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state.ABACStatus))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceBucketABAC) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceBucketABACModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := state.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input s3.PutBucketAbacInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
			return conn.PutBucketAbac(ctx, &input)
		}, errCodeNoSuchBucket)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, bucket)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceBucketABAC) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceBucketABACModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)
	bucket := state.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = r.Meta().S3ExpressClient(ctx)
	}

	var input s3.PutBucketAbacInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.AbacStatus = &awstypes.AbacStatus{
		Status: awstypes.BucketAbacStatusDisabled,
	}

	_, err := conn.PutBucketAbac(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, bucket)
		return
	}
}

func (r *resourceBucketABAC) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	bucket, expectedBucketOwner, err := parseResourceID(request.ID)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrBucket), bucket)...)
	if expectedBucketOwner != "" {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrExpectedBucketOwner), expectedBucketOwner)...)
	}
}

func findBucketABAC(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*awstypes.AbacStatus, error) {
	input := s3.GetBucketAbacInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	out, err := conn.GetBucketAbac(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AbacStatus == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AbacStatus, nil
}

type resourceBucketABACModel struct {
	framework.WithRegionModel
	ABACStatus          fwtypes.ListNestedObjectValueOf[abacStatusModel] `tfsdk:"abac_status"`
	Bucket              types.String                                     `tfsdk:"bucket"`
	ExpectedBucketOwner types.String                                     `tfsdk:"expected_bucket_owner"`
}

type abacStatusModel struct {
	Status types.String `tfsdk:"status"`
}
