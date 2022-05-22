package kendra_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kendra"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccCheckIndexDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_index" {
			continue
		}

		input := &kendra.DescribeIndexInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeIndex(input)

		if err == nil {
			if aws.StringValue(resp.Id) == rs.Primary.ID {
				return fmt.Errorf("Index '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckIndexExists(name string, index *kendra.DescribeIndexOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn
		input := &kendra.DescribeIndexInput{
			Id: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeIndex(input)

		if err != nil {
			return err
		}

		*index = *resp

		return nil
	}
}
