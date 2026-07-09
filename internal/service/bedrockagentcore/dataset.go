// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_dataset", name="Dataset")
// @Tags(identifierAttribute="dataset_arn")
// @IdentityAttribute("dataset_id", identityDuplicateAttributes="dataset_id")
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="dataset_id")
// @Testing(importIgnore="source", plannableImportAction="Replace")
func newDatasetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &datasetResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type datasetResource struct {
	framework.ResourceWithModel[datasetResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *datasetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_arn": framework.ARNAttributeComputedOnly(),
			"dataset_id":  framework.IDAttribute(),
			"dataset_version": schema.StringAttribute{
				Computed: true,
			},
			// description is Optional+Computed: the UpdateDataset API treats a nil
			// description as "leave unchanged" (never clears), so removing it from
			// config retains the prior value rather than producing an
			// inconsistent-result-after-apply. Empirically verified against the API.
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"draft_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DraftStatus](),
				Computed:   true,
			},
			"example_count": schema.Int64Attribute{
				Computed: true,
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
			},
			"schema_type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.DatasetSchemaType](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DatasetStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			// updated_at advances on every update; NO UseStateForUnknown (it must stay
			// unknown-in-plan and be re-read after apply).
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			// source is the create-time seed of examples and is NOT returned by
			// GetDataset (write-only); it is retained from configuration and any
			// change forces replacement.
			"source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[datasetSourceModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"inline_examples": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[datasetInlineExamplesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("inline_examples"),
									path.MatchRelative().AtParent().AtName("s3_source"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"examples": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Required:    true,
										Sensitive:   true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
						"s3_source": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[datasetS3SourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											// The SDK pattern bounds the object key with {1,1024},
											// which exceeds Go regexp's 1000 repeat limit; the key
											// length is enforced by the service. Bucket + scheme are
											// still validated here.
											stringvalidator.RegexMatches(regexache.MustCompile(`^s3://[a-z0-9][a-z0-9.\-]{1,61}[a-z0-9]/.+$`), ""),
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *datasetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data datasetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateDatasetInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Source = data.expandSource(ctx, &response.Diagnostics)
	input.Tags = getTagsIn(ctx)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateDataset(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DatasetName.ValueString())
		return
	}

	data.DatasetID = fwflex.StringToFramework(ctx, out.DatasetId)

	created, err := waitDatasetCreated(ctx, conn, data.DatasetID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DatasetID.ValueString())
		return
	}

	// source is write-only (absent from Get); flattenReadBack leaves it untouched.
	smerr.AddEnrich(ctx, &response.Diagnostics, data.flattenReadBack(ctx, created))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *datasetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data datasetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	out, err := findDatasetByID(ctx, conn, data.DatasetID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DatasetID.ValueString())
		return
	}

	// source is write-only (absent from Get); flattenReadBack leaves the prior
	// state value untouched.
	smerr.AddEnrich(ctx, &response.Diagnostics, data.flattenReadBack(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *datasetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state datasetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	// description is the only mutable input (UpdateDataset accepts nothing else);
	// everything else is RequiresReplace.
	if !plan.Description.Equal(state.Description) {
		input := bedrockagentcorecontrol.UpdateDatasetInput{
			DatasetId:   fwflex.StringFromFramework(ctx, plan.DatasetID),
			ClientToken: aws.String(create.UniqueId(ctx)),
			Description: fwflex.StringFromFramework(ctx, plan.Description),
		}
		if _, err := conn.UpdateDataset(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.DatasetID.ValueString())
			return
		}

		if _, err := waitDatasetUpdated(ctx, conn, plan.DatasetID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.DatasetID.ValueString())
			return
		}
	}

	// Always re-read so computed values that advance on update (updated_at,
	// draft_status, example_count) are known after apply — including tags-only
	// updates where no dataset field changed.
	out, err := findDatasetByID(ctx, conn, plan.DatasetID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.DatasetID.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flattenReadBack(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *datasetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data datasetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.DeleteDatasetInput{
		DatasetId: fwflex.StringFromFramework(ctx, data.DatasetID),
	}
	if _, err := conn.DeleteDataset(ctx, &input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DatasetID.ValueString())
		return
	}

	if _, err := waitDatasetDeleted(ctx, conn, data.DatasetID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DatasetID.ValueString())
		return
	}
}

