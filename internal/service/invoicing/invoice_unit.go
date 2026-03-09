// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package invoicing

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/invoicing/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_invoicing_invoice_unit", name="Invoice Unit")
// @Tags(identifierAttribute="arn")
// @Region(overrideDeprecated=true)
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/invoicing;invoicing.GetInvoiceUnitOutput")
// @Testing(requireEnvVar="INVOICING_INVOICE_TESTS_ENABLED")
// @Testing(preIdentityVersion="6.28.0")
// @Testing(tagsTest=false)
// @Testing(serialize=true)
func newInvoiceUnitResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &invoiceUnitResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type invoiceUnitResource struct {
	framework.ResourceWithModel[invoiceUnitResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *invoiceUnitResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

type invoiceUnitResourceModel struct {
	framework.WithRegionModel
	ARN                    types.String                               `tfsdk:"arn"`
	Description            types.String                               `tfsdk:"description"`
	InvoiceReceiver        types.String                               `tfsdk:"invoice_receiver"`
	LastModified           timetypes.RFC3339                          `tfsdk:"last_modified"`
	Name                   types.String                               `tfsdk:"name"`
	Rule                   fwtypes.ListNestedObjectValueOf[ruleModel] `tfsdk:"rule"`
	TaxInheritanceDisabled types.Bool                                 `tfsdk:"tax_inheritance_disabled"`
	Tags                   tftags.Map                                 `tfsdk:"tags"`
	TagsAll                tftags.Map                                 `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                             `tfsdk:"timeouts"`
}

type ruleModel struct {
	LinkedAccounts fwtypes.SetValueOf[types.String] `tfsdk:"linked_accounts"`
}

func (r *invoiceUnitResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data invoiceUnitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	input := invoicing.CreateInvoiceUnitInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ResourceTags = getTagsIn(ctx)

	output, err := conn.CreateInvoiceUnit(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.InvoiceUnitArn)

	invoiceUnit, err := waitInvoiceUnitCreated(ctx, conn, data.ARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrARN), data.ARN) // Set 'arn' so as to taint the resource.
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	// Set values for unknowns after creation is complete
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, invoiceUnit, &data, fwflex.WithFieldNamePrefix("InvoiceUnit")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *invoiceUnitResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data invoiceUnitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	output, err := findInvoiceUnitByARN(ctx, conn, data.ARN.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix("InvoiceUnit")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *invoiceUnitResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new invoiceUnitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Rule.Equal(old.Rule) ||
		!new.TaxInheritanceDisabled.Equal(old.TaxInheritanceDisabled) {
		input := invoicing.UpdateInvoiceUnitInput{
			InvoiceUnitArn: new.ARN.ValueStringPointer(),
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateInvoiceUnit(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ARN.ValueString())
			return
		}

		var output *invoicing.GetInvoiceUnitOutput
		output, err = waitInvoiceUnitUpdated(ctx, conn, new.ARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ARN.ValueString())
			return
		}
		new.LastModified = fwflex.TimeToFramework(ctx, output.LastModified)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, new))
}

func (r *invoiceUnitResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data invoiceUnitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().InvoicingClient(ctx)

	input := invoicing.DeleteInvoiceUnitInput{
		InvoiceUnitArn: data.ARN.ValueStringPointer(),
	}
	_, err := conn.DeleteInvoiceUnit(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	_, err = waitInvoiceUnitDeleted(ctx, conn, data.ARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ARN.ValueString())
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
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func waitInvoiceUnitCreated(ctx context.Context, conn *invoicing.Client, arn string, timeout time.Duration) (*invoicing.GetInvoiceUnitOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{"AVAILABLE"},
		Refresh: statusInvoiceUnit(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*invoicing.GetInvoiceUnitOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitInvoiceUnitUpdated(ctx context.Context, conn *invoicing.Client, arn string, timeout time.Duration) (*invoicing.GetInvoiceUnitOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{"AVAILABLE"},
		Refresh: statusInvoiceUnit(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*invoicing.GetInvoiceUnitOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitInvoiceUnitDeleted(ctx context.Context, conn *invoicing.Client, arn string, timeout time.Duration) (*invoicing.GetInvoiceUnitOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"AVAILABLE"},
		Target:  []string{},
		Refresh: statusInvoiceUnit(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*invoicing.GetInvoiceUnitOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusInvoiceUnit(_ context.Context, conn *invoicing.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInvoiceUnitByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, "AVAILABLE", nil
	}
}
