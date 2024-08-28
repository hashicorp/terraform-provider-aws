// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

type trafficPolicyDocEndpointType string

const (
	trafficPolicyDocEndpointALB              trafficPolicyDocEndpointType = "application-load-balancer"
	trafficPolicyDocEndpointCloudFront       trafficPolicyDocEndpointType = "cloudfront"
	trafficPolicyDocEndpointElasticBeanstalk trafficPolicyDocEndpointType = "elastic-beanstalk"
	trafficPolicyDocEndpointELB              trafficPolicyDocEndpointType = "elastic-load-balancer"
	trafficPolicyDocEndpointNLB              trafficPolicyDocEndpointType = "network-load-balancer"
	trafficPolicyDocEndpointS3Website        trafficPolicyDocEndpointType = "s3-website"
	trafficPolicyDocEndpointValue            trafficPolicyDocEndpointType = "value" // nosemgrep:ci.literal-value-string-constant
)

func (trafficPolicyDocEndpointType) Values() []trafficPolicyDocEndpointType {
	return []trafficPolicyDocEndpointType{
		trafficPolicyDocEndpointALB,
		trafficPolicyDocEndpointCloudFront,
		trafficPolicyDocEndpointElasticBeanstalk,
		trafficPolicyDocEndpointELB,
		trafficPolicyDocEndpointNLB,
		trafficPolicyDocEndpointS3Website,
		trafficPolicyDocEndpointValue, // nosemgrep:ci.literal-value-string-constant
	}
}

type route53TrafficPolicyDoc struct {
	AWSPolicyFormatVersion string                            `json:",omitempty"`
	RecordType             string                            `json:",omitempty"`
	StartEndpoint          string                            `json:",omitempty"`
	StartRule              string                            `json:",omitempty"`
	Endpoints              map[string]*trafficPolicyEndpoint `json:",omitempty"`
	Rules                  map[string]*trafficPolicyRule     `json:",omitempty"`
}

type trafficPolicyEndpoint struct {
	Type   string `json:",omitempty"`
	Region string `json:",omitempty"`
	Value  string `json:",omitempty"`
}

type trafficPolicyRule struct {
	RuleType              string                               `json:",omitempty"`
	Primary               *trafficPolicyFailoverRule           `json:",omitempty"`
	Secondary             *trafficPolicyFailoverRule           `json:",omitempty"`
	Locations             []*trafficPolicyGeolocationRule      `json:",omitempty"`
	GeoProximityLocations []*trafficPolicyGeoproximityRule     `json:"GeoproximityLocations,omitempty"`
	Regions               []*trafficPolicyLatencyRule          `json:",omitempty"`
	Items                 []*trafficPolicyMultiValueAnswerRule `json:",omitempty"`
}

type trafficPolicyFailoverRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type trafficPolicyGeolocationRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	IsDefault            *bool  `json:",omitempty"`
	Continent            string `json:",omitempty"`
	Country              string `json:",omitempty"`
	Subdivision          string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type trafficPolicyGeoproximityRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	Region               string `json:",omitempty"`
	Latitude             string `json:",omitempty"`
	Longitude            string `json:",omitempty"`
	Bias                 string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type trafficPolicyLatencyRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	Region               string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type trafficPolicyMultiValueAnswerRule struct {
	EndpointReference string `json:",omitempty"`
	HealthCheck       string `json:",omitempty"`
}
