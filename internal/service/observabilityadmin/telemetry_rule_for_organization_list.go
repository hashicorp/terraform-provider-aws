// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_observabilityadmin_telemetry_rule_for_organization")
func newTelemetryRuleForOrganizationResourceAsListResource() list.ListResourceWithConfigure {
	return &telemetryRuleForOrganizationListResource{}
}

var _ list.ListResource = &telemetryRuleForOrganizationListResource{}

type telemetryRuleForOrganizationListResource struct {
	telemetryRuleForOrganizationResource
	framework.WithList
}

func (l *telemetryRuleForOrganizationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ObservabilityAdminClient(ctx)

	var query listTelemetryRuleForOrganizationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input observabilityadmin.ListTelemetryRulesForOrganizationInput
		for item, err := range listTelemetryRulesForOrganization(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.RuleArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			ruleName := aws.ToString(item.RuleName)
			output, err := findTelemetryRuleForOrganization(ctx, conn, ruleName)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data telemetryRuleForOrganizationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = ruleName
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTelemetryRuleForOrganizationModel struct {
	framework.WithRegionModel
}

func listTelemetryRulesForOrganization(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.ListTelemetryRulesForOrganizationInput) iter.Seq2[awstypes.TelemetryRuleSummary, error] {
	return func(yield func(awstypes.TelemetryRuleSummary, error) bool) {
		pages := observabilityadmin.NewListTelemetryRulesForOrganizationPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.TelemetryRuleSummary](), fmt.Errorf("listing CloudWatch Observability Admin Telemetry Rules For Organization: %w", err))
				return
			}

			for _, item := range page.TelemetryRuleSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
