// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_iam_access_key")
func newAccessKeyResourceAsListResource() inttypes.ListResourceForSDK {
	l := accessKeyListResource{}
	l.SetResourceSchema(resourceAccessKey())
	return &l
}

var _ list.ListResource = &accessKeyListResource{}

type accessKeyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *accessKeyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for user, err := range listUsers(ctx, conn, &iam.ListUsersInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			username := aws.ToString(user.UserName)
			accessKeys, err := findAccessKeysByUser(ctx, conn, username)
			if retry.NotFound(err) {
				tflog.Warn(ctx, "User disappeared during listing, skipping", map[string]any{
					logging.ResourceAttributeKey("user"): username,
				})
				continue
			}
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, accessKey := range accessKeys {
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("user"), username)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), aws.ToString(accessKey.AccessKeyId))

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(aws.ToString(accessKey.AccessKeyId))

				if request.IncludeResource {
					resourceAccessKeyFlatten(rd, &accessKey)
				}

				result.DisplayName = fmt.Sprintf("User: %s - Access Key: %s", username, aws.ToString(accessKey.AccessKeyId))

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
	}
}
