// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_memory", name="Memory")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types;awstypes.Memory")
// @Testing(generator="randomMemoryName(t)")
func newMemoryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &memoryResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type memoryResource struct {
	framework.ResourceWithModel[memoryResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *memoryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"event_expiry_duration": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.Between(7, 365),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"memory_execution_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"indexed_key": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[indexedKeyModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIf(
						requiresReplaceIfIndexedKeyRemoved,
						"Removing or changing an existing indexed key requires replacement; new keys are added in place.",
						"Removing or changing an existing indexed key requires replacement; new keys are added in place.",
					),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\s._:/=+@-]*$`), ""),
							},
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MetadataValueType](),
							Required:   true,
						},
					},
				},
			},
			"stream_delivery_resources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[streamDeliveryResourcesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[streamDeliveryResourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"kinesis": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisResourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"data_stream_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"content_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[contentConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeBetween(1, 1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ContentType](),
																Required:   true,
															},
															"level": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ContentLevel](),
																Optional:   true,
																Computed:   true,
															},
														},
													},
												},
											},
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

func (r *memoryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data memoryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateMemoryInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateMemoryOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateMemory(ctx, &input)

		// IAM propagation - retry if role validation fails
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Role validation failed") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "valid trust policy") {
			return tfresource.RetryableError(err)
		}
		// IAM propagation - retry while the execution role's permissions to a
		// referenced resource (e.g. a Kinesis stream in stream_delivery_resources)
		// have not yet propagated.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "does not have access to provided") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	memoryID := aws.ToString(out.Memory.Id)

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Memory, &data, fwflex.WithFieldNamePrefix("Memory")))
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitMemoryCreated(ctx, conn, memoryID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root(names.AttrID), memoryID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *memoryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data memoryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	memoryID := fwflex.StringValueFromFramework(ctx, data.ID)
	out, err := findMemoryByID(ctx, conn, memoryID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("Memory")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *memoryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old memoryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		memoryID := fwflex.StringValueFromFramework(ctx, new.ID)
		var input bedrockagentcorecontrol.UpdateMemoryInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))
		input.MemoryId = aws.String(memoryID)

		// AddIndexedKeys is additive and does not map from the indexed_key model field, so send only
		// the keys that are not already present. Removals force replacement (see the schema plan modifier).
		newKeys, d := new.IndexedKeys.ToSlice(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		oldKeys, d := old.IndexedKeys.ToSlice(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		existing := make(map[string]struct{}, len(oldKeys))
		for _, k := range oldKeys {
			existing[indexedKeyIdentity(k)] = struct{}{}
		}
		for _, k := range newKeys {
			if _, ok := existing[indexedKeyIdentity(k)]; !ok {
				input.AddIndexedKeys = append(input.AddIndexedKeys, awstypes.IndexedKey{
					Key:  aws.String(k.Key.ValueString()),
					Type: k.Type.ValueEnum(),
				})
			}
		}

		_, err := conn.UpdateMemory(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *memoryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data memoryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	memoryID := fwflex.StringValueFromFramework(ctx, data.ID)
	input := bedrockagentcorecontrol.DeleteMemoryInput{
		MemoryId: aws.String(memoryID),
	}
	_, err := conn.DeleteMemory(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryID)
		return
	}

	if _, err := waitMemoryDeleted(ctx, conn, memoryID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryID)
		return
	}
}

func waitMemoryCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*awstypes.Memory, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MemoryStatusCreating),
		Target:                    enum.Slice(awstypes.MemoryStatusActive),
		Refresh:                   statusMemory(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Memory); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMemoryDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*awstypes.Memory, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MemoryStatusDeleting, awstypes.MemoryStatusActive),
		Target:  []string{},
		Refresh: statusMemory(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Memory); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMemory(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findMemoryByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findMemoryByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*awstypes.Memory, error) {
	input := bedrockagentcorecontrol.GetMemoryInput{
		MemoryId: aws.String(id),
	}

	return findMemory(ctx, conn, &input)
}

func findMemory(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetMemoryInput) (*awstypes.Memory, error) {
	out, err := conn.GetMemory(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Memory == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Memory, nil
}

type memoryResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	Description             types.String                                                  `tfsdk:"description"`
	EncryptionKeyARN        fwtypes.ARN                                                   `tfsdk:"encryption_key_arn"`
	EventExpiryDuration     types.Int32                                                   `tfsdk:"event_expiry_duration"`
	ID                      types.String                                                  `tfsdk:"id"`
	IndexedKeys             fwtypes.SetNestedObjectValueOf[indexedKeyModel]               `tfsdk:"indexed_key"`
	MemoryExecutionRoleARN  fwtypes.ARN                                                   `tfsdk:"memory_execution_role_arn"`
	Name                    types.String                                                  `tfsdk:"name"`
	StreamDeliveryResources fwtypes.ListNestedObjectValueOf[streamDeliveryResourcesModel] `tfsdk:"stream_delivery_resources"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

