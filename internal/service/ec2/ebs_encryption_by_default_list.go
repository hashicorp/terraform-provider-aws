// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ebs_encryption_by_default")
func newEBSEncryptionByDefaultResourceAsListResource() inttypes.ListResourceForSDK {
	l := ebsEncryptionByDefaultListResource{}
	l.SetResourceSchema(resourceEBSEncryptionByDefault())
	return &l
}

var _ list.ListResource = &ebsEncryptionByDefaultListResource{}

type ebsEncryptionByDefaultListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *ebsEncryptionByDefaultListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)
	awsClient := l.Meta()

	tflog.Info(ctx, "Listing EBS Encryption By Default")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ec2.GetEbsEncryptionByDefaultInput{}
		resp, err := conn.GetEbsEncryptionByDefault(ctx, input)
		if err != nil {
			yield(fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading EBS Encryption By Default: %w", err)))
			return
		}

		region := awsClient.Region(ctx)
		ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrRegion), region)

		result := request.NewListResult(ctx)
		rd := l.ResourceData()
		rd.SetId(region)
		rd.Set(names.AttrEnabled, resp.EbsEncryptionByDefault)

		if request.IncludeResource {
			if err := resourceEBSEncryptionByDefaultFlatten(ctx, resp, rd); err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("including EBS Encryption By Default: %w", err)))
				return
			}
		}

		state := "disabled"
		if aws.ToBool(resp.EbsEncryptionByDefault) {
			state = "enabled"
		}
		result.DisplayName = fmt.Sprintf("EBS Encryption By Default (%s, %s)", region, state)

		l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
		if result.Diagnostics.HasError() {
			yield(result)
			return
		}

		if !yield(result) {
			return
		}
	}
}
