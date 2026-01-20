// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_route53_record")
func newRecordResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceRecord{}
	l.SetResourceSchema(resourceRecord())
	return &l
}

var _ list.ListResource = &listResourceRecord{}

type listResourceRecord struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceRecord) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"zone_id": listschema.StringAttribute{
				Required:    true,
				Description: "The ID of the hosted zone to list records from",
			},
		},
	}
}

func (l *listResourceRecord) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().Route53Client(ctx)

	var query listRecordModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Route 53 Records")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := &route53.ListResourceRecordSetsInput{
			HostedZoneId: query.ZoneID.ValueStringPointer(),
		}
		for item, err := range listRecords(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			// Create a unique ID for this record
			recordID := createRecordIDFromResourceRecordSet(query.ZoneID.ValueString(), item)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), recordID)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(recordID)

			// Set identity attributes
			rd.Set("zone_id", query.ZoneID.ValueString())
			rd.Set(names.AttrName, aws.ToString(item.Name))
			rd.Set(names.AttrType, string(item.Type))
			if item.SetIdentifier != nil {
				rd.Set("set_identifier", aws.ToString(item.SetIdentifier))
			}

			tflog.Info(ctx, "Reading Route 53 Record")
			diags := resourceRecordRead(ctx, rd, l.Meta())
			if diags.HasError() {
				result.Diagnostics.Append(fwdiag.FromSDKDiagnostics(diags)...)
				yield(result)
				return
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

type listRecordModel struct {
	ZoneID types.String `tfsdk:"zone_id"`
}

func createRecordIDFromResourceRecordSet(zoneID string, rrs awstypes.ResourceRecordSet) string {
	parts := enum.Slice(
		zoneID,
		strings.ToLower(aws.ToString(rrs.Name)),
		string(rrs.Type),
	)
	if rrs.SetIdentifier != nil {
		parts = append(parts, aws.ToString(rrs.SetIdentifier))
	}
	return strings.Join(parts, "_")
}

func listRecords(ctx context.Context, conn *route53.Client, input *route53.ListResourceRecordSetsInput) iter.Seq2[awstypes.ResourceRecordSet, error] {
	return func(yield func(awstypes.ResourceRecordSet, error) bool) {
		pages := route53.NewListResourceRecordSetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ResourceRecordSet{}, fmt.Errorf("listing Route 53 Record resources: %w", err))
				return
			}

			for _, item := range page.ResourceRecordSets {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
