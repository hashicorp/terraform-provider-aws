// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"
	awstypes "github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_paymentcryptography_key_alias",name="Key Alias")
func newResourceKeyAlias(context.Context) (resource.ResourceWithConfigure, error) {
	r := &keyAliasResource{}

	return r, nil
}

type keyAliasResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*keyAliasResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_paymentcryptography_key_alias"
}

func (r *keyAliasResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alias_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^alias/[a-zA-Z0-9/\-_]+$`), "An alias must begin with alias/ followed by a name, for example alias/ExampleAlias. It can contain only alphanumeric characters, forward slashes (/), underscores (_), and dashes (-)"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_arn": schema.StringAttribute{
				Optional: true,
				// Validators: []validator.String{
				// 	stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws:payment-cryptography:[a-z]{2}-[a-z]{1,16}-[0-9]+:[0-9]{12}:key/[0-9a-zA-Z]{16,64}$`), "valid arn is required. Minimum length of 70. Maximum length of 150"),
				// },
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *keyAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data keyAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PaymentCryptographyClient(ctx)

	input := &paymentcryptography.CreateAliasInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateAlias(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating PaymentCryptography Alias", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Alias.AliasName)

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Alias, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data keyAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().PaymentCryptographyClient(ctx)

	output, err := findkeyAliasByName(ctx, conn, data.AliasName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading PaymentCryptography Alias (%s)", data.ID.String()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyAliasResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new keyAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PaymentCryptographyClient(ctx)

	if !new.AliasName.Equal(old.AliasName) ||
		!new.KeyARN.Equal(old.KeyARN) {
		input := &paymentcryptography.UpdateAliasInput{}

		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAlias(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading PaymentCryptography key Alias (%s)", new.ID.String()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *keyAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data keyAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PaymentCryptographyClient(ctx)

	_, err := conn.DeleteAlias(ctx, &paymentcryptography.DeleteAliasInput{
		AliasName: fwflex.StringFromFramework(ctx, data.AliasName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting PaymentCryptography key Alias (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type keyAliasResourceModel struct {
	KeyARN    types.String `tfsdk:"key_arn"`
	AliasName types.String `tfsdk:"alias_name"`
	ID        types.String `tfsdk:"id"`
}

func (m *keyAliasResourceModel) InitFromID() error {
	m.AliasName = m.ID

	return nil
}

func findkeyAliasByName(ctx context.Context, conn *paymentcryptography.Client, name string) (*awstypes.Alias, error) {
	input := &paymentcryptography.GetAliasInput{
		AliasName: aws.String(name),
	}

	output, err := conn.GetAlias(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Alias == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Alias, nil
}
