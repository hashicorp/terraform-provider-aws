package connect_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

func testAccCheckRoutingProfileExists(resourceName string, function *connect.DescribeRoutingProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Routing Profile not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Routing Profile ID not set")
		}
		instanceID, routingProfileID, err := tfconnect.RoutingProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(routingProfileID),
		}

		getFunction, err := conn.DescribeRoutingProfile(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckRoutingProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_routing_profile" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, routingProfileID, err := tfconnect.RoutingProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(routingProfileID),
		}

		_, experr := conn.DescribeRoutingProfile(params)
		// Verify the error is what we want
		if experr != nil {
			if awsErr, ok := experr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return experr
		}
	}
	return nil
}
