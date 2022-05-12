package opensearch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UpgradeStatusUnknown = "Unknown"
	ConfigStatusNotFound = "NotFound"
	ConfigStatusUnknown  = "Unknown"
	ConfigStatusExists   = "Exists"
)

func statusUpgradeStatus(conn *opensearchservice.OpenSearchService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.GetUpgradeStatus(&opensearchservice.GetUpgradeStatusInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			return nil, UpgradeStatusUnknown, err
		}

		// opensearch upgrades consist of multiple steps:
		// https://docs.aws.amazon.com/opensearch-service/latest/developerguide/opensearch-version-migration.html
		// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
		if aws.StringValue(out.StepStatus) == opensearchservice.UpgradeStatusSucceeded && aws.StringValue(out.UpgradeStep) != opensearchservice.UpgradeStepUpgrade {
			return out, opensearchservice.UpgradeStatusInProgress, nil
		}

		return out, aws.StringValue(out.StepStatus), nil
	}
}

func domainConfigStatus(conn *opensearchservice.OpenSearchService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeDomainConfig(&opensearchservice.DescribeDomainConfigInput{
			DomainName: aws.String(name),
		})

		if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
			// if first return value is nil, WaitForState treats as not found - here not found is treated differently
			return "not nil", ConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, ConfigStatusUnknown, err
		}

		return out, ConfigStatusExists, nil
	}
}
