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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_key_pair")
func newKeyPairResourceAsListResource() inttypes.ListResourceForSDK {
	l := keyPairListResource{}
	l.SetResourceSchema(resourceKeyPair())
	return &l
}

var _ list.ListResource = &keyPairListResource{}

type keyPairListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *keyPairListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listKeyPairs(ctx, conn, &ec2.DescribeKeyPairsInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			keyName := aws.ToString(item.KeyName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("key_name"), keyName)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(keyName)
			rd.Set("key_name", keyName)

			if request.IncludeResource {
				resourceKeyPairFlatten(ctx, l.Meta(), &item, rd)
			}

			result.DisplayName = keyName

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

func listKeyPairs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeKeyPairsInput) iter.Seq2[awstypes.KeyPairInfo, error] {
	return func(yield func(awstypes.KeyPairInfo, error) bool) {
		items, err := findKeyPairs(ctx, conn, input)
		if err != nil {
			yield(awstypes.KeyPairInfo{}, fmt.Errorf("listing EC2 Key Pair resources: %w", err))
			return
		}

		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}
