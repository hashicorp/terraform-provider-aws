package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/dms/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func EndpointStatus(conn *dms.DatabaseMigrationService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.EndpointByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
