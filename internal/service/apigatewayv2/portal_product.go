// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_apigatewayv2_portal_product", name="Portal Product")
// @Tags(identifierAttribute="portal_product_arn")
// @IdentityAttribute("portal_product_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetPortalProductOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="portal_product_id")
func newPortalProductResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &portalProductResource{}, nil
}

type portalProductResource struct {
	framework.ResourceWithModel[portalProductResourceModel]
	framework.WithImportByIdentity
}

func (r *portalProductResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"last_modified": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"portal_product_arn": framework.ARNAttributeComputedOnly(),
			"portal_product_id":  framework.IDAttribute(),
			names.AttrTags:       tftags.TagsAttribute(),
			names.AttrTagsAll:    tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *portalProductResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data portalProductResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	var input apigatewayv2.CreatePortalProductInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePortalProduct(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.DisplayName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *portalProductResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data portalProductResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalProductID)
	out, err := findPortalProductByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *portalProductResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state portalProductResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	id := fwflex.StringValueFromFramework(ctx, plan.PortalProductID)

	if diff.HasChanges() {
		var input apigatewayv2.UpdatePortalProductInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// UpdatePortalProduct is a PATCH: nil fields are omitted rather than cleared,
		// so removing the argument would silently leave the old description in place.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.Description = aws.String("")
		}

		out, err := conn.UpdatePortalProduct(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		plan.LastModified = state.LastModified
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *portalProductResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data portalProductResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalProductID)
	input := apigatewayv2.DeletePortalProductInput{
		PortalProductId: aws.String(id),
	}

	_, err := conn.DeletePortalProduct(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

// flatten is shared by Create, Read, Update and the list resource.
func (r *portalProductResource) flatten(ctx context.Context, portalProduct any, data *portalProductResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, portalProduct, data)...)

	return diags
}

func findPortalProductByID(ctx context.Context, conn *apigatewayv2.Client, id string) (*apigatewayv2.GetPortalProductOutput, error) {
	input := apigatewayv2.GetPortalProductInput{
		PortalProductId: aws.String(id),
	}

	return findPortalProduct(ctx, conn, &input)
}

func findPortalProduct(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetPortalProductInput) (*apigatewayv2.GetPortalProductOutput, error) {
	out, err := conn.GetPortalProduct(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type portalProductResourceModel struct {
	framework.WithRegionModel
	Description      types.String      `tfsdk:"description"`
	DisplayName      types.String      `tfsdk:"display_name"`
	LastModified     timetypes.RFC3339 `tfsdk:"last_modified"`
	PortalProductARN types.String      `tfsdk:"portal_product_arn"`
	PortalProductID  types.String      `tfsdk:"portal_product_id"`
	Tags             tftags.Map        `tfsdk:"tags"`
	TagsAll          tftags.Map        `tfsdk:"tags_all"`
}
