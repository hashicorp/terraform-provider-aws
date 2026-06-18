// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// Custom Route 53 service lister functions using the same format as generated code.

func listTrafficPolicyInstancesPages(ctx context.Context, conn *route53.Client, input *route53.ListTrafficPolicyInstancesInput, fn func(*route53.ListTrafficPolicyInstancesOutput, bool) bool, optFns ...func(*route53.Options)) error {
	for {
		output, err := conn.ListTrafficPolicyInstances(ctx, input, optFns...)
		if err != nil {
			return err
		}

		lastPage := !output.IsTruncated
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.HostedZoneIdMarker = output.HostedZoneIdMarker
		input.TrafficPolicyInstanceNameMarker = output.TrafficPolicyInstanceNameMarker
		input.TrafficPolicyInstanceTypeMarker = output.TrafficPolicyInstanceTypeMarker
	}
	return nil
}
