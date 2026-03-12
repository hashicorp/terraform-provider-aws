// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package transfer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfstringplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/privatestate"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_transfer_host_key", name="Host Key")
// @Tags(identifierAttribute="arn")
func newHostKeyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hostKeyResource{}

	return r, nil
}

type hostKeyResource struct {
	framework.ResourceWithModel[hostKeyResourceModel]
}

const (
	hostKeyBodyWOKey = "host_key_body_wo"
)

func (r *hostKeyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 200),
				},
			},
			"host_key_body": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 4096),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("host_key_body"),
						path.MatchRoot("host_key_body_wo"),
					),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("host_key_body_wo")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_key_body_wo": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 4096),
				},
				PlanModifiers: []planmodifier.String{
					tfstringplanmodifier.RequiresReplaceWO(hostKeyBodyWOKey),
				},
			},
			"host_key_fingerprint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_key_id": framework.IDAttribute(),
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *hostKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config hostKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	serverID := fwflex.StringValueFromFramework(ctx, plan.ServerID)
	var input transfer.ImportHostKeyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Prefer write-only value. It's only in Config, not Plan.
	if !config.HostKeyBodyWO.IsNull() {
		input.HostKeyBody = fwflex.StringFromFramework(ctx, config.HostKeyBodyWO)
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	out, err := conn.ImportHostKey(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Transfer Host Key (%s)", serverID), err.Error())

		return
	}

	// Store hash of write-only value.
	if !config.HostKeyBodyWO.IsNull() {
		woStore := privatestate.NewWriteOnlyValueStore(response.Private, hostKeyBodyWOKey)
		response.Diagnostics.Append(woStore.SetValue(ctx, config.HostKeyBodyWO)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	hostKeyID := aws.ToString(out.HostKeyId)
	hostKey, err := findHostKeyByTwoPartKey(ctx, conn, serverID, hostKeyID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Host Key (%s)", hostKeyID), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, hostKey, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *hostKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data hostKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	serverID, hostKeyID := fwflex.StringValueFromFramework(ctx, data.ServerID), fwflex.StringValueFromFramework(ctx, data.HostKeyID)
	out, err := findHostKeyByTwoPartKey(ctx, conn, serverID, hostKeyID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Host Key (%s)", hostKeyID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *hostKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old hostKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		hostKeyID := fwflex.StringValueFromFramework(ctx, new.HostKeyID)
		var input transfer.UpdateHostKeyInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateHostKey(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Transfer Host Key (%s)", hostKeyID), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *hostKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data hostKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	serverID, hostKeyID := fwflex.StringValueFromFramework(ctx, data.ServerID), fwflex.StringValueFromFramework(ctx, data.HostKeyID)
	input := transfer.DeleteHostKeyInput{
		HostKeyId: aws.String(hostKeyID),
		ServerId:  aws.String(serverID),
	}
	_, err := conn.DeleteHostKey(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Transfer Host Key (%s)", hostKeyID), err.Error())

		return
	}
}

func (r *hostKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		hostKeyIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, hostKeyIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("server_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("host_key_id"), parts[1])...)
}

func findHostKeyByTwoPartKey(ctx context.Context, conn *transfer.Client, serverID, hostKeyID string) (*awstypes.DescribedHostKey, error) {
	input := transfer.DescribeHostKeyInput{
		HostKeyId: aws.String(hostKeyID),
		ServerId:  aws.String(serverID),
	}

	return findHostKey(ctx, conn, &input)
}

func findHostKey(ctx context.Context, conn *transfer.Client, input *transfer.DescribeHostKeyInput) (*awstypes.DescribedHostKey, error) {
	output, err := conn.DescribeHostKey(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HostKey == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.HostKey, nil
}

type hostKeyResourceModel struct {
	framework.WithRegionModel
	ARN                types.String `tfsdk:"arn"`
	Description        types.String `tfsdk:"description"`
	HostKeyBody        types.String `tfsdk:"host_key_body"`
	HostKeyBodyWO      types.String `tfsdk:"host_key_body_wo"`
	HostKeyFingerprint types.String `tfsdk:"host_key_fingerprint"`
	HostKeyID          types.String `tfsdk:"host_key_id"`
	ServerID           types.String `tfsdk:"server_id"`
	Tags               tftags.Map   `tfsdk:"tags"`
	TagsAll            tftags.Map   `tfsdk:"tags_all"`
}
