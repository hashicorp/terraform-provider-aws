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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_configuration_set", name="Configuration Set")
// @Tags(identifierAttribute="arn")
func newConfigurationSetResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &configurationSetResource{}

	return r, nil
}

type configurationSetResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *configurationSetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"default_message_type": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.MessageType](),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^(TRANSACTIONAL|PROMOTIONAL)$`), "must be either TRANSACTIONAL or PROMOTIONAL"),
				},
			},
			"default_sender_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_-]{1,11}$`), "must be between 1 and 11 characters long and contain only letters, numbers, underscores, and dashes"),
				},
			},
			names.AttrID: framework.IDAttribute(),
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

func (r *configurationSetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data configurationSetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := data.ConfigurationSetName.ValueString()
	input := &pinpointsmsvoicev2.CreateConfigurationSetInput{
		ClientToken:          aws.String(sdkid.UniqueId()),
		ConfigurationSetName: aws.String(name),
		Tags:                 getTagsIn(ctx),
	}

	output, err := conn.CreateConfigurationSet(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating End User Messaging SMS Configuration Set (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ConfigurationSetARN = fwflex.StringToFramework(ctx, output.ConfigurationSetArn)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *configurationSetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data configurationSetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findConfigurationSetByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading End User Messaging SMS Configuration Set (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *configurationSetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new configurationSetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := new.ConfigurationSetName.ValueString()
	if !new.DefaultSenderID.Equal(old.DefaultSenderID) {
		if new.DefaultSenderID.IsNull() {
			err := deleteDefaultSenderID(ctx, conn, name)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("deleting default sender ID for End User Messaging SMS Configuration Set (%s)", name), err.Error())

				return
			}
		} else {
			err := setDefaultSenderID(ctx, conn, name, new.DefaultSenderID.ValueString())

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("setting default sender ID for End User Messaging SMS Configuration Set (%s) to %s", name, new.DefaultSenderID.ValueString()), err.Error())

				return
			}
		}
	}
	if !new.DefaultMessageType.Equal(old.DefaultMessageType) {
		if new.DefaultMessageType.IsNull() {
			err := deleteDefaultMessageType(ctx, conn, name)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("deleting default message type for End User Messaging SMS Configuration Set (%s)", name), err.Error())

				return
			}
		} else {
			err := setDefaultMessageType(ctx, conn, name, new.DefaultMessageType.ValueEnum())

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("setting default message type for End User Messaging SMS Configuration Set (%s) to %s", name, new.DefaultMessageType.ValueString()), err.Error())

				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *configurationSetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data configurationSetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	_, err := conn.DeleteConfigurationSet(ctx, &pinpointsmsvoicev2.DeleteConfigurationSetInput{
		ConfigurationSetName: data.ID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting End User Messaging SMS Configuration Set (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type configurationSetResourceModel struct {
	ID                   types.String                             `tfsdk:"id"`
	ConfigurationSetARN  types.String                             `tfsdk:"arn"`
	ConfigurationSetName types.String                             `tfsdk:"name"`
	DefaultMessageType   fwtypes.StringEnum[awstypes.MessageType] `tfsdk:"default_message_type"`
	DefaultSenderID      types.String                             `tfsdk:"default_sender_id"`
	Tags                 tftags.Map                               `tfsdk:"tags"`
	TagsAll              tftags.Map                               `tfsdk:"tags_all"`
}

func (model *configurationSetResourceModel) InitFromID() error {
	model.ConfigurationSetName = model.ID

	return nil
}

func (model *configurationSetResourceModel) setID() {
	model.ID = model.ConfigurationSetName
}

func findConfigurationSetByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.ConfigurationSetInformation, error) {
	input := &pinpointsmsvoicev2.DescribeConfigurationSetsInput{
		ConfigurationSetNames: []string{id},
	}

	return findConfigurationSet(ctx, conn, input)
}

func findConfigurationSet(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeConfigurationSetsInput) (*awstypes.ConfigurationSetInformation, error) {
	output, err := findConfigurationSets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationSets(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeConfigurationSetsInput) ([]awstypes.ConfigurationSetInformation, error) {
	var output []awstypes.ConfigurationSetInformation

	pages := pinpointsmsvoicev2.NewDescribeConfigurationSetsPaginator(conn, input)
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

		output = append(output, page.ConfigurationSets...)
	}

	return output, nil
}

func setDefaultSenderID(ctx context.Context, conn *pinpointsmsvoicev2.Client, configurationSetName, senderID string) error {
	input := &pinpointsmsvoicev2.SetDefaultSenderIdInput{
		ConfigurationSetName: aws.String(configurationSetName),
		SenderId:             aws.String(senderID),
	}

	_, err := conn.SetDefaultSenderId(ctx, input)

	return err
}

func setDefaultMessageType(ctx context.Context, conn *pinpointsmsvoicev2.Client, configurationSetName string, messageType awstypes.MessageType) error {
	input := &pinpointsmsvoicev2.SetDefaultMessageTypeInput{
		ConfigurationSetName: aws.String(configurationSetName),
		MessageType:          messageType,
	}

	_, err := conn.SetDefaultMessageType(ctx, input)

	return err
}

func deleteDefaultSenderID(ctx context.Context, conn *pinpointsmsvoicev2.Client, configurationSetName string) error {
	input := &pinpointsmsvoicev2.DeleteDefaultSenderIdInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}

	_, err := conn.DeleteDefaultSenderId(ctx, input)

	return err
}

func deleteDefaultMessageType(ctx context.Context, conn *pinpointsmsvoicev2.Client, configurationSetName string) error {
	input := &pinpointsmsvoicev2.DeleteDefaultMessageTypeInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}

	_, err := conn.DeleteDefaultMessageType(ctx, input)

	return err
}
