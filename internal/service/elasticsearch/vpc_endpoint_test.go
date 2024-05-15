// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticsearch "github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestVPCEndpointErrorsNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		apiObjects []*elasticsearchservice.VpcEndpointError
		notFound   bool
	}{
		{
			name: "nil input",
		},
		{
			name:       "slice of nil input",
			apiObjects: []*elasticsearchservice.VpcEndpointError{nil, nil},
		},
		{
			name: "single SERVER_ERROR",
			apiObjects: []*elasticsearchservice.VpcEndpointError{{
				ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeServerError),
				ErrorMessage:  aws.String("fail"),
				VpcEndpointId: aws.String("aos-12345678"),
			}},
		},
		{
			name: "single ENDPOINT_NOT_FOUND",
			apiObjects: []*elasticsearchservice.VpcEndpointError{{
				ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeEndpointNotFound),
				ErrorMessage:  aws.String("Endpoint does not exist"),
				VpcEndpointId: aws.String("aos-12345678"),
			}},
			notFound: true,
		},
		{
			name: "no ENDPOINT_NOT_FOUND in many",
			apiObjects: []*elasticsearchservice.VpcEndpointError{
				{
					ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("fail"),
					VpcEndpointId: aws.String("aos-abcd0123"),
				},
				{
					ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("crash"),
					VpcEndpointId: aws.String("aos-12345678"),
				},
			},
		},
		{
			name: "single ENDPOINT_NOT_FOUND in many",
			apiObjects: []*elasticsearchservice.VpcEndpointError{
				{
					ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("fail"),
					VpcEndpointId: aws.String("aos-abcd0123"),
				},
				{
					ErrorCode:     aws.String(elasticsearchservice.VpcEndpointErrorCodeEndpointNotFound),
					ErrorMessage:  aws.String("Endpoint does not exist"),
					VpcEndpointId: aws.String("aos-12345678"),
				},
			},
			notFound: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := tfresource.NotFound(tfelasticsearch.VPCEndpointsError(testCase.apiObjects)), testCase.notFound; got != want {
				t.Errorf("NotFound = %v, want %v", got, want)
			}
		})
	}
}

func TestAccElasticsearchVPCEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v elasticsearchservice.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.availability_zones.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_options.0.vpc_id"),
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

func TestAccElasticsearchVPCEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v elasticsearchservice.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticsearch.ResourceVPCEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticsearchVPCEndpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v elasticsearchservice.VpcEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.availability_zones.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_options.0.vpc_id"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_updated(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.availability_zones.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_options.0.vpc_id"),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointExists(ctx context.Context, n string, v *elasticsearchservice.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)

		output, err := tfelasticsearch.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticsearch_vpc_endpoint" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)

			_, err := tfelasticsearch.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("elasticsearch VPC Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCEndpointConfig_base(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[2]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.medium.elasticsearch"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }
}

resource "aws_vpc" "client" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "client" {
  count = 2

  vpc_id            = aws_vpc.client.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.client.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "client" {
  count = 2

  name   = "%[1]s-client-${count.index}"
  vpc_id = aws_vpc.client.id

  tags = {
    Name = %[1]q
  }
}
`, rName, domainName))
}

func testAccVPCEndpointConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_base(rName, domainName), `
resource "aws_elasticsearch_vpc_endpoint" "test" {
  domain_arn = aws_elasticsearch_domain.test.arn

  vpc_options {
    subnet_ids = aws_subnet.client[*].id
  }
}
`)
}

func testAccVPCEndpointConfig_updated(rName, domainName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_base(rName, domainName), `
resource "aws_elasticsearch_vpc_endpoint" "test" {
  domain_arn = aws_elasticsearch_domain.test.arn

  vpc_options {
    subnet_ids         = aws_subnet.client[*].id
    security_group_ids = aws_security_group.client[*].id
  }
}
`)
}