// expandSource builds the DataSourceType union from the source block. The union
// can't be built by AutoFlex, and inline_examples.examples is a []document.Interface
// assembled from JSON strings.
func (m datasetResourceModel) expandSource(ctx context.Context, diags *diag.Diagnostics) awstypes.DataSourceType {
	if m.Source.IsNull() || m.Source.IsUnknown() {
		return nil
	}
	src, d := m.Source.ToPtr(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() {
		return nil
	}

	switch {
	case !src.S3Source.IsNull():
		s3, d := src.S3Source.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() {
			return nil
		}
		return &awstypes.DataSourceTypeMemberS3Source{
			Value: awstypes.S3Source{
				S3Uri: fwflex.StringFromFramework(ctx, s3.S3URI),
			},
		}

	case !src.InlineExamples.IsNull():
		ie, d := src.InlineExamples.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() {
			return nil
		}
		var exStrs []string
		smerr.AddEnrich(ctx, diags, ie.Examples.ElementsAs(ctx, &exStrs, false))
		if diags.HasError() {
			return nil
		}
		docs := make([]document.Interface, 0, len(exStrs))
		for _, s := range exStrs {
			doc, err := tfsmithy.DocumentFromJSONString(s, document.NewLazyDocument)
			if err != nil {
				diags.AddError("creating Smithy document", err.Error())
				return nil
			}
			docs = append(docs, doc)
		}
		return &awstypes.DataSourceTypeMemberInlineExamples{
			Value: awstypes.InlineExamplesSource{
				Examples: docs,
			},
		}
	}

	return nil
}

// flattenReadBack copies the round-tripping fields from a Get output into the
// model WITHOUT touching the write-only source block (absent from Get).
func (m *datasetResourceModel) flattenReadBack(ctx context.Context, out *bedrockagentcorecontrol.GetDatasetOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	m.DatasetARN = fwflex.StringToFramework(ctx, out.DatasetArn)
	m.DatasetID = fwflex.StringToFramework(ctx, out.DatasetId)
	m.DatasetName = fwflex.StringToFramework(ctx, out.DatasetName)
	m.DatasetVersion = fwflex.StringToFramework(ctx, out.DatasetVersion)
	m.Description = fwflex.StringToFramework(ctx, out.Description)
	m.ExampleCount = fwflex.Int64ToFramework(ctx, out.ExampleCount)
	m.FailureReason = fwflex.StringToFramework(ctx, out.FailureReason)
	m.SchemaType = fwtypes.StringEnumValue(out.SchemaType)
	m.Status = fwtypes.StringEnumValue(out.Status)
	m.CreatedAt = timetypes.NewRFC3339TimePointerValue(out.CreatedAt)
	m.UpdatedAt = timetypes.NewRFC3339TimePointerValue(out.UpdatedAt)

	if out.DraftStatus != "" {
		m.DraftStatus = fwtypes.StringEnumValue(out.DraftStatus)
	} else {
		m.DraftStatus = fwtypes.StringEnumNull[awstypes.DraftStatus]()
	}

	if out.KmsKeyArn != nil {
		m.KMSKeyARN = fwtypes.ARNValue(aws.ToString(out.KmsKeyArn))
	} else {
		m.KMSKeyARN = fwtypes.ARNNull()
	}

	return diags
}

func findDatasetByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetDatasetOutput, error) {
	input := bedrockagentcorecontrol.GetDatasetInput{
		DatasetId: aws.String(id),
	}

	out, err := conn.GetDataset(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

func statusDataset(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findDatasetByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}
		return out, string(out.Status), nil
	}
}

func waitDatasetCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetDatasetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DatasetStatusCreating),
		Target:                    enum.Slice(awstypes.DatasetStatusActive),
		Refresh:                   statusDataset(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetDatasetOutput); ok {
		if out.FailureReason != nil {
			retry.SetLastError(err, fmt.Errorf("%s", aws.ToString(out.FailureReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitDatasetUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetDatasetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DatasetStatusUpdating),
		Target:                    enum.Slice(awstypes.DatasetStatusActive),
		Refresh:                   statusDataset(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetDatasetOutput); ok {
		if out.FailureReason != nil {
			retry.SetLastError(err, fmt.Errorf("%s", aws.ToString(out.FailureReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitDatasetDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetDatasetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DatasetStatusDeleting, awstypes.DatasetStatusActive),
		Target:  []string{},
		Refresh: statusDataset(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetDatasetOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type datasetResourceModel struct {
	framework.WithRegionModel
	CreatedAt      timetypes.RFC3339                                   `tfsdk:"created_at"`
	DatasetARN     types.String                                        `tfsdk:"dataset_arn"`
	DatasetID      types.String                                        `tfsdk:"dataset_id"`
	DatasetName    types.String                                        `tfsdk:"name"`
	DatasetVersion types.String                                        `tfsdk:"dataset_version"`
	Description    types.String                                        `tfsdk:"description"`
	DraftStatus    fwtypes.StringEnum[awstypes.DraftStatus]            `tfsdk:"draft_status"`
	ExampleCount   types.Int64                                         `tfsdk:"example_count"`
	FailureReason  types.String                                        `tfsdk:"failure_reason"`
	KMSKeyARN      fwtypes.ARN                                         `tfsdk:"kms_key_arn"`
	SchemaType     fwtypes.StringEnum[awstypes.DatasetSchemaType]      `tfsdk:"schema_type"`
	Source         fwtypes.ListNestedObjectValueOf[datasetSourceModel] `tfsdk:"source" autoflex:"-"`
	Status         fwtypes.StringEnum[awstypes.DatasetStatus]          `tfsdk:"status"`
	Tags           tftags.Map                                          `tfsdk:"tags"`
	TagsAll        tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts       timeouts.Value                                      `tfsdk:"timeouts"`
	UpdatedAt      timetypes.RFC3339                                   `tfsdk:"updated_at"`
}

type datasetSourceModel struct {
	InlineExamples fwtypes.ListNestedObjectValueOf[datasetInlineExamplesModel] `tfsdk:"inline_examples"`
	S3Source       fwtypes.ListNestedObjectValueOf[datasetS3SourceModel]       `tfsdk:"s3_source"`
}

type datasetInlineExamplesModel struct {
	Examples fwtypes.ListOfString `tfsdk:"examples"`
}

type datasetS3SourceModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}
