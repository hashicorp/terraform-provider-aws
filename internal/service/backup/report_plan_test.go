package backup_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccCheckReportPlanDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_report_plan" {
			continue
		}

		input := &backup.DescribeReportPlanInput{
			ReportPlanName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeReportPlan(input)

		if err == nil {
			if aws.StringValue(resp.ReportPlan.ReportPlanName) == rs.Primary.ID {
				return fmt.Errorf("Backup Report Plan '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckReportPlanExists(name string, reportPlan *backup.DescribeReportPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
		input := &backup.DescribeReportPlanInput{
			ReportPlanName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeReportPlan(input)

		if err != nil {
			return err
		}

		*reportPlan = *resp

		return nil
	}
}
