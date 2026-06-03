// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_volume_attachment", name="EBS Volume Attachment")
func newEBSVolumeAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := ebsVolumeAttachmentListResource{}
	l.SetResourceSchema(resourceVolumeAttachment())
	return &l
}

var _ list.ListResource = &ebsVolumeAttachmentListResource{}

type ebsVolumeAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *ebsVolumeAttachmentListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrInstanceID: listschema.StringAttribute{
				Required:    true,
				Description: "ID of the EC2 Instance to list Volume Attachments for.",
			},
		},
	}
}

func (l *ebsVolumeAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listEBSVolumeAttachmentModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	instanceID := query.InstanceID.ValueString()

	tflog.Info(ctx, "Listing resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrInstanceID): instanceID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		for attachment, err := range listEBSVolumeAttachments(ctx, conn, instanceID) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			deviceName := aws.ToString(attachment.Device)
			instanceID := aws.ToString(attachment.InstanceId)
			volumeID := aws.ToString(attachment.VolumeId)

			if deviceName == "" || instanceID == "" || volumeID == "" {
				continue
			}

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
			rd.Set(names.AttrDeviceName, deviceName)
			rd.Set("volume_id", volumeID)
			rd.Set(names.AttrInstanceID, instanceID)

			if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling.
				// No-op, all readable attributes are already populated above.
			}

			result.DisplayName = fmt.Sprintf("%s (%s - %s)", instanceID, deviceName, volumeID)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listEBSVolumeAttachmentModel struct {
	framework.WithRegionModel
	InstanceID types.String `tfsdk:"instance_id"`
}

func listEBSVolumeAttachments(ctx context.Context, conn *ec2.Client, instanceID string) iter.Seq2[awstypes.VolumeAttachment, error] {
	return func(yield func(awstypes.VolumeAttachment, error) bool) {
		attachments, err := findVolumeAttachments(ctx, conn, instanceID)
		if err != nil {
			yield(awstypes.VolumeAttachment{}, fmt.Errorf("listing EC2 EBS Volume Attachment resources for instance (%s): %w", instanceID, err))
			return
		}

		for _, attachment := range attachments {
			if !yield(attachment, nil) {
				return
			}
		}
	}
}
