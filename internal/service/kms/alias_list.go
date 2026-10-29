// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_kms_alias")
func aliasResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceAlias(), framework.SDKListSpec[awstypes.AliasListEntry, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.AliasListEntry, error], diag.Diagnostics) {
			var input kms.ListAliasesInput

			return listAliases(ctx, c.KMSClient(ctx), &input), nil
		},
		ID: func(a awstypes.AliasListEntry, _ framework.RegionQuery) string {
			return aws.ToString(a.AliasName)
		},
		Flatten: resourceAliasFlatten,
	})
}

func listAliases(ctx context.Context, conn *kms.Client, input *kms.ListAliasesInput) iter.Seq2[awstypes.AliasListEntry, error] {
	return tfiter.PageItems[kms.Options](ctx, "listing KMS Aliases resources", kms.NewListAliasesPaginator(conn, input),
		func(o *kms.ListAliasesOutput) []awstypes.AliasListEntry { return o.Aliases })
}
