// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_opt_out_list", name="Opt-out List")
// @Tags(identifierAttribute="arn")
func newOptOutListResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &optOutListResource{}

	return r, nil
}

type optOutListResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[optOutListResourceModel]
	framework.WithImportByID
}

func (r *optOutListResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{0,63}$`), ""),
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

func (r *optOutListResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data optOutListResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := data.OptOutListName.ValueString()
	input := &pinpointsmsvoicev2.CreateOptOutListInput{
		ClientToken:    aws.String(sdkid.UniqueId()),
		OptOutListName: aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreateOptOutList(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating End User Messaging SMS Opt-out List (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.OptOutListARN = fwflex.StringToFramework(ctx, output.OptOutListArn)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *optOutListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data optOutListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findOptOutListByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading End User Messaging SMS Opt-out List (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *optOutListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data optOutListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	_, err := conn.DeleteOptOutList(ctx, &pinpointsmsvoicev2.DeleteOptOutListInput{
		OptOutListName: data.ID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting End User Messaging SMS Opt-out List (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type optOutListResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OptOutListARN  types.String `tfsdk:"arn"`
	OptOutListName types.String `tfsdk:"name"`
	Tags           tftags.Map   `tfsdk:"tags"`
	TagsAll        tftags.Map   `tfsdk:"tags_all"`
}

func (model *optOutListResourceModel) InitFromID() error {
	model.OptOutListName = model.ID

	return nil
}

func (model *optOutListResourceModel) setID() {
	model.ID = model.OptOutListName
}

func findOptOutListByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.OptOutListInformation, error) {
	input := &pinpointsmsvoicev2.DescribeOptOutListsInput{
		OptOutListNames: []string{id},
	}

	return findOptOutList(ctx, conn, input)
}

func findOptOutList(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeOptOutListsInput) (*awstypes.OptOutListInformation, error) {
	output, err := findOptOutLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOptOutLists(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeOptOutListsInput) ([]awstypes.OptOutListInformation, error) {
	var output []awstypes.OptOutListInformation

	pages := pinpointsmsvoicev2.NewDescribeOptOutListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.OptOutLists...)
	}

	return output, nil
}
