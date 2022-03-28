package opensearch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UpgradeStatusUnknown = "Unknown"
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
