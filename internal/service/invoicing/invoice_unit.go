// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package invoicing

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/invoicing/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_invoicing_invoice_unit", name="Invoice Unit")
// @Tags(identifierAttribute="arn")
func newInvoiceUnitResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &invoiceUnitResource{}

	return r, nil
}

type invoiceUnitResource struct {
	framework.ResourceWithModel[invoiceUnitResourceModel]
	framework.WithImportByID
}

func (r *invoiceUnitResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"invoice_receiver": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_modified": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tax_inheritance_disabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"linked_accounts": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

type invoiceUnitResourceModel struct {
	framework.WithRegionModel
	ARN                    types.String                               `tfsdk:"arn"`
	Description            types.String                               `tfsdk:"description"`
	ID                     types.String                               `tfsdk:"id"`
	InvoiceReceiver        types.String                               `tfsdk:"invoice_receiver"`
	LastModified           timetypes.RFC3339                          `tfsdk:"last_modified"`
	Name                   types.String                               `tfsdk:"name"`
	Rule                   fwtypes.ListNestedObjectValueOf[ruleModel] `tfsdk:"rule"`
	TaxInheritanceDisabled types.Bool                                 `tfsdk:"tax_inheritance_disabled"`
	Tags                   tftags.Map                                 `tfsdk:"tags"`
	TagsAll                tftags.Map                                 `tfsdk:"tags_all"`
}

type ruleModel struct {
	LinkedAccounts fwtypes.SetValueOf[types.String] `tfsdk:"linked_accounts"`
}

func (r *invoiceUnitResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data invoiceUnitResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	input := invoicing.CreateInvoiceUnitInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ResourceTags = getTagsIn(ctx)

	output, err := conn.CreateInvoiceUnit(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating Invoice Unit", err.Error())
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.InvoiceUnitArn)
	data.ARN = fwflex.StringToFramework(ctx, output.InvoiceUnitArn)

	// Read the resource to get computed values
	invoiceUnit, err := findInvoiceUnitByARN(ctx, conn, data.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("reading Invoice Unit after create", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, invoiceUnit, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringToFramework(ctx, invoiceUnit.InvoiceUnitArn)
	data.ARN = fwflex.StringToFramework(ctx, invoiceUnit.InvoiceUnitArn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *invoiceUnitResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data invoiceUnitResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	output, err := findInvoiceUnitByARN(ctx, conn, data.ID.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("reading Invoice Unit", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.InvoiceUnitArn)
	data.ARN = fwflex.StringToFramework(ctx, output.InvoiceUnitArn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *invoiceUnitResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new invoiceUnitResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Rule.Equal(old.Rule) ||
		!new.TaxInheritanceDisabled.Equal(old.TaxInheritanceDisabled) {

		input := invoicing.UpdateInvoiceUnitInput{
			InvoiceUnitArn: new.ID.ValueStringPointer(),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateInvoiceUnit(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError("updating Invoice Unit", err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *invoiceUnitResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data invoiceUnitResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	input := invoicing.DeleteInvoiceUnitInput{
		InvoiceUnitArn: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteInvoiceUnit(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		response.Diagnostics.AddError("deleting Invoice Unit", err.Error())
	}
}

func findInvoiceUnitByARN(ctx context.Context, conn *invoicing.Client, arn string) (*invoicing.GetInvoiceUnitOutput, error) {
	input := invoicing.GetInvoiceUnitInput{
		InvoiceUnitArn: aws.String(arn),
	}

	output, err := conn.GetInvoiceUnit(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
