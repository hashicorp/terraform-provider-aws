package route53_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfrouter53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccRoute53TrafficPolicyDocumentDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyDocumentDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicySameJSON("data.aws_route53_traffic_policy_document.test",
						testAccTrafficPolicyDocumentConfigExpectedJSON()),
				),
			},
		},
	})
}

func TestAccRoute53TrafficPolicyDocumentDataSource_complete(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyDocumentDataSourceConfig_complete,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicySameJSON("data.aws_route53_traffic_policy_document.test",
						testAccTrafficPolicyDocumentConfigCompleteExpectedJSON()),
				),
			},
		},
	})
}

func testAccCheckTrafficPolicySameJSON(resourceName, jsonExpected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		var j, j2 tfrouter53.Route53TrafficPolicyDoc
		if err := json.Unmarshal([]byte(rs.Primary.Attributes["json"]), &j); err != nil {
			return fmt.Errorf("[ERROR] json.Unmarshal %v", err)
		}
		if err := json.Unmarshal([]byte(jsonExpected), &j2); err != nil {
			return fmt.Errorf("[ERROR] json.Unmarshal %v", err)
		}
		// Marshall again so it can re order the json data because of arrays
		jsonDoc, err := json.Marshal(j)
		if err != nil {
			return fmt.Errorf("[ERROR] json.marshal %v", err)
		}
		jsonDoc2, err := json.Marshal(j2)
		if err != nil {
			return fmt.Errorf("[ERROR] json.marshal %v", err)
		}
		if err = json.Unmarshal(jsonDoc, &j); err != nil {
			return fmt.Errorf("[ERROR] json.Unmarshal %v", err)
		}
		if err = json.Unmarshal(jsonDoc2, &j); err != nil {
			return fmt.Errorf("[ERROR] json.Unmarshal %v", err)
		}

		if !awsutil.DeepEqual(&j, &j2) {
			return fmt.Errorf("expected out to be %v, got %v", j, j2)
		}

		return nil
	}
}

func testAccTrafficPolicyDocumentConfigCompleteExpectedJSON() string {
	return fmt.Sprintf(`{
  "AWSPolicyFormatVersion":"2015-10-01",
  "RecordType":"A",
  "StartRule":"geo_restriction",
  "Endpoints":{
    "east_coast_lb1":{
      "Type":"elastic-load-balancer",
      "Value":"elb-111111.%[1]s.elb.amazonaws.com"
    },
    "east_coast_lb2":{
      "Type":"elastic-load-balancer",
      "Value":"elb-222222.%[1]s.elb.amazonaws.com"
    },
    "west_coast_lb1":{
      "Type":"elastic-load-balancer",
      "Value":"elb-111111.%[2]s.elb.amazonaws.com"
    },
    "west_coast_lb2":{
      "Type":"elastic-load-balancer",
      "Value":"elb-222222.%[2]s.elb.amazonaws.com"
    },
    "denied_message":{
      "Type":"s3-website",
      "Region":"%[1]s",
      "Value":"video.example.com"
    }
  },
  "Rules":{
    "geo_restriction":{
      "RuleType":"geo",
      "Locations":[
        {
          "EndpointReference":"denied_message",
          "IsDefault":true
        },
        {
          "RuleReference":"region_selector",
          "Country":"US"
        }
      ]
    },
    "region_selector":{
      "RuleType":"latency",
      "Regions":[
        {
          "Region":"%[1]s",
          "RuleReference":"east_coast_region"
        },
        {
          "Region":"%[2]s",
          "RuleReference":"west_coast_region"
        }
      ]
    },
    "east_coast_region":{
      "RuleType":"failover",
      "Primary":{
        "EndpointReference":"east_coast_lb1"
      },
      "Secondary":{
        "EndpointReference":"east_coast_lb2"
      }
    },
    "west_coast_region":{
      "RuleType":"failover",
      "Primary":{
        "EndpointReference":"west_coast_lb1"
      },
      "Secondary":{
        "EndpointReference":"west_coast_lb2"
      }
    }
  }
}`, acctest.Region(), acctest.AlternateRegion())
}

const testAccTrafficPolicyDocumentDataSourceConfig_basic = `
data "aws_region" "current" {}

data "aws_route53_traffic_policy_document" "test" {
  record_type = "A"
  start_rule  = "site_switch"

  endpoint {
    id    = "my_elb"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_region.current.name}.elb.amazonaws.com"
  }
  endpoint {
    id     = "site_down_banner"
    type   = "s3-website"
    region = data.aws_region.current.name
    value  = "www.example.com"
  }

  rule {
    id   = "site_switch"
    type = "failover"

    primary {
      endpoint_reference = "my_elb"
    }
    secondary {
      endpoint_reference = "site_down_banner"
    }
  }
}
`

const testAccTrafficPolicyDocumentDataSourceConfig_complete = `
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_route53_traffic_policy_document" "test" {
  version     = "2015-10-01"
  record_type = "A"
  start_rule  = "geo_restriction"

  endpoint {
    id    = "east_coast_lb1"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_availability_zones.available.names[0]}.elb.amazonaws.com"
  }
  endpoint {
    id    = "east_coast_lb2"
    type  = "elastic-load-balancer"
    value = "elb-222222.${data.aws_availability_zones.available.names[0]}.elb.amazonaws.com"
  }
  endpoint {
    id    = "west_coast_lb1"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_availability_zones.available.names[1]}.elb.amazonaws.com"
  }
  endpoint {
    id    = "west_coast_lb2"
    type  = "elastic-load-balancer"
    value = "elb-222222.${data.aws_availability_zones.available.names[1]}.elb.amazonaws.com"
  }
  endpoint {
    id     = "denied_message"
    type   = "s3-website"
    region = data.aws_availability_zones.available.names[0]
    value  = "video.example.com"
  }

  rule {
    id   = "geo_restriction"
    type = "geo"

    location {
      endpoint_reference = "denied_message"
      is_default         = true
    }
    location {
      rule_reference = "region_selector"
      country        = "US"
    }
  }
  rule {
    id   = "region_selector"
    type = "latency"

    region {
      region         = data.aws_availability_zones.available.names[0]
      rule_reference = "east_coast_region"
    }
    region {
      region         = data.aws_availability_zones.available.names[1]
      rule_reference = "west_coast_region"
    }
  }
  rule {
    id   = "east_coast_region"
    type = "failover"

    primary {
      endpoint_reference = "east_coast_lb1"
    }
    secondary {
      endpoint_reference = "east_coast_lb2"
    }
  }
  rule {
    id   = "west_coast_region"
    type = "failover"

    primary {
      endpoint_reference = "west_coast_lb1"
    }
    secondary {
      endpoint_reference = "west_coast_lb2"
    }
  }
}
`

func testAccTrafficPolicyDocumentConfigExpectedJSON() string {
	return fmt.Sprintf(`{
   "AWSPolicyFormatVersion":"2015-10-01",
   "RecordType":"A",
   "StartRule":"site_switch",
   "Endpoints":{
      "my_elb":{
         "Type":"elastic-load-balancer",
         "Value":"elb-111111.%[1]s.elb.amazonaws.com"
      },
      "site_down_banner":{
         "Type":"s3-website",
         "Region":"%[1]s",
         "Value":"www.example.com"
      }
   },
   "Rules":{
      "site_switch":{
         "RuleType":"failover",
         "Primary":{
            "EndpointReference":"my_elb"
         },
         "Secondary":{
            "EndpointReference":"site_down_banner"
         }
      }
   }
}`, acctest.Region())
}
