// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_iam_role")
func newRoleResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceRole(), framework.SDKListSpec[awstypes.Role, struct{}]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ struct{}) (iter.Seq2[awstypes.Role, error], diag.Diagnostics) {
			var input iam.ListRolesInput

			return listNonServiceLinkedRoles(ctx, c.IAMClient(ctx), &input), nil
		},
		ID: func(r awstypes.Role, _ struct{}) string {
			return aws.ToString(r.RoleName)
		},
		Flatten:     resourceRoleFlatten,
		DisplayName: resourceRoleDisplayName,
	})
}

func resourceRoleDisplayName(role awstypes.Role) string {
	var buf strings.Builder

	path := aws.ToString(role.Path)
	buf.WriteString(strings.TrimPrefix(path, "/"))

	buf.WriteString(aws.ToString(role.RoleName))

	return buf.String()
}
