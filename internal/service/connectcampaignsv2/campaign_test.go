// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connectcampaignsv2

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connectcampaignsv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestCampaignResourceSchemaIncludesCoreAttributes(t *testing.T) {
	t.Parallel()

	res, err := newCampaignResource(context.Background())
	if err != nil {
		t.Fatalf("newCampaignResource returned error: %s", err)
	}

	var resp resource.SchemaResponse
	res.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %s", resp.Diagnostics.Errors())
	}

	for _, name := range []string{
		"arn",
		"connect_campaign_flow_arn",
		"connect_instance_id",
		"id",
		"name",
		"tags",
		"tags_all",
		"type",
	} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Fatalf("expected schema attribute %q", name)
		}
	}

	for _, name := range []string{
		"entry_limits_config",
		"schedule",
		"source",
	} {
		if _, ok := resp.Schema.Blocks[name]; !ok {
			t.Fatalf("expected schema block %q", name)
		}
	}
}

func TestCampaignResourceSchemaSourceIsUpdatable(t *testing.T) {
	t.Parallel()

	res, err := newCampaignResource(context.Background())
	if err != nil {
		t.Fatalf("newCampaignResource returned error: %s", err)
	}

	var resp resource.SchemaResponse
	res.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %s", resp.Diagnostics.Errors())
	}

	block, ok := resp.Schema.Blocks["source"].(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("source block type = %T, want schema.ListNestedBlock", resp.Schema.Blocks["source"])
	}
	if len(block.PlanModifiers) != 0 {
		t.Fatalf("source PlanModifiers length = %d, want 0", len(block.PlanModifiers))
	}
}

func TestServicePackageRegistersCampaignResource(t *testing.T) {
	t.Parallel()

	resources := (&servicePackage{}).FrameworkResources(context.Background())
	if len(resources) != 1 {
		t.Fatalf("expected 1 framework resource, got %d", len(resources))
	}
	if got, want := resources[0].TypeName, "aws_connectcampaignsv2_campaign"; got != want {
		t.Fatalf("resource type = %q, want %q", got, want)
	}
	if resources[0].Factory == nil {
		t.Fatal("resource factory is nil")
	}
}

func TestExpandEntryLimitsConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	got, err := expandEntryLimitsConfig(ctx, fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &entryLimitsConfigModel{
		MaxEntryCount:    types.Int64Value(2),
		MinEntryInterval: types.StringValue("PT1H"),
	}))
	if err != nil {
		t.Fatalf("expandEntryLimitsConfig returned error: %s", err)
	}
	if got == nil {
		t.Fatal("expanded value is nil")
	}
	if got.MaxEntryCount == nil || *got.MaxEntryCount != 2 {
		t.Fatalf("MaxEntryCount = %v, want 2", got.MaxEntryCount)
	}
	if got.MinEntryInterval == nil || *got.MinEntryInterval != "PT1H" {
		t.Fatalf("MinEntryInterval = %v, want PT1H", got.MinEntryInterval)
	}
}

func TestExpandEntryLimitsConfigAllowsRemoval(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	got, err := expandEntryLimitsConfig(ctx, fwtypes.NewListNestedObjectValueOfNull[entryLimitsConfigModel](ctx))
	if err != nil {
		t.Fatalf("expandEntryLimitsConfig returned error: %s", err)
	}
	if got != nil {
		t.Fatalf("expanded value = %#v, want nil", got)
	}
}

func TestExpandSchedule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	got, err := expandSchedule(ctx, fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &scheduleModel{
		EndTime:          types.StringValue("2026-06-18T11:00:00Z"),
		RefreshFrequency: types.StringValue("PT24H"),
		StartTime:        types.StringValue("2026-06-18T10:00:00Z"),
	}))
	if err != nil {
		t.Fatalf("expandSchedule returned error: %s", err)
	}
	if got == nil {
		t.Fatal("expanded value is nil")
	}
	if got.StartTime == nil || got.StartTime.Format("2006-01-02T15:04:05Z07:00") != "2026-06-18T10:00:00Z" {
		t.Fatalf("StartTime = %v, want 2026-06-18T10:00:00Z", got.StartTime)
	}
	if got.EndTime == nil || got.EndTime.Format("2006-01-02T15:04:05Z07:00") != "2026-06-18T11:00:00Z" {
		t.Fatalf("EndTime = %v, want 2026-06-18T11:00:00Z", got.EndTime)
	}
	if got.RefreshFrequency == nil || *got.RefreshFrequency != "PT24H" {
		t.Fatalf("RefreshFrequency = %v, want PT24H", got.RefreshFrequency)
	}
}

func TestExpandSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const segmentARN = "arn:aws:profile:us-east-1:123456789012:domains/domain/segments/segment" //lintignore:AWSAT003,AWSAT005
	got, err := expandSource(ctx, fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceModel{
		CustomerProfilesSegmentARN: types.StringValue(segmentARN),
		EventTrigger:               fwtypes.NewListNestedObjectValueOfNull[eventTriggerModel](ctx),
	}))
	if err != nil {
		t.Fatalf("expandSource returned error: %s", err)
	}
	segment, ok := got.(*awstypes.SourceMemberCustomerProfilesSegmentArn)
	if !ok {
		t.Fatalf("source type = %T, want *SourceMemberCustomerProfilesSegmentArn", got)
	}
	if segment.Value != segmentARN {
		t.Fatalf("segment ARN = %q", segment.Value)
	}
}

func TestExpandSourceEventTrigger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const domainARN = "arn:aws:profile:us-east-1:123456789012:domains/domain" //lintignore:AWSAT003,AWSAT005
	got, err := expandSource(ctx, fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceModel{
		CustomerProfilesSegmentARN: types.StringNull(),
		EventTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &eventTriggerModel{
			CustomerProfilesDomainARN: types.StringValue(domainARN),
		}),
	}))
	if err != nil {
		t.Fatalf("expandSource returned error: %s", err)
	}
	trigger, ok := got.(*awstypes.SourceMemberEventTrigger)
	if !ok {
		t.Fatalf("source type = %T, want *SourceMemberEventTrigger", got)
	}
	if trigger.Value.CustomerProfilesDomainArn == nil || *trigger.Value.CustomerProfilesDomainArn != domainARN {
		t.Fatalf("domain ARN = %v", trigger.Value.CustomerProfilesDomainArn)
	}
}

func TestExpandSourceRejectsMultipleUnionMembers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := expandSource(ctx, fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceModel{
		CustomerProfilesSegmentARN: types.StringValue("arn"),
		EventTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &eventTriggerModel{
			CustomerProfilesDomainARN: types.StringValue("arn"),
		}),
	}))
	if err == nil {
		t.Fatal("expected error")
	}
}
