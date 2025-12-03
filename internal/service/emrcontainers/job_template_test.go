// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/emrcontainers/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfemrcontainers "github.com/hashicorp/terraform-provider-aws/internal/service/emrcontainers"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRContainersJobTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.JobTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_job_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "job_template_data.0.execution_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.0.spark_sql_job_driver.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.0.spark_sql_job_driver.0.entry_point", "default"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.release_label", "emr-6.10.0-latest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEMRContainersJobTemplate_configurationOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.JobTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_job_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_configurationOverrides(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "job_template_data.0.execution_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.0.spark_sql_job_driver.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.job_driver.0.spark_sql_job_driver.0.entry_point", "default"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.release_label", "emr-6.10.0-latest"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.application_configuration.0.classification", "spark-defaults"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.application_configuration.0.properties.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.application_configuration.0.properties.spark.executor.memory", "4G"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.application_configuration.0.properties.spark.driver.memory", "2G"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.monitoring_configuration.0.cloud_watch_monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.monitoring_configuration.0.cloud_watch_monitoring_configuration.0.log_group_name", "/emr/"+rName),
					resource.TestCheckResourceAttr(resourceName, "job_template_data.0.configuration_overrides.0.monitoring_configuration.0.cloud_watch_monitoring_configuration.0.log_stream_name_prefix", "spark-job-logs"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEMRContainersJobTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.JobTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_job_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemrcontainers.ResourceJobTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRContainersJobTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.JobTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_job_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckJobTemplateExists(ctx context.Context, n string, v *awstypes.JobTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Containers Job Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

		output, err := tfemrcontainers.FindJobTemplateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emrcontainers_job_template" {
				continue
			}

			_, err := tfemrcontainers.FindJobTemplateByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EMR Containers Job Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccJobTemplateConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "eks.${data.aws_partition.current.dns_suffix}",
          "eks-nodegroup.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccJobTemplateConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccJobTemplateConfig_base(rName),
		fmt.Sprintf(`
resource "aws_emrcontainers_job_template" "test" {
  job_template_data {
    execution_role_arn = aws_iam_role.test.arn
    release_label      = "emr-6.10.0-latest"

    job_driver {
      spark_sql_job_driver {
        entry_point = "default"
      }
    }
  }

  name = %[1]q
}
`, rName))
}

func testAccJobTemplateConfig_configurationOverrides(rName string) string {
	return acctest.ConfigCompose(
		testAccJobTemplateConfig_base(rName),
		fmt.Sprintf(`
resource "aws_emrcontainers_job_template" "test" {
  job_template_data {
    execution_role_arn = aws_iam_role.test.arn
    release_label      = "emr-6.10.0-latest"

    job_driver {
      spark_sql_job_driver {
        entry_point = "default"
      }
    }
    configuration_overrides {
      application_configuration {
        classification = "spark-defaults"
        properties = {
          "spark.executor.memory" = "4G"
          "spark.driver.memory"   = "2G"
        }
      }
      monitoring_configuration {
        cloud_watch_monitoring_configuration {
          log_group_name         = "/emr/%[1]s"
          log_stream_name_prefix = "spark-job-logs"
        }
      }
    }
  }

  name = %[1]q
}
`, rName))
}

func testAccJobTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccJobTemplateConfig_base(rName),
		fmt.Sprintf(`
resource "aws_emrcontainers_job_template" "test" {
  job_template_data {
    execution_role_arn = aws_iam_role.test.arn
    release_label      = "emr-6.10.0-latest"

    job_driver {
      spark_sql_job_driver {
        entry_point = "default"
      }
    }
  }

  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

}
`, rName, tagKey1, tagValue1))
}
