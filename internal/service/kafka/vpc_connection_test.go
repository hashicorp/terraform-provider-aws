package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
)

func TestAccKafkaVpcConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcconnection kafka.DescribeVpcConnectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_vpc_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVpcConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcConnectionExists(ctx, resourceName, &vpcconnection),
					// acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`configuration/.+`)),
					// resource.TestCheckResourceAttr(resourceName, "description", ""),
					// resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "latest_revision", "1"),
					// resource.TestCheckResourceAttr(resourceName, "name", rName),
					// resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(`auto.create.topics.enable = true`)),
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

// func TestAccKafkaVpcConnection_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var vpcconnection kafka.DescribeVpcConnectionResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_kafka_vpc_connection.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.KafkaEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckVpcConnectionDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccVpcConnectionConfig_basic(rName, testAccVpcConnectionVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckVpcConnectionExists(ctx, resourceName, &vpcconnection),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceVpcConnection = newResourceVpcConnection
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceVpcConnection, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckVpcConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_vpc_connection" {
				continue
			}

			_, err := tfkafka.FindVpcConnectionByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK VPC Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
func testAccCheckVpcConnectionExists(ctx context.Context, name string, vpcconnection *kafka.DescribeVpcConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Serverless Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		output, err := tfkafka.FindVpcConnectionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*vpcconnection = *output

		return nil
	}
}

// func (ctx context.Context, name string, vpcconnection *kafka.CreateVpcConnectionOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameVpcConnection, name, errors.New("not found"))
// 		}

// 		if rs.Primary.ID == "" {
// 			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameVpcConnection, name, errors.New("not set"))
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)
// 		resp, err := conn.DescribeVpcConnection(ctx, &kafka.DescribeVpcConnectionInput{
// 			Arn: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameVpcConnection, rs.Primary.ID, err)
// 		}

// 		*vpcconnection = *resp

// 		return nil
// 	}
// }

// func testAccCheckVpcConnectionNotRecreated(before, after *kafka.DescribeVpcConnectionResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.VpcConnectionId), aws.ToString(after.VpcConnectionId); before != after {
// 			return create.Error(names.Kafka, create.ErrActionCheckingNotRecreated, tfkafka.ResNameVpcConnection, aws.ToString(before.VpcConnectionId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccVpcConnectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_vpc_connection" "test" {
	authentication = "IAM"
	target_cluster_arn = "arn:aws:kafka:eu-west-2:926562225508:cluster/demo-cluster-1/a7640874-7bdf-4a38-be10-24465449a333-2"
	vpc_id = aws_vpc.test.id
	client_subnets = aws_subnet.test[*].id
	security_groups = [aws_security_group.test.id]
}

`, rName))
}

// IAM, TLS and SASL/SCRAM
