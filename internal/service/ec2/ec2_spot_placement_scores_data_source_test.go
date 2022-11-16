package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2SpotPlacementScoresDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_placement_scores.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2SpotPlacementScoresDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs(
						dataSourceName,
						"spot_placement_score_sets.*",
						map[string]*regexp.Regexp{
							"region": regexp.MustCompile(`^[a-z1-3\-]*$`),
							"score":  regexp.MustCompile(`^[1-9]0?$`),
						},
					),
				),
			},
		},
	})
}

func TestAccEC2SpotPlacementScoresDataSource_singleAz(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_placement_scores.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2SpotPlacementScoresDataSourceConfig_singleAz(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs(
						dataSourceName,
						"spot_placement_score_sets.*",
						map[string]*regexp.Regexp{
							"availability_zone_id": regexp.MustCompile(`^[a-z1-3\-]*$`),
							"region":               regexp.MustCompile(`^[a-z1-3\-]*$`),
							"score":                regexp.MustCompile(`^[1-9]0?$`),
						},
					),
				),
			},
		},
	})
}

func TestAccEC2SpotPlacementScoresDataSource_withMetadata(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_placement_scores.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2SpotPlacementScoresDataSourceConfig_withMetadata(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs(
						dataSourceName,
						"spot_placement_score_sets.*",
						map[string]*regexp.Regexp{
							"availability_zone_id": regexp.MustCompile(`^[a-z1-3\-]*$`),
							"region":               regexp.MustCompile(`^[a-z1-3\-]*$`),
							"score":                regexp.MustCompile(`^[1-9]0?$`),
						},
					),
				),
			},
		},
	})
}

func testAccEC2SpotPlacementScoresDataSourceConfig_basic() string {
	return acctest.ConfigCompose(`
data "aws_ec2_spot_placement_scores" "test" {
  max_results = 10
  target_capacity = 10
  instance_types = ["m3.xlarge","t4g.small","p4d.24xlarge"]
}
`)
}

func testAccEC2SpotPlacementScoresDataSourceConfig_singleAz() string {
	return acctest.ConfigCompose(`
data "aws_ec2_spot_placement_scores" "test" {
  max_results = 10
  single_availability_zone = true
  target_capacity = 10
  instance_types = ["m3.xlarge","t4g.small","p4d.24xlarge"]
}
`)
}

func testAccEC2SpotPlacementScoresDataSourceConfig_withMetadata() string {
	return acctest.ConfigCompose(`
data "aws_ec2_spot_placement_scores" "test" {
  max_results = 10
  single_availability_zone = true
  target_capacity = 1
  region_names = ["us-east-1","us-east-2","eu-west-1","eu-west-2"]
  
  instance_requirements_with_metadata {
    architecture_types = ["x86_64","arm64","x86_64_mac"]
    instance_requirements {
      vcpu_count {
        min = 1
        max = 8
      }

      memory_mib {
        min = 100
        max = 50000
      }

      instance_generations = ["current"]
    }
    virtualization_types = ["paravirtual", "hvm"]
  }
}
`)
}
