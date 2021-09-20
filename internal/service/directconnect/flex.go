package dx

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandRouteFilterPrefixes(vPrefixes *schema.Set) []*directconnect.RouteFilterPrefix {
	routeFilterPrefixes := []*directconnect.RouteFilterPrefix{}

	for _, vPrefix := range vPrefixes.List() {
		routeFilterPrefixes = append(routeFilterPrefixes, &directconnect.RouteFilterPrefix{
			Cidr: aws.String(vPrefix.(string)),
		})
	}

	return routeFilterPrefixes
}

func flattenRouteFilterPrefixes(routeFilterPrefixes []*directconnect.RouteFilterPrefix) *schema.Set {
	vPrefixes := []interface{}{}

	for _, routeFilterPrefix := range routeFilterPrefixes {
		vPrefixes = append(vPrefixes, aws.StringValue(routeFilterPrefix.Cidr))
	}

	return schema.NewSet(schema.HashString, vPrefixes)
}
