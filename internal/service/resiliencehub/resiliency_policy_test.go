package resiliencehub

import (
	"context"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResiliencyPolicy_basic(t *testing.T) {
	resourceName := "aws_resiliency_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, account.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps:             nil,
	})
}

func testResiliencyPolicyDestroy(s *terraform.State) {
	ctx := context.TODO()
	conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_resiliency_policy" {
			continue
		}

	}

	return nil
}
