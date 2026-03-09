// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_secretsmanager_secret")
func newSecretResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceSecret{}
	l.SetResourceSchema(resourceSecret())
	return &l
}

var _ list.ListResource = &listResourceSecret{}

type listResourceSecret struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceSecret) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecretsManagerClient(ctx)

	var query listSecretModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Secrets Manager Secret")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input secretsmanager.ListSecretsInput
		for item, err := range listSecrets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ARN)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)

			tflog.Info(ctx, "Reading Secrets Manager Secret")
			diags := resourceSecretRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading Secrets Manager Secret", map[string]any{
					names.AttrID: arn,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listSecretModel struct {
	framework.WithRegionModel
}

func listSecrets(ctx context.Context, conn *secretsmanager.Client, input *secretsmanager.ListSecretsInput) iter.Seq2[awstypes.SecretListEntry, error] {
	return func(yield func(awstypes.SecretListEntry, error) bool) {
		pages := secretsmanager.NewListSecretsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecretListEntry{}, fmt.Errorf("listing Secrets Manager Secret resources: %w", err))
				return
			}

			for _, item := range page.SecretList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
