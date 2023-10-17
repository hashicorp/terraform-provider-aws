package apigatewayv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayV2VPCLinkDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_vpc_link.test"
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkDataSourceConfig_base(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "Arn", resourceName, "Arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "Id", resourceName, "Id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "Name", resourceName, "Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "SecurityGroupIds", resourceName, "SecurityGroupIds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "SubnetIds", resourceName, "SubnetIds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key1", resourceName, "Value1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key2", resourceName, "Value2"),
				),
			},
		},
	})
}

func testAccVPCLinkDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test" {
		cidr_block = "10.0.0.0/16"
	  
		tags = {
		  Name = %[1]q
		}
	  }
	  
	  data "aws_availability_zones" "available" {
		state = "available"
	  
		filter {
		  name   = "opt-in-status"
		  values = ["opt-in-not-required"]
		}
	  }
	  
	  resource "aws_subnet" "test" {
		count = 2
	  
		vpc_id            = aws_vpc.test.id
		cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
		availability_zone = data.aws_availability_zones.available.names[count.index]
	  
		tags = {
		  Name = %[1]q
		}
	  }

	  resource "aws_security_group" "test" {
		vpc_id = aws_vpc.test.id
	  
		tags = {
		  Name = %[1]q
		}
	  }

	  resource "aws_apigatewayv2_vpc_link" "test" {
		name               = %[1]q
		security_group_ids = [aws_security_group.test.id]
		subnet_ids         = aws_subnet.test[*].id
		
		tags = {
			Key1 = "Value1"
			Key2 = "Value2"
		  }
	  }

	  data "aws_apigatewayv2_vpc_link" "test" {
		id = aws_apigatewayv2_vpc_link.test.id
	  }
`, rName)
}
