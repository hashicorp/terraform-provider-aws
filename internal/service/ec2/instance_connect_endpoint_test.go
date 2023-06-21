package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceConnectEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	//var ec2ice ec2.Ec2InstanceConnectEndpoint
	var ec2DescribeIce ec2v2.DescribeInstanceConnectEndpointsOutput
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceConnectEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName, &ec2DescribeIce),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`instance-connect-endpoint/eice-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
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

func TestAccEC2InstanceConnectEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	//var endpoint ec2.Ec2InstanceConnectEndpoint
	var ec2DescribeIce ec2v2.DescribeInstanceConnectEndpointsOutput
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName, &ec2DescribeIce),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceInstanceConnectEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceConnectEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_instance_connect_endpoint" {
				continue
			}

			_, err := tfec2.FindInstanceConnectEndpointByID(ctx, conn, rs.Primary.ID)

			//fmt.Println("Error from Destroy: DescribeInstanceConnectEndpoints :", err)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(ec2.EndpointsID, create.ErrActionCheckingDestroyed, tfec2.ResNameInstanceConnectEndpoint, rs.Primary.ID, errors.New("not destroyed"))

		}

		return nil
	}
}

func testAccCheckInstanceConnectEndpointExists(ctx context.Context, name string, v *ec2v2.DescribeInstanceConnectEndpointsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameInstanceConnectEndpoint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameInstanceConnectEndpoint, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := &ec2v2.DescribeInstanceConnectEndpointsInput{
			InstanceConnectEndpointIds: []string{rs.Primary.ID},
		}
		output, err := conn.DescribeInstanceConnectEndpoints(ctx, input)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceConnectEndpointConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
 vpc_id     = aws_vpc.test.id
 cidr_block = "10.0.1.0/24"

 tags = {
   Name = %[1]q
 }
}

resource "aws_ec2_instance_connect_endpoint" "test" {
  subnet_id          = aws_subnet.test.id
  security_group_ids = [aws_security_group.test.id]	
  preserve_client_ip = false
}
`, rName)
}

/* func findInstanceConnectEndpointByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Ec2InstanceConnectEndpoint, error) {
	in := &ec2.DescribeInstanceConnectEndpointsInput{
		InstanceConnectEndpointIds: aws.StringSlice([]string{id}),
	}

	out, err := findInstanceConnectEndpoint(ctx, conn, in)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(out.State); state == ec2.Ec2InstanceConnectEndpointStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: in,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(out.InstanceConnectEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: in,
		}
	}

	return out, nil
}

func findInstanceConnectEndpoint(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceConnectEndpointsInput) (*ec2.Ec2InstanceConnectEndpoint, error) {
	out, err := findInstanceConnectEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(out) == 0 || out[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out[0], nil
}

func findInstanceConnectEndpoints(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceConnectEndpointsInput) ([]*ec2.Ec2InstanceConnectEndpoint, error) {
	var out []*ec2.Ec2InstanceConnectEndpoint

	err := conn.DescribeInstanceConnectEndpointsPagesWithContext(ctx, input, func(page *ec2.DescribeInstanceConnectEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceConnectEndpoints {
			if v != nil {
				out = append(out, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}
*/
