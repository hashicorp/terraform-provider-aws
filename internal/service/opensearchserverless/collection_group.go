// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_collection_group", name="Collection Group")
// @Tags(identifierAttribute="arn")
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
	ARN             types.String                           `tfsdk:"arn"`
	CapacityLimits  fwtypes.ObjectValueOf[capacityLimitsModel] `tfsdk:"capacity_limits"`
	CreatedDate     types.Int64                            `tfsdk:"created_date"`
	Description     types.String                           `tfsdk:"description"`
	ID              types.String                           `tfsdk:"id"`
	Name            types.String                           `tfsdk:"name"`
	StandbyReplicas types.String                           `tfsdk:"standby_replicas"`
	Tags            tftags.Map                             `tfsdk:"tags"`
	TagsAll         tftags.Map                             `tfsdk:"tags_all"`
}

type capacityLimitsModel struct {
	MinIndexingCapacityInOCU types.Float64 `tfsdk:"min_indexing_capacity_in_ocu"`
	MaxIndexingCapacityInOCU types.Float64 `tfsdk:"max_indexing_capacity_in_ocu"`
	MinSearchCapacityInOCU   types.Float64 `tfsdk:"min_search_capacity_in_ocu"`
	MaxSearchCapacityInOCU   types.Float64 `tfsdk:"max_search_capacity_in_ocu"`
}

