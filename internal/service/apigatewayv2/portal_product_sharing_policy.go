// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_apigatewayv2_portal_product_sharing_policy", name="Portal Product Sharing Policy")
// @IdentityAttribute("portal_product_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetPortalProductSharingPolicyOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="portal_product_id")
// API Gateway normalizes the policy document, and import verification compares strings.
// @Testing(importIgnore="policy_document")
func newPortalProductSharingPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &portalProductSharingPolicyResource{}, nil
}

type portalProductSharingPolicyResource struct {
	framework.ResourceWithModel[portalProductSharingPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *portalProductSharingPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_document": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"portal_product_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *portalProductSharingPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data portalProductSharingPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalProductID)
	if err := putPortalProductSharingPolicy(ctx, conn, &data); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *portalProductSharingPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data portalProductSharingPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalProductID)
	out, err := findPortalProductSharingPolicyByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	data.PolicyDocument = fwtypes.IAMPolicyValue(aws.ToString(out.PolicyDocument))
	data.PortalProductID = fwflex.StringToFramework(ctx, out.PortalProductId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *portalProductSharingPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old portalProductSharingPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	if !new.PolicyDocument.Equal(old.PolicyDocument) {
		id := fwflex.StringValueFromFramework(ctx, new.PortalProductID)
		if err := putPortalProductSharingPolicy(ctx, conn, &new); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *portalProductSharingPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data portalProductSharingPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalProductID)
	input := apigatewayv2.DeletePortalProductSharingPolicyInput{
		PortalProductId: aws.String(id),
	}
	_, err := conn.DeletePortalProductSharingPolicy(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

// putPortalProductSharingPolicy creates or replaces the sharing policy. PutPortalProductSharingPolicy returns no output.
func putPortalProductSharingPolicy(ctx context.Context, conn *apigatewayv2.Client, data *portalProductSharingPolicyResourceModel) error {
	input := apigatewayv2.PutPortalProductSharingPolicyInput{
		PolicyDocument:  fwflex.StringFromFramework(ctx, data.PolicyDocument),
		PortalProductId: fwflex.StringFromFramework(ctx, data.PortalProductID),
	}

	if _, err := conn.PutPortalProductSharingPolicy(ctx, &input); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

func findPortalProductSharingPolicyByID(ctx context.Context, conn *apigatewayv2.Client, id string) (*apigatewayv2.GetPortalProductSharingPolicyOutput, error) {
	input := apigatewayv2.GetPortalProductSharingPolicyInput{
		PortalProductId: aws.String(id),
	}

	return findPortalProductSharingPolicy(ctx, conn, &input)
}

func findPortalProductSharingPolicy(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetPortalProductSharingPolicyInput) (*apigatewayv2.GetPortalProductSharingPolicyOutput, error) {
	out, err := conn.GetPortalProductSharingPolicy(ctx, input)

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

	// A portal product without a sharing policy returns an empty JSON object.
	if v := aws.ToString(out.PolicyDocument); v == "" || v == "{}" {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: "no sharing policy",
		})
	}

	return out, nil
}

type portalProductSharingPolicyResourceModel struct {
	framework.WithRegionModel
	PolicyDocument  fwtypes.IAMPolicy `tfsdk:"policy_document"`
	PortalProductID types.String      `tfsdk:"portal_product_id"`
}
