// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_collection_group", name="Collection Group")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types;types.CollectionGroupDetail")
func newCollectionGroupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &collectionGroupResource{}, nil
}

const (
	ResNameCollectionGroup = "Collection Group"
)

type collectionGroupResource struct {
	framework.ResourceWithModel[collectionGroupResourceModel]
	framework.WithImportByIdentity
}

type collectionGroupResourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                         `tfsdk:"arn"`
	CapacityLimits  fwtypes.ListNestedObjectValueOf[capacityLimitsModel] `tfsdk:"capacity_limits"`
	CreatedDate     timetypes.RFC3339                                    `tfsdk:"created_date"`
	Description     types.String                                         `tfsdk:"description"`
	ID              types.String                                         `tfsdk:"id"`
	Name            types.String                                         `tfsdk:"name"`
	StandbyReplicas fwtypes.StringEnum[awstypes.StandbyReplicas]         `tfsdk:"standby_replicas"`
	Tags            tftags.Map                                           `tfsdk:"tags"`
	TagsAll         tftags.Map                                           `tfsdk:"tags_all"`
}

type capacityLimitsModel struct {
	MinIndexingCapacityInOCU types.Float32 `tfsdk:"min_indexing_capacity_in_ocu"`
	MaxIndexingCapacityInOCU types.Float32 `tfsdk:"max_indexing_capacity_in_ocu"`
	MinSearchCapacityInOCU   types.Float32 `tfsdk:"min_search_capacity_in_ocu"`
	MaxSearchCapacityInOCU   types.Float32 `tfsdk:"max_search_capacity_in_ocu"`
}

func (r *collectionGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			"capacity_limits": framework.ResourceOptionalComputedListOfObjectsAttribute[capacityLimitsModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "Date the collection group was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the collection group.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Description: "Name of the collection group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z][0-9a-z-]+$`),
						`must start with any lower case letter and can include any lower case letter, number, or "-"`),
				},
			},
			"standby_replicas": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.StandbyReplicas](),
				Description: "Indicates whether standby replicas should be used for collections in this group. One of `ENABLED` or `DISABLED`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *collectionGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan collectionGroupResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	input := opensearchserverless.CreateCollectionGroupInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateCollectionGroup(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	if output == nil || output.CreateCollectionGroupDetail == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.CreateCollectionGroupDetail, &plan, fwflex.WithIgnoredFieldNamesAppend("CreatedDate")))
	if response.Diagnostics.HasError() {
		return
	}

	plan.CreatedDate = timetypes.NewRFC3339ValueMust(flex.Int64ToRFC3339StringValue(output.CreateCollectionGroupDetail.CreatedDate))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectionGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionGroupResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	input := opensearchserverless.BatchGetCollectionGroupInput{
		Ids: []string{state.ID.ValueString()},
	}
	output, err := findCollectionGroup(ctx, conn, &input)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &state, fwflex.WithIgnoredFieldNamesAppend("CreatedDate")))
	if response.Diagnostics.HasError() {
		return
	}

	state.CreatedDate = timetypes.NewRFC3339ValueMust(flex.Int64ToRFC3339StringValue(output.CreatedDate))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *collectionGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state collectionGroupResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input opensearchserverless.UpdateCollectionGroupInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateCollectionGroup(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		if output == nil || output.UpdateCollectionGroupDetail == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectionGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionGroupResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCollectionGroup(ctx, &opensearchserverless.DeleteCollectionGroupInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		Id:          state.ID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}
