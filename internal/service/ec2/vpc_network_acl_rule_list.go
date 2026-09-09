// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_network_acl_rule")
func newNetworkACLRuleResourceAsListResource() inttypes.ListResourceForSDK {
	l := networkACLRuleListResource{}
	l.SetResourceSchema(resourceNetworkACLRule())

	return &l
}

var _ list.ListResource = &networkACLRuleListResource{}

type networkACLRuleListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *networkACLRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		nacls, err := findNetworkACLs(ctx, conn, &ec2.DescribeNetworkAclsInput{})
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing EC2 Network ACL Rules: %w", err))
			yield(result)
			return
		}

		for _, nacl := range nacls {
			naclID := aws.ToString(nacl.NetworkAclId)

			for _, entry := range nacl.Entries {
				egress := aws.ToBool(entry.Egress)
				ruleNumber := int(aws.ToInt32(entry.RuleNumber))

				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("network_acl_id"), naclID)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("rule_number"), ruleNumber)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("egress"), egress)

				protocolNumber, err := networkACLProtocolNumber(aws.ToString(entry.Protocol))
				if err != nil {
					tflog.Error(ctx, "Reading EC2 Network ACL Rule", map[string]any{
						"error": err.Error(),
					})
					continue
				}
				protocol := strconv.Itoa(protocolNumber)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.Set("network_acl_id", naclID)
				rd.Set("egress", egress)
				rd.Set("rule_number", ruleNumber)
				rd.Set(names.AttrProtocol, protocol)

				id := networkACLRuleCreateResourceID(naclID, ruleNumber, egress, protocol)
				rd.SetId(id)

				if request.IncludeResource {
					if diags := resourceNetworkACLRuleFlatten(rd, &entry); diags.HasError() {
						tflog.Error(ctx, "Reading EC2 Network ACL Rule", map[string]any{
							"diags": sdkdiag.DiagnosticsString(diags),
						})
						continue
					}
				}

				result.DisplayName = id

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
}
