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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_secretsmanager_secret_policy")
func newSecretPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := secretPolicyListResource{}
	l.SetResourceSchema(resourceSecretPolicy())
	return &l
}

var _ list.ListResource = &secretPolicyListResource{}

type secretPolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *secretPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecretsManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input secretsmanager.ListSecretsInput
		for item, err := range listSecretPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ARN)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set("secret_arn", arn)

			if request.IncludeResource {
				rd.Set(names.AttrPolicy, aws.ToString(item.ResourcePolicy))
			}

			result.DisplayName = aws.ToString(item.Name)

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

func listSecretPolicies(ctx context.Context, conn *secretsmanager.Client, input *secretsmanager.ListSecretsInput) iter.Seq2[*secretsmanager.GetResourcePolicyOutput, error] {
	return func(yield func(*secretsmanager.GetResourcePolicyOutput, error) bool) {
		pages := secretsmanager.NewListSecretsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(&secretsmanager.GetResourcePolicyOutput{}, fmt.Errorf("listing Secrets Manager Secret Policy resources: %w", err))
				return
			}

			for _, item := range page.SecretList {
				arn := aws.ToString(item.ARN)
				output, err := findSecretPolicyByID(ctx, conn, arn)

				if err != nil {
					// Skip deleted secrets or those without policies
					if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
						errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "You can't perform this operation on the secret because it was marked for deletion") ||
						retry.NotFound(err) ||
						output.ResourcePolicy == nil ||
						aws.ToString(output.ResourcePolicy) == "" {
						continue
					}

					yield(&secretsmanager.GetResourcePolicyOutput{}, fmt.Errorf("getting resource policy for secret %s: %w", arn, err))
					return
				}

				if !yield(output, nil) {
					return
				}
			}
		}
	}
}