type indexedKeyModel struct {
	Key  types.String                                   `tfsdk:"key"`
	Type fwtypes.StringEnum[awstypes.MetadataValueType] `tfsdk:"type"`
}

func indexedKeyIdentity(m *indexedKeyModel) string {
	return m.Key.ValueString() + "|" + m.Type.ValueString()
}

// requiresReplaceIfIndexedKeyRemoved forces replacement only when an existing indexed key is removed
// or changed. Indexed keys can only be added via UpdateMemory (previously indexed keys cannot be
// removed), so a plan that is a superset of the prior keys is applied in place.
func requiresReplaceIfIndexedKeyRemoved(ctx context.Context, request planmodifier.SetRequest, response *setplanmodifier.RequiresReplaceIfFuncResponse) {
	if request.StateValue.IsNull() || request.StateValue.IsUnknown() || request.PlanValue.IsUnknown() {
		return
	}

	var stateVal, planVal fwtypes.SetNestedObjectValueOf[indexedKeyModel]
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.GetAttribute(ctx, request.Path, &stateVal))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, request.Path, &planVal))
	if response.Diagnostics.HasError() {
		return
	}

	stateKeys, d := stateVal.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	planKeys, d := planVal.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	planned := make(map[string]struct{}, len(planKeys))
	for _, k := range planKeys {
		planned[indexedKeyIdentity(k)] = struct{}{}
	}
	for _, k := range stateKeys {
		if _, ok := planned[indexedKeyIdentity(k)]; !ok {
			response.RequiresReplace = true
			return
		}
	}
}

type streamDeliveryResourcesModel struct {
	Resources fwtypes.ListNestedObjectValueOf[streamDeliveryResourceModel] `tfsdk:"resource"`
}

// streamDeliveryResourceModel maps to the awstypes.StreamDeliveryResource union.
type streamDeliveryResourceModel struct {
	Kinesis fwtypes.ListNestedObjectValueOf[kinesisResourceModel] `tfsdk:"kinesis"`
}

var (
	_ fwflex.Expander  = streamDeliveryResourceModel{}
	_ fwflex.Flattener = &streamDeliveryResourceModel{}
)

func (m *streamDeliveryResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.StreamDeliveryResourceMemberKinesis:
		var data kinesisResourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.Kinesis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("stream delivery resource flatten: %T", v),
		)
	}
	return diags
}

func (m streamDeliveryResourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Kinesis.IsNull():
		data, d := m.Kinesis.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.StreamDeliveryResourceMemberKinesis
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type kinesisResourceModel struct {
	ContentConfigurations fwtypes.ListNestedObjectValueOf[contentConfigurationModel] `tfsdk:"content_configuration"`
	DataStreamARN         fwtypes.ARN                                                `tfsdk:"data_stream_arn"`
}

type contentConfigurationModel struct {
	Level fwtypes.StringEnum[awstypes.ContentLevel] `tfsdk:"level"`
	Type  fwtypes.StringEnum[awstypes.ContentType]  `tfsdk:"type"`
}