func (r *collectionGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"capacity_limits": schema.ObjectAttribute{
				CustomType:  fwtypes.NewObjectTypeOf[capacityLimitsModel](ctx),
				Description: "Capacity limits for the collection group.",
				Optional:    true,
			},
			"created_date": schema.Int64Attribute{
				Description: "Date the collection group was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
				Description: "Indicates whether standby replicas should be used for collections in this group. One of `ENABLED` or `DISABLED`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.StandbyReplicas](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *collectionGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan collectionGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	in := &opensearchserverless.CreateCollectionGroupInput{
		Name:            plan.Name.ValueStringPointer(),
		Description:     plan.Description.ValueStringPointer(),
		StandbyReplicas: awstypes.StandbyReplicas(plan.StandbyReplicas.ValueString()),
		ClientToken:     aws.String(id.UniqueId()),
		Tags:            getTagsIn(ctx),
	}

	// Only include capacity_limits if it was specified in the plan
	if !plan.CapacityLimits.IsNull() {
		capacityLimitsPtr, d := plan.CapacityLimits.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		capacityLimits := *capacityLimitsPtr

		in.CapacityLimits = &awstypes.CollectionGroupCapacityLimits{}
		if !capacityLimits.MinIndexingCapacityInOCU.IsNull() {
			in.CapacityLimits.MinIndexingCapacityInOCU = aws.Float32(float32(capacityLimits.MinIndexingCapacityInOCU.ValueFloat64()))
		}
		if !capacityLimits.MaxIndexingCapacityInOCU.IsNull() {
			in.CapacityLimits.MaxIndexingCapacityInOCU = aws.Float32(float32(capacityLimits.MaxIndexingCapacityInOCU.ValueFloat64()))
		}
		if !capacityLimits.MinSearchCapacityInOCU.IsNull() {
			in.CapacityLimits.MinSearchCapacityInOCU = aws.Float32(float32(capacityLimits.MinSearchCapacityInOCU.ValueFloat64()))
		}
		if !capacityLimits.MaxSearchCapacityInOCU.IsNull() {
			in.CapacityLimits.MaxSearchCapacityInOCU = aws.Float32(float32(capacityLimits.MaxSearchCapacityInOCU.ValueFloat64()))
		}
	}

	out, err := conn.CreateCollectionGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameCollectionGroup, plan.Name.ValueString(), nil),
			err.Error(),
		)
		return
	}

	state := plan

	// Manually flatten the response
	state.ARN = types.StringPointerValue(out.CreateCollectionGroupDetail.Arn)
	state.CreatedDate = types.Int64PointerValue(out.CreateCollectionGroupDetail.CreatedDate)
	state.Description = types.StringPointerValue(out.CreateCollectionGroupDetail.Description)
	state.ID = types.StringPointerValue(out.CreateCollectionGroupDetail.Id)
	state.Name = types.StringPointerValue(out.CreateCollectionGroupDetail.Name)
	state.StandbyReplicas = types.StringValue(string(out.CreateCollectionGroupDetail.StandbyReplicas))

	// Only populate capacity_limits if it was specified in the plan
	// AWS returns default values even when not configured, so we must preserve plan intent
	if !plan.CapacityLimits.IsNull() && out.CreateCollectionGroupDetail.CapacityLimits != nil {
		capacityLimits := capacityLimitsModel{
			MinIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.CreateCollectionGroupDetail.CapacityLimits.MinIndexingCapacityInOCU))),
			MaxIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.CreateCollectionGroupDetail.CapacityLimits.MaxIndexingCapacityInOCU))),
			MinSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.CreateCollectionGroupDetail.CapacityLimits.MinSearchCapacityInOCU))),
			MaxSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.CreateCollectionGroupDetail.CapacityLimits.MaxSearchCapacityInOCU))),
		}
		state.CapacityLimits = fwtypes.NewObjectValueOfMust(ctx, &capacityLimits)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *collectionGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if this is an import (ARN will be empty during import)
	isImport := state.ARN.IsNull()
	
	// Remember if capacity_limits was null before reading from API
	capacityLimitsWasNull := state.CapacityLimits.IsNull()

	out, err := findCollectionGroupByID(ctx, conn, state.ID.ValueString())
	if err != nil {
		if retry.NotFound(err) {
			tflog.Warn(ctx, "OpenSearchServerless CollectionGroup not found, removing from state", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, ResNameCollectionGroup, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		tflog.Warn(ctx, "OpenSearchServerless CollectionGroup response is nil, removing from state", map[string]interface{}{
			"id": state.ID.ValueString(),
		})
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(nil))
		resp.State.RemoveResource(ctx)
		return
	}

	// Manually flatten the response
	state.ARN = types.StringPointerValue(out.Arn)
	state.CreatedDate = types.Int64PointerValue(out.CreatedDate)
	state.Description = types.StringPointerValue(out.Description)
	state.ID = types.StringPointerValue(out.Id)
	state.Name = types.StringPointerValue(out.Name)
	state.StandbyReplicas = types.StringValue(string(out.StandbyReplicas))

	// Populate capacity_limits if:
	// 1. This is an import AND the values are non-default (not just API defaults), OR
	// 2. It was originally set (not null in current state)
	// API defaults are: min=0, max=96 for both indexing and search
	shouldPopulateCapacityLimits := false
	if isImport && out.CapacityLimits != nil {
		// Check if any values differ from defaults
		minIndexing := aws.ToFloat32(out.CapacityLimits.MinIndexingCapacityInOCU)
		maxIndexing := aws.ToFloat32(out.CapacityLimits.MaxIndexingCapacityInOCU)
		minSearch := aws.ToFloat32(out.CapacityLimits.MinSearchCapacityInOCU)
		maxSearch := aws.ToFloat32(out.CapacityLimits.MaxSearchCapacityInOCU)
		
		// Only populate if values differ from defaults (0 for min, 96 for max)
		if minIndexing != 0 || maxIndexing != 96 || minSearch != 0 || maxSearch != 96 {
			shouldPopulateCapacityLimits = true
		}
	} else if !capacityLimitsWasNull {
		shouldPopulateCapacityLimits = true
	}
	
	if shouldPopulateCapacityLimits && out.CapacityLimits != nil {
		capacityLimits := capacityLimitsModel{
			MinIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.CapacityLimits.MinIndexingCapacityInOCU))),
			MaxIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.CapacityLimits.MaxIndexingCapacityInOCU))),
			MinSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.CapacityLimits.MinSearchCapacityInOCU))),
			MaxSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.CapacityLimits.MaxSearchCapacityInOCU))),
		}
		state.CapacityLimits = fwtypes.NewObjectValueOfMust(ctx, &capacityLimits)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *collectionGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state collectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if description or capacity_limits changed
	if !plan.Description.Equal(state.Description) || !plan.CapacityLimits.Equal(state.CapacityLimits) {
		input := &opensearchserverless.UpdateCollectionGroupInput{
			Id:          plan.ID.ValueStringPointer(),
			Description: plan.Description.ValueStringPointer(),
			ClientToken: aws.String(id.UniqueId()),
		}

		// Only include capacity_limits if it was specified in the plan
		if !plan.CapacityLimits.IsNull() {
			capacityLimitsPtr, d := plan.CapacityLimits.ToPtr(ctx)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			capacityLimits := *capacityLimitsPtr

			input.CapacityLimits = &awstypes.CollectionGroupCapacityLimits{}
			if !capacityLimits.MinIndexingCapacityInOCU.IsNull() {
				input.CapacityLimits.MinIndexingCapacityInOCU = aws.Float32(float32(capacityLimits.MinIndexingCapacityInOCU.ValueFloat64()))
			}
			if !capacityLimits.MaxIndexingCapacityInOCU.IsNull() {
				input.CapacityLimits.MaxIndexingCapacityInOCU = aws.Float32(float32(capacityLimits.MaxIndexingCapacityInOCU.ValueFloat64()))
			}
			if !capacityLimits.MinSearchCapacityInOCU.IsNull() {
				input.CapacityLimits.MinSearchCapacityInOCU = aws.Float32(float32(capacityLimits.MinSearchCapacityInOCU.ValueFloat64()))
			}
			if !capacityLimits.MaxSearchCapacityInOCU.IsNull() {
				input.CapacityLimits.MaxSearchCapacityInOCU = aws.Float32(float32(capacityLimits.MaxSearchCapacityInOCU.ValueFloat64()))
			}
		}

		out, err := conn.UpdateCollectionGroup(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionUpdating, ResNameCollectionGroup, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		// Manually flatten the response
		plan.ARN = types.StringPointerValue(out.UpdateCollectionGroupDetail.Arn)
		plan.CreatedDate = types.Int64PointerValue(out.UpdateCollectionGroupDetail.CreatedDate)
		plan.Description = types.StringPointerValue(out.UpdateCollectionGroupDetail.Description)
		plan.ID = types.StringPointerValue(out.UpdateCollectionGroupDetail.Id)
		plan.Name = types.StringPointerValue(out.UpdateCollectionGroupDetail.Name)

		// Only populate capacity_limits if it was specified in the plan
		// AWS returns default values even when not configured, so we must preserve plan intent
		if !plan.CapacityLimits.IsNull() && out.UpdateCollectionGroupDetail.CapacityLimits != nil {
			capacityLimits := capacityLimitsModel{
				MinIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.UpdateCollectionGroupDetail.CapacityLimits.MinIndexingCapacityInOCU))),
				MaxIndexingCapacityInOCU: types.Float64Value(float64(aws.ToFloat32(out.UpdateCollectionGroupDetail.CapacityLimits.MaxIndexingCapacityInOCU))),
				MinSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.UpdateCollectionGroupDetail.CapacityLimits.MinSearchCapacityInOCU))),
				MaxSearchCapacityInOCU:   types.Float64Value(float64(aws.ToFloat32(out.UpdateCollectionGroupDetail.CapacityLimits.MaxSearchCapacityInOCU))),
			}
			plan.CapacityLimits = fwtypes.NewObjectValueOfMust(ctx, &capacityLimits)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *collectionGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCollectionGroup(ctx, &opensearchserverless.DeleteCollectionGroupInput{
		ClientToken: aws.String(id.UniqueId()),
		Id:          state.ID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameCollectionGroup, state.Name.ValueString(), nil),
			err.Error(),
		)
	}
}
