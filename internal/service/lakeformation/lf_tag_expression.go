package lakeformation

import (
	"context"
	"fmt"
	"strings"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"

	lfTypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_lakeformation_lf_tag_expression", name="LF Tag Expression")
func newLFTagExpressionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &lfTagExpressionResource{
		ResourceWithModel: framework.ResourceWithModel[lfTagExpressionResourceModel]{},
	}, nil
}

type lfTagExpressionResource struct {
	framework.ResourceWithConfigure
	framework.ResourceWithModel[lfTagExpressionResourceModel]
}

func (r *lfTagExpressionResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages an AWS Lake Formation Tag Expression.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Primary identifier (catalog_id:name)",
			},
			"catalog_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the Data Catalog.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the LF-Tag Expression.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the LF-Tag Expression.",
			},
			"tag_expression": schema.MapAttribute{
				Required:    true,
				ElementType: types.SetType{ElemType: types.StringType},
				Description: "Mapping of tag keys to lists of allowed values.",
			},
		},
	}
}

type lfTagExpressionResourceModel struct {
	framework.WithRegionModel
	ID            types.String         `tfsdk:"id"`
	CatalogId     types.String         `tfsdk:"catalog_id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	TagExpression map[string]types.Set `tfsdk:"tag_expression"`
}

func (r *lfTagExpressionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data lfTagExpressionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LakeFormationClient(ctx)

	expr, expandDiags := expandLFTagExpression(ctx, data.TagExpression)
	response.Diagnostics.Append(expandDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get catalog ID, defaulting to account ID if not set
	catalogId := data.CatalogId.ValueString()
	if catalogId == "" {
		catalogId = r.Meta().AccountID(ctx)
		data.CatalogId = types.StringValue(catalogId)
	}

	input := &lakeformation.CreateLFTagExpressionInput{
		CatalogId: aws.String(catalogId),
		Name:      aws.String(data.Name.ValueString()),
		Description: func() *string {
			if data.Description.IsNull() {
				return nil
			}
			return aws.String(data.Description.ValueString())
		}(),
		Expression: expr,
	}

	_, err := conn.CreateLFTagExpression(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Creating LF-Tag Expression",
			fmt.Sprintf("Could not create LF-Tag Expression: %s", err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.CatalogId.ValueString(), data.Name.ValueString()))
	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *lfTagExpressionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data lfTagExpressionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LakeFormationClient(ctx)
	name := data.Name.ValueString()

	output, err := conn.GetLFTagExpression(ctx, &lakeformation.GetLFTagExpressionInput{
		CatalogId: aws.String(data.CatalogId.ValueString()),
		Name:      aws.String(name),
	})

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading LF-Tag Expression",
			fmt.Sprintf("Could not read LF-Tag Expression %s: %s", name, err),
		)
		return
	}

	// Manually populate the model fields from the API response
	if output.CatalogId != nil {
		data.CatalogId = types.StringValue(*output.CatalogId)
	}
	if output.Name != nil {
		data.Name = types.StringValue(*output.Name)
	}
	if output.Description != nil {
		data.Description = types.StringValue(*output.Description)
	}

	// Convert the Expression from AWS API format to types.Set map
	tagExprMap := make(map[string]types.Set)
	for _, tag := range output.Expression {
		if tag.TagKey != nil {
			setValue, diags := types.SetValueFrom(ctx, types.StringType, tag.TagValues)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			tagExprMap[*tag.TagKey] = setValue
		}
	}
	data.TagExpression = tagExprMap

	// Ensure ID is properly set (consistent with Create/Update)
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.CatalogId.ValueString(), data.Name.ValueString()))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *lfTagExpressionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan lfTagExpressionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	expr, expandDiags := expandLFTagExpression(ctx, plan.TagExpression)
	response.Diagnostics.Append(expandDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get catalog ID, defaulting to account ID if not set
	catalogId := plan.CatalogId.ValueString()
	if catalogId == "" {
		catalogId = r.Meta().AccountID(ctx)
		plan.CatalogId = types.StringValue(catalogId)
	}

	input := &lakeformation.UpdateLFTagExpressionInput{
		CatalogId: aws.String(catalogId),
		Name:      aws.String(plan.Name.ValueString()),
		Description: func() *string {
			if plan.Description.IsNull() {
				return nil
			}
			return aws.String(plan.Description.ValueString())
		}(),
		Expression: expr,
	}

	if _, err := r.Meta().LakeFormationClient(ctx).UpdateLFTagExpression(ctx, input); err != nil {
		response.Diagnostics.AddError(
			"Error Updating LF-Tag Expression",
			fmt.Sprintf("Could not update LF-Tag Expression: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.CatalogId.ValueString(), plan.Name.ValueString()))
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *lfTagExpressionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state lfTagExpressionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &lakeformation.DeleteLFTagExpressionInput{
		CatalogId: aws.String(state.CatalogId.ValueString()),
		Name:      aws.String(state.Name.ValueString()),
	}
	if _, err := r.Meta().LakeFormationClient(ctx).DeleteLFTagExpression(ctx, input); err != nil {
		if !tfresource.NotFound(err) {
			response.Diagnostics.AddError(
				"Error Deleting LF-Tag Expression",
				fmt.Sprintf("Could not delete LF-Tag Expression: %s", err),
			)
		}
	}
}

func (r *lfTagExpressionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID which should be in format "catalog_id:name"
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format 'catalog_id:name'",
		)
		return
	}

	catalogId := parts[0]
	name := parts[1]

	// Set the parsed values in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("catalog_id"), catalogId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// expandLFTagExpression converts the native Go map into the AWS LFTag slice.
func expandLFTagExpression(ctx context.Context, m map[string]types.Set) ([]lfTypes.LFTag, diag.Diagnostics) {
	var expr []lfTypes.LFTag
	var diags diag.Diagnostics

	// Sort the keys for deterministic ordering
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// For each key, sort its values and append to the LFTag list
	for _, k := range keys {
		set := m[k]
		var vals []string
		diags.Append(set.ElementsAs(ctx, &vals, false)...)
		if diags.HasError() {
			return nil, diags
		}

		sort.Strings(vals)
		expr = append(expr, lfTypes.LFTag{
			TagKey:    aws.String(k),
			TagValues: vals,
		})
	}

	return expr, diags
}
