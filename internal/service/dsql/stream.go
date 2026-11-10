// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dsql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dsql_stream", name="Stream")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("cluster_identifier")
// @IdentityAttribute("stream_identifier")
// @ImportIDHandler("streamImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dsql;dsql;dsql.GetStreamOutput")
// @Testing(importStateIdAttributes="cluster_identifier;stream_identifier", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(preCheck="testAccPreCheck")
func newStreamResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &streamResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type streamResource struct {
	framework.ResourceWithModel[streamResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *streamResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrClusterIdentifier: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9]{26}$`), ""),
				},
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.StreamFormat](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ordering": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.StreamOrdering](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.StreamStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stream_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"target_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetDefinitionModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"kinesis": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisTargetDefinitionModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthBetween(20, 2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws(-[^:]+)?:iam::[0-9]{12}:role(/[a-zA-Z0-9+=,.@_-]+)+$`), ""),
										},
									},
									names.AttrStreamARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws.*:kinesis:.*:\d{12}:stream/\S+$`), ""),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *streamResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan streamResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	var input dsql.CreateStreamInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	// IAM is eventually consistent: the call is retried while the service cannot use a role created just before it.
	out, err := tfresource.RetryWhenIsAErrorMessageContains[*dsql.CreateStreamOutput, *awstypes.ValidationException](ctx, iamPropagationTimeout, func(ctx context.Context) (*dsql.CreateStreamOutput, error) {
		return conn.CreateStream(ctx, &input)
	}, "Aurora DSQL can't access the specified IAM role")
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	// Every identity attribute is in state before the next call, so a failure from here on taints the resource.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrClusterIdentifier), plan.ClusterIdentifier))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("stream_identifier"), plan.StreamIdentifier))
	if response.Diagnostics.HasError() {
		return
	}

	clusterIdentifier := fwflex.StringValueFromFramework(ctx, plan.ClusterIdentifier)
	streamIdentifier := fwflex.StringValueFromFramework(ctx, plan.StreamIdentifier)

	output, err := waitStreamCreated(ctx, conn, clusterIdentifier, streamIdentifier, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrClusterIdentifier, clusterIdentifier, "stream_identifier", streamIdentifier)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *streamResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data streamResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	clusterIdentifier := fwflex.StringValueFromFramework(ctx, data.ClusterIdentifier)
	streamIdentifier := fwflex.StringValueFromFramework(ctx, data.StreamIdentifier)
	output, err := findStreamByTwoPartKey(ctx, conn, clusterIdentifier, streamIdentifier)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrClusterIdentifier, clusterIdentifier, "stream_identifier", streamIdentifier)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

// flatten sets the model from the finder's value, for Read and for an item the list resource includes.
func (r *streamResource) flatten(ctx context.Context, output *dsql.GetStreamOutput, data *streamResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, output, data)...)
	if diags.HasError() {
		return diags
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func (r *streamResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data streamResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	clusterIdentifier := fwflex.StringValueFromFramework(ctx, data.ClusterIdentifier)
	streamIdentifier := fwflex.StringValueFromFramework(ctx, data.StreamIdentifier)
	input := dsql.DeleteStreamInput{
		ClusterIdentifier: aws.String(clusterIdentifier),
		StreamIdentifier:  aws.String(streamIdentifier),
		ClientToken:       aws.String(create.UniqueId(ctx)),
	}
	_, err := conn.DeleteStream(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrClusterIdentifier, clusterIdentifier, "stream_identifier", streamIdentifier)
		return
	}

	if _, err := waitStreamDeleted(ctx, conn, clusterIdentifier, streamIdentifier, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrClusterIdentifier, clusterIdentifier, "stream_identifier", streamIdentifier)
	}
}

const iamPropagationTimeout = 2 * time.Minute

func findStreamByTwoPartKey(ctx context.Context, conn *dsql.Client, clusterIdentifier, streamIdentifier string) (*dsql.GetStreamOutput, error) {
	input := dsql.GetStreamInput{
		ClusterIdentifier: aws.String(clusterIdentifier),
		StreamIdentifier:  aws.String(streamIdentifier),
	}

	output, err := conn.GetStream(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func statusStream(conn *dsql.Client, clusterIdentifier, streamIdentifier string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findStreamByTwoPartKey(ctx, conn, clusterIdentifier, streamIdentifier)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *dsql.Client, clusterIdentifier, streamIdentifier string, timeout time.Duration) (*dsql.GetStreamOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StreamStatusCreating, awstypes.StreamStatusImpaired),
		Target:  enum.Slice(awstypes.StreamStatusActive),
		Refresh: statusStream(conn, clusterIdentifier, streamIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*dsql.GetStreamOutput); ok {
		if v := output.StatusReason; v != nil {
			retry.SetLastError(err, errors.New(string(v.Error)))
		}
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitStreamDeleted(ctx context.Context, conn *dsql.Client, clusterIdentifier, streamIdentifier string, timeout time.Duration) (*dsql.GetStreamOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StreamStatusActive, awstypes.StreamStatusCreating, awstypes.StreamStatusDeleted, awstypes.StreamStatusDeleting, awstypes.StreamStatusFailed, awstypes.StreamStatusImpaired),
		Target:  []string{},
		Refresh: statusStream(conn, clusterIdentifier, streamIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*dsql.GetStreamOutput); ok {
		if v := output.StatusReason; v != nil {
			retry.SetLastError(err, errors.New(string(v.Error)))
		}
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

var (
	_ inttypes.ImportIDParser = streamImportID{}
)

type streamImportID struct{}

func (streamImportID) Parse(id string) (string, map[string]any, error) {
	const (
		streamIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, streamIDParts, false)
	if err != nil {
		return "", nil, smarterr.NewError(err)
	}

	return id, map[string]any{
		names.AttrClusterIdentifier: parts[0],
		"stream_identifier":         parts[1],
	}, nil
}

type streamResourceModel struct {
	framework.WithRegionModel
	ARN               fwtypes.ARN                                            `tfsdk:"arn"`
	ClusterIdentifier types.String                                           `tfsdk:"cluster_identifier"`
	CreationTime      timetypes.RFC3339                                      `tfsdk:"creation_time"`
	Format            fwtypes.StringEnum[awstypes.StreamFormat]              `tfsdk:"format"`
	Ordering          fwtypes.StringEnum[awstypes.StreamOrdering]            `tfsdk:"ordering"`
	Status            fwtypes.StringEnum[awstypes.StreamStatus]              `tfsdk:"status"`
	StreamIdentifier  types.String                                           `tfsdk:"stream_identifier"`
	Tags              tftags.Map                                             `tfsdk:"tags"`
	TagsAll           tftags.Map                                             `tfsdk:"tags_all"`
	TargetDefinition  fwtypes.ListNestedObjectValueOf[targetDefinitionModel] `tfsdk:"target_definition"`
	Timeouts          timeouts.Value                                         `tfsdk:"timeouts" autoflex:"-"`
}

type kinesisTargetDefinitionModel struct {
	RoleARN   fwtypes.ARN `tfsdk:"role_arn"`
	StreamARN fwtypes.ARN `tfsdk:"stream_arn"`
}

type targetDefinitionModel struct {
	Kinesis fwtypes.ListNestedObjectValueOf[kinesisTargetDefinitionModel] `tfsdk:"kinesis"`
}

var (
	_ fwflex.Expander  = targetDefinitionModel{}
	_ fwflex.Flattener = &targetDefinitionModel{}
)

func (m *targetDefinitionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.TargetDefinitionMemberKinesis:
		var model kinesisTargetDefinitionModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Kinesis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, t.Tag, &diags)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("TargetDefinition flatten: %T", v))
	}

	return diags
}

func (m targetDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.Kinesis.IsNull():
		data, d := m.Kinesis.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.TargetDefinitionMemberKinesis
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	diags.AddError("Invalid Union", "TargetDefinition expand: no member is set")

	return nil, diags
}
