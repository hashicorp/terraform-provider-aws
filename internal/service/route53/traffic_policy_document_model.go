package route53

const (
	trafficPolicyDocEndpointValue      = "value"
	trafficPolicyDocEndpointCloudFront = "cloudfront"
	trafficPolicyDocEndpointElastic    = "elastic-load-balancer"
	trafficPolicyDocEndpointS3         = "s3-website"
)

// TrafficPolicyDocEndpointType_values returns all elements of the endpoints types
func TrafficPolicyDocEndpointType_values() []string {
	return []string{
		trafficPolicyDocEndpointValue,
		trafficPolicyDocEndpointCloudFront,
		trafficPolicyDocEndpointElastic,
		trafficPolicyDocEndpointS3,
	}
}

type Route53TrafficPolicyDoc struct {
	AWSPolicyFormatVersion string                            `json:",omitempty"`
	RecordType             string                            `json:",omitempty"`
	StartEndpoint          string                            `json:",omitempty"`
	StartRule              string                            `json:",omitempty"`
	Endpoints              map[string]*TrafficPolicyEndpoint `json:",omitempty"`
	Rules                  map[string]*TrafficPolicyRule     `json:",omitempty"`
}

type TrafficPolicyEndpoint struct {
	Type   string `json:",omitempty"`
	Region string `json:",omitempty"`
	Value  string `json:",omitempty"`
}

type TrafficPolicyRule struct {
	RuleType              string                               `json:",omitempty"`
	Primary               *TrafficPolicyFailoverRule           `json:",omitempty"`
	Secondary             *TrafficPolicyFailoverRule           `json:",omitempty"`
	Locations             []*TrafficPolicyGeolocationRule      `json:",omitempty"`
	GeoProximityLocations []*TrafficPolicyGeoproximityRule     `json:",omitempty"`
	Regions               []*TrafficPolicyLatencyRule          `json:",omitempty"`
	Items                 []*TrafficPolicyMultiValueAnswerRule `json:",omitempty"`
}

type TrafficPolicyFailoverRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type TrafficPolicyGeolocationRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	IsDefault            *bool  `json:",omitempty"`
	Continent            string `json:",omitempty"`
	Country              string `json:",omitempty"`
	Subdivision          string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type TrafficPolicyGeoproximityRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	Region               string `json:",omitempty"`
	Latitude             string `json:",omitempty"`
	Longitude            string `json:",omitempty"`
	Bias                 string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type TrafficPolicyLatencyRule struct {
	EndpointReference    string `json:",omitempty"`
	RuleReference        string `json:",omitempty"`
	Region               string `json:",omitempty"`
	EvaluateTargetHealth *bool  `json:",omitempty"`
	HealthCheck          string `json:",omitempty"`
}

type TrafficPolicyMultiValueAnswerRule struct {
	EndpointReference string `json:",omitempty"`
	HealthCheck       string `json:",omitempty"`
}
