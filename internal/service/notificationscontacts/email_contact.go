// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationscontacts

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notificationscontacts/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notificationscontacts_email_contact", name="Email Contact")
// @Tags(identifierAttribute="arn")
func newEmailContactResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &emailContactResource{}

	return r, nil
}

type emailContactResource struct {
	framework.ResourceWithModel[emailContactResourceModel]
}

func (r *emailContactResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"email_address": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 254),
					stringvalidator.RegexMatches(regexache.MustCompile(`(.+)@(.+)`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w.~-]+`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *emailContactResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data emailContactResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsContactsClient(ctx)

	var inputCEC notificationscontacts.CreateEmailContactInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &inputCEC)...)
	if response.Diagnostics.HasError() {
		return
	}
	inputCEC.Tags = getTagsIn(ctx)

	output, err := conn.CreateEmailContact(ctx, &inputCEC)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Contacts Email Contact (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.Arn)
	data.ARN = fwflex.StringValueToFramework(ctx, arn)

	inputSAC := notificationscontacts.SendActivationCodeInput{
		Arn: aws.String(arn),
	}
	_, err = conn.SendActivationCode(ctx, &inputSAC)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("activating User Notifications Contacts Email Contact (%s)", arn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *emailContactResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data emailContactResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsContactsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	output, err := findEmailContactByARN(ctx, conn, arn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Contacts Email Contact (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.EmailAddress = fwflex.StringToFramework(ctx, output.Address)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *emailContactResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data emailContactResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsContactsClient(ctx)

	input := notificationscontacts.DeleteEmailContactInput{
		Arn: fwflex.StringFromFramework(ctx, data.ARN),
	}
	_, err := conn.DeleteEmailContact(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Contacts Email Contact (%s)", data.ARN.ValueString()), err.Error())

		return
	}
}

func (r *emailContactResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func findEmailContactByARN(ctx context.Context, conn *notificationscontacts.Client, arn string) (*awstypes.EmailContact, error) {
	input := notificationscontacts.GetEmailContactInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetEmailContact(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.EmailContact == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return output.EmailContact, nil
}

type emailContactResourceModel struct {
	ARN          types.String `tfsdk:"arn"`
	EmailAddress types.String `tfsdk:"email_address"`
	Name         types.String `tfsdk:"name"`
	Tags         tftags.Map   `tfsdk:"tags"`
	TagsAll      tftags.Map   `tfsdk:"tags_all"`
}
