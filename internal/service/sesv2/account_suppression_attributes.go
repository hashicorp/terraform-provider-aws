// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_account_suppression_attributes", name="Account Suppression Attributes")
func newAccountSuppressionAttributesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountSuppressionAttributesResource{}

	return r, nil
}

type accountSuppressionAttributesResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
	framework.WithImportByID
}

func (r *accountSuppressionAttributesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"suppressed_reasons": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.SuppressionListReason]](ctx),
				Required:    true,
				ElementType: fwtypes.StringEnumType[awstypes.SuppressionListReason](),
			},
		},
	}
}

func (r *accountSuppressionAttributesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	id := r.Meta().AccountID(ctx)
	input := &sesv2.PutAccountSuppressionAttributesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutAccountSuppressionAttributes(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SESv2 Account Suppression Attributes (%s)", id), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSuppressionAttributesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	suppressionAttributes, err := findAccountSuppressionAttributes(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SESv2 Account Suppression Attributes (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, suppressionAttributes, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSuppressionAttributesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	input := &sesv2.PutAccountSuppressionAttributesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutAccountSuppressionAttributes(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating SESv2 Account Suppression Attributes (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func findAccountSuppressionAttributes(ctx context.Context, conn *sesv2.Client) (*awstypes.SuppressionAttributes, error) {
	output, err := findAccount(ctx, conn)

	if err != nil {
		return nil, err
	}

	if output.SuppressionAttributes == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.SuppressionAttributes, nil
}

type accountSuppressionAttributesResourceModel struct {
	ID                types.String                                                           `tfsdk:"id"`
	SuppressedReasons fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.SuppressionListReason]] `tfsdk:"suppressed_reasons"`
}
