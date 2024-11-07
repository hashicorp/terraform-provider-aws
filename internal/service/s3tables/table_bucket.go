// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_bucket", name="Table Bucket")
func newResourceTableBucket(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTableBucket{}, nil
}

const (
	resNameTableBucket = "Table Bucket"
)

type resourceTableBucket struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceTableBucket) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_s3tables_table_bucket"
}

func (r *resourceTableBucket) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z-]*$`), "must contain only lowercase letters, numbers, or hyphens"),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z].*[0-9a-z]$`), "must start and end with a letter or number"),
					validators.PrefixNoneOf(
						"xn--",
						"sthree-",
						"sthree-configurator",
						"amzn-s3-demo-",
					),
					validators.SuffixNoneOf(
						"-s3alias",
						"--ol-s3",
						".mrap",
						"--x-s3",
					),
				},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceTableBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableBucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.CreateTableBucketInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateTableBucket(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	bucket, err := findTableBucket(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameTableBucket, plan.Name.String(), err),
			err.Error(),
		)
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, bucket, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTableBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTableBucket(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionSetting, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTableBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &s3tables.DeleteTableBucketInput{
		TableBucketARN: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTableBucket(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, resNameTableBucket, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTableBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func findTableBucket(ctx context.Context, conn *s3tables.Client, arn string) (*s3tables.GetTableBucketOutput, error) {
	in := s3tables.GetTableBucketInput{
		TableBucketARN: aws.String(arn),
	}

	out, err := conn.GetTableBucket(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceTableBucketModel struct {
	ARN            types.String      `tfsdk:"arn"`
	CreatedAt      timetypes.RFC3339 `tfsdk:"created_at"`
	Name           types.String      `tfsdk:"name"`
	OwnerAccountID types.String      `tfsdk:"owner_account_id"`
}
