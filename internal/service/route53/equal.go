// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-provider-aws/internal/dns"
)

// resourceRecordSetEqual determines whether two record sets are semantically equal
func resourceRecordSetEqual(s1, s2 awstypes.ResourceRecordSet) bool {
	return resourceRecordSetIdentifiersEqual(s1, s2) &&
		// Root level attributes
		(s1.Failover == "" && s2.Failover == "" || s1.Failover == s2.Failover) &&
		aws.ToString(s1.HealthCheckId) == aws.ToString(s2.HealthCheckId) &&
		aws.ToBool(s1.MultiValueAnswer) == aws.ToBool(s2.MultiValueAnswer) &&
		(s1.Region == "" && s2.Region == "" || s1.Region == s2.Region) &&
		aws.ToString(s1.TrafficPolicyInstanceId) == aws.ToString(s2.TrafficPolicyInstanceId) &&
		aws.ToInt64(s1.TTL) == aws.ToInt64(s2.TTL) &&
		aws.ToInt64(s1.Weight) == aws.ToInt64(s2.Weight) &&
		// Nested attributes
		aliasTargetEqual(s1.AliasTarget, s2.AliasTarget) &&
		cidrRoutingConfigEqual(s1.CidrRoutingConfig, s2.CidrRoutingConfig) &&
		geolocationEqual(s1.GeoLocation, s2.GeoLocation) &&
		geoProximityLocationEqual(s1.GeoProximityLocation, s2.GeoProximityLocation) &&
		resourceRecordsEqual(s1.ResourceRecords, s2.ResourceRecords)
}

// resourceRecordSetIdentifiersEqual determines whether the identifier fields
// of two records are equal
//
// This function can be to determine where record identifiers match, but the entirety
// of the record is not equal. In these scenarios the record can use an Upsert change
// in the batch, rather than one Delete and one Create change.
//
// From the AWS documentation:
//
// > In a group of resource record sets that have the same name and type, the value
// > of SetIdentifier must be unique for each resource record set.
//
// Ref: https://docs.aws.amazon.com/Route53/latest/APIReference/API_ResourceRecordSet.html#Route53-Type-ResourceRecordSet-SetIdentifier
func resourceRecordSetIdentifiersEqual(s1, s2 awstypes.ResourceRecordSet) bool {
	return dns.Normalize(aws.ToString(s1.Name)) == dns.Normalize(aws.ToString(s2.Name)) &&
		s1.Type == s2.Type &&
		aws.ToString(s1.SetIdentifier) == aws.ToString(s2.SetIdentifier)
}

func aliasTargetEqual(s1, s2 *awstypes.AliasTarget) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil && s2 != nil || s1 != nil && s2 == nil {
		return false
	}

	return dns.Normalize(aws.ToString(s1.DNSName)) == dns.Normalize(aws.ToString(s2.DNSName)) &&
		s1.EvaluateTargetHealth == s2.EvaluateTargetHealth &&
		aws.ToString(s1.HostedZoneId) == aws.ToString(s2.HostedZoneId)
}

func cidrRoutingConfigEqual(s1, s2 *awstypes.CidrRoutingConfig) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil && s2 != nil || s1 != nil && s2 == nil {
		return false
	}

	return aws.ToString(s1.CollectionId) == aws.ToString(s2.CollectionId) &&
		aws.ToString(s1.LocationName) == aws.ToString(s2.LocationName)
}

func geolocationEqual(s1, s2 *awstypes.GeoLocation) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil && s2 != nil || s1 != nil && s2 == nil {
		return false
	}

	return aws.ToString(s1.ContinentCode) == aws.ToString(s2.ContinentCode) &&
		aws.ToString(s1.CountryCode) == aws.ToString(s2.CountryCode) &&
		aws.ToString(s1.SubdivisionCode) == aws.ToString(s2.SubdivisionCode)
}

func geoProximityLocationEqual(s1, s2 *awstypes.GeoProximityLocation) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil && s2 != nil || s1 != nil && s2 == nil {
		return false
	}

	return aws.ToString(s1.AWSRegion) == aws.ToString(s2.AWSRegion) &&
		aws.ToInt32(s1.Bias) == aws.ToInt32(s2.Bias) &&
		coordinatesEqual(s1.Coordinates, s2.Coordinates) &&
		aws.ToString(s1.LocalZoneGroup) == aws.ToString(s2.LocalZoneGroup)
}

func coordinatesEqual(s1, s2 *awstypes.Coordinates) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil && s2 != nil || s1 != nil && s2 == nil {
		return false
	}

	return aws.ToString(s1.Latitude) == aws.ToString(s2.Latitude) &&
		aws.ToString(s1.Longitude) == aws.ToString(s2.Longitude)
}

func resourceRecordsEqual(s1, s2 []awstypes.ResourceRecord) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if aws.ToString(s1[i].Value) != aws.ToString(s2[i].Value) {
			return false
		}
	}

	return true
}
