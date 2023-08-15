// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestVPCEndpointErrorsNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		apiObjects []*opensearchservice.VpcEndpointError
		notFound   bool
	}{
		{
			name: "nil input",
		},
		{
			name:       "slice of nil input",
			apiObjects: []*opensearchservice.VpcEndpointError{nil, nil},
		},
		{
			name: "single SERVER_ERROR",
			apiObjects: []*opensearchservice.VpcEndpointError{&opensearchservice.VpcEndpointError{
				ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeServerError),
				ErrorMessage:  aws.String("fail"),
				VpcEndpointId: aws.String("aos-12345678"),
			}},
		},
		{
			name: "single ENDPOINT_NOT_FOUND",
			apiObjects: []*opensearchservice.VpcEndpointError{&opensearchservice.VpcEndpointError{
				ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeEndpointNotFound),
				ErrorMessage:  aws.String("Endpoint does not exist"),
				VpcEndpointId: aws.String("aos-12345678"),
			}},
			notFound: true,
		},
		{
			name: "no ENDPOINT_NOT_FOUND in many",
			apiObjects: []*opensearchservice.VpcEndpointError{
				&opensearchservice.VpcEndpointError{
					ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("fail"),
					VpcEndpointId: aws.String("aos-abcd0123"),
				},
				&opensearchservice.VpcEndpointError{
					ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("crash"),
					VpcEndpointId: aws.String("aos-12345678"),
				},
			},
		},
		{
			name: "single ENDPOINT_NOT_FOUND in many",
			apiObjects: []*opensearchservice.VpcEndpointError{
				&opensearchservice.VpcEndpointError{
					ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeServerError),
					ErrorMessage:  aws.String("fail"),
					VpcEndpointId: aws.String("aos-abcd0123"),
				},
				&opensearchservice.VpcEndpointError{
					ErrorCode:     aws.String(opensearchservice.VpcEndpointErrorCodeEndpointNotFound),
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

			if got, want := tfresource.NotFound(tfopensearch.VPCEndpointsError(testCase.apiObjects)), testCase.notFound; got != want {
				t.Errorf("NotFound = %v, want %v", got, want)
			}
		})
	}
}

func TestAccOpenSearchVPCEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_vpc_endpoint.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					resource.TestCheckResourceAttr(resourceName, "connection_status", "ACTIVE"),
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

func TestAccOpenSearchVPCEndpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_vpc_endpoint.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_status", "ACTIVE"),
				),
			},
			{
				Config: testAccVPCEndpointConfigUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "connection_status", "ACTIVE"),
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

func TestAccOpenSearchVPCEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	ri := sdkacctest.RandString(10)
	name := fmt.Sprintf("tf-test-%s", ri)
	resourceName := "aws_opensearch_vpc_endpoint.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.domain_1", &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceVPCEndpoint(), resourceName),
				),
			},
		},
	})
}

func testAccVPCEndpointConfig(name string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "domain_1" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t3.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}

resource "aws_opensearch_vpc_endpoint" "foo" {
  domain_arn = aws_opensearch_domain.domain_1.arn
  vpc_options {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, name))
}

func testAccVPCEndpointConfigUpdate(name string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "domain_1" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t3.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}

resource "aws_opensearch_vpc_endpoint" "foo" {
  domain_arn = aws_opensearch_domain.domain_1.arn
  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, name))
}
