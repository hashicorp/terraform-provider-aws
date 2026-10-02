// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCapacityReservationDataSourceInput(t *testing.T) {
	t.Parallel()

	input := capacityReservationDataSourceInput(context.Background(), types.StringValue("cr-1234567890abcdef0"), []awstypes.Filter{
		{
			Name:   aws.String("reservation-type"),
			Values: []string{"capacity-block"},
		},
		{
			Name:   aws.String("instance-type"),
			Values: []string{"t3.micro"},
		},
	})

	expected := []awstypes.Filter{
		{
			Name:   aws.String("instance-type"),
			Values: []string{"t3.micro"},
		},
		{
			Name:   aws.String("reservation-type"),
			Values: []string{"default"},
		},
	}

	if diff := cmp.Diff(input.CapacityReservationIds, []string{"cr-1234567890abcdef0"}); diff != "" {
		t.Fatalf("unexpected capacity reservation IDs (+wanted, -got): %s", diff)
	}

	if diff := cmp.Diff(input.Filters, expected, cmp.AllowUnexported(awstypes.Filter{})); diff != "" {
		t.Errorf("unexpected filters (+wanted, -got): %s", diff)
	}
}

func TestCapacityReservationDataSourceConfigValidators(t *testing.T) {
	t.Parallel()

	validator := (&capacityReservationDataSource{}).ConfigValidators(context.Background())[0]
	response := &datasource.ValidateConfigResponse{}

	validator.ValidateDataSource(context.Background(), datasource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					names.AttrID:     schema.StringAttribute{Optional: true},
					names.AttrFilter: schema.StringAttribute{Optional: true},
				},
			},
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					names.AttrID:     tftypes.String,
					names.AttrFilter: tftypes.String,
				},
			}, map[string]tftypes.Value{
				names.AttrID:     tftypes.NewValue(tftypes.String, nil),
				names.AttrFilter: tftypes.NewValue(tftypes.String, nil),
			}),
		},
	}, response)

	if !response.Diagnostics.HasError() {
		t.Error("expected an error for a data source configuration without id or filter")
	}
}
