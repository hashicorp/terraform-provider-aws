package elasticsearch

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UpgradeStatusUnknown = "Unknown"
	ConfigStatusNotFound = "NotFound"
	ConfigStatusUnknown  = "Unknown"
	ConfigStatusExists   = "Exists"
)

func statusUpgradeStatus(conn *elasticsearch.ElasticsearchService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.GetUpgradeStatus(&elasticsearch.GetUpgradeStatusInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			return nil, UpgradeStatusUnknown, err
		}

		// Elasticsearch upgrades consist of multiple steps:
		// https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html
		// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
		if aws.StringValue(out.StepStatus) == elasticsearch.UpgradeStatusSucceeded && aws.StringValue(out.UpgradeStep) != elasticsearch.UpgradeStepUpgrade {
			return out, elasticsearch.UpgradeStatusInProgress, nil
		}

		return out, aws.StringValue(out.StepStatus), nil
	}
}

func domainConfigStatus(conn *elasticsearch.ElasticsearchService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
			DomainName: aws.String(name),
		})

		if tfawserr.ErrCodeEquals(err, elasticsearch.ErrCodeResourceNotFoundException) {
			// if first return value is nil, WaitForState treats as not found - here not found is treated differently
			return "not nil", ConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, ConfigStatusUnknown, err
		}

		return out, ConfigStatusExists, nil
	}
}
