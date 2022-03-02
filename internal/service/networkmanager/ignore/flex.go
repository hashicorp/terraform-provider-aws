package networkmanager

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenLocation(location *networkmanager.Location) []interface{} {
	if location == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"address":   aws.StringValue(location.Address),
		"latitude":  aws.StringValue(location.Latitude),
		"longitude": aws.StringValue(location.Longitude),
	}
	return []interface{}{m}
}

func expandLocation(l []interface{}) *networkmanager.Location {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})
	logs := &networkmanager.Location{
		Address:   aws.String(m["address"].(string)),
		Latitude:  aws.String(m["latitude"].(string)),
		Longitude: aws.String(m["longitude"].(string)),
	}
	return logs
}
