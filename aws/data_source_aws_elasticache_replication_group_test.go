package aws

import (
	"fmt"
	"testing"
)

func TestAccDataSourceAwsElasticacheReplicationGroup_basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(),
      },
    },
  }
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_basic() string {
	return fmt.Sprintf(``)
}
