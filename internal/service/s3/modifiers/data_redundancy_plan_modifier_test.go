package modifiers

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestDataRedundancyPlanModifier_Description(t *testing.T) {
	// Test Description method
	modifier := dataRedundancyPlanModifier{}
	desc := modifier.Description(context.Background())
	assert.Equal(t, "Sets the default value for data_redundancy based on the location type.", desc)
}

func TestDataRedundancyPlanModifier_MarkdownDescription(t *testing.T) {
	// Test MarkdownDescription method
	modifier := dataRedundancyPlanModifier{}
	mdDesc := modifier.MarkdownDescription(context.Background())
	assert.Equal(t, "Sets the default value for `data_redundancy` based on the `location` type.", mdDesc)
}

func TestDataRedundancyPlanModifier_PlanModifyString(t *testing.T) {
	tests := []struct {
		name           string
		configValue    types.String
		location       []LocationInfo
		expectedResult string
	}{
		{
			name:           "When location is LocalZone",
			configValue:    types.StringValue(""),
			location:       []LocationInfo{{Type: "local-zone"}},
			expectedResult: string(types.DataRedundancySingleLocalZone),
		},
		{
			name:           "When location is AvailabilityZone",
			configValue:    types.StringValue(""),
			location:       []LocationInfo{{Type: "availability-zone"}},
			expectedResult: string(types.DataRedundancySingleAvailabilityZone),
		},
		{
			name:           "When location is empty",
			configValue:    types.StringValue(""),
			location:       nil,
			expectedResult: string(types.DataRedundancySingleAvailabilityZone), // default if empty location
		},
		{
			name:           "When config value is already set",
			configValue:    types.StringValue("existing_value"),
			location:       nil,
			expectedResult: "existing_value", // should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the request and response
			req := planmodifier.StringRequest{
				ConfigValue: tt.configValue,
				Config:      resource.TestConfigSchema{}, // You may need to replace this with actual config schema or mock object
			}
			resp := &planmodifier.StringResponse{}

			// Setting location in config attribute
			mockLocationList := createMockLocationList(tt.location)
			req.Config.SetAttribute(context.Background(), "location", mockLocationList)

			// Call PlanModifyString
			modifier := dataRedundancyPlanModifier{}
			modifier.PlanModifyString(context.Background(), req, resp)

			// Assert the result
			assert.Equal(t, tt.expectedResult, resp.PlanValue.Value)
		})
	}
}

// Helper function to create a mock location list (simplified for this example)
func createMockLocationList(locations []LocationInfo) fwtypes.ListNestedObjectValueOf[models.LocationInfoModel] {
	var locationList fwtypes.ListNestedObjectValueOf[models.LocationInfoModel]
	for _, loc := range locations {
		locationList.Elements = append(locationList.Elements, models.LocationInfoModel{
			Type: loc.Type,
		})
	}
	return locationList
}

// Mock LocationInfo structure to mimic location info behavior
type LocationInfo struct {
	Type string
}
