// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressEquivalentBusNameOrARN provides custom difference suppression
// for event_bus_name values that may be specified as either a bus name
// or its equivalent ARN.
func suppressEquivalentBusNameOrARN(_, old, new string, _ *schema.ResourceData) bool {
	return busNameFromNameOrARN(old) == busNameFromNameOrARN(new)
}

// busNameFromNameOrARN extracts the event bus name from either a plain
// name string or a full event bus ARN.
func busNameFromNameOrARN(nameOrARN string) string {
	parsed, err := arn.Parse(nameOrARN)
	if err != nil {
		return nameOrARN
	}
	// ARN resource format: "event-bus/{bus-name}"
	_, name, ok := strings.Cut(parsed.Resource, "/")
	if ok && name != "" {
		return name
	}
	return nameOrARN
}
