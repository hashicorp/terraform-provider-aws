package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncResolver_basic(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_resolver.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolver_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile("apis/.+/types/.+/resolvers/.+")),
					resource.TestCheckResourceAttr(resourceName, "data_source", rName),
					resource.TestCheckResourceAttrSet(resourceName, "request_template"),
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

func TestAccAwsAppsyncResolver_DataSource(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_resolver.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolver_DataSource(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppsyncResolver_DataSource(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source", rName2),
				),
			},
		},
	})
}

func TestAccAwsAppsyncResolver_RequestTemplate(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_resolver.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolver_RequestTemplate(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "request_template", regexp.MustCompile("resourcePath\": \"/\"")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppsyncResolver_RequestTemplate(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "request_template", regexp.MustCompile("resourcePath\": \"/test\"")),
				),
			},
		},
	})
}

func TestAccAwsAppsyncResolver_ResponseTemplate(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_resolver.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncResolverDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncResolver_ResponseTemplate(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "response_template", regexp.MustCompile("ctx\\.result\\.statusCode == 200")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppsyncResolver_ResponseTemplate(rName, 201),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncResolverExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "response_template", regexp.MustCompile("ctx\\.result\\.statusCode == 201")),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncResolverDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_resolver" {
			continue
		}

		apiID, typeName, fieldName, err := decodeAppsyncResolverID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appsync.GetResolverInput{
			ApiId:     aws.String(apiID),
			TypeName:  aws.String(typeName),
			FieldName: aws.String(fieldName),
		}

		_, err = conn.GetResolver(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncResolverExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource has no ID: %s", name)
		}

		apiID, typeName, fieldName, err := decodeAppsyncResolverID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetResolverInput{
			ApiId:     aws.String(apiID),
			TypeName:  aws.String(typeName),
			FieldName: aws.String(fieldName),
		}

		_, err = conn.GetResolver(input)

		return err
	}
}

func testAccAppsyncResolver_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = <<EOF
type Mutation {
	putPost(id: ID!, title: String!): Post
}

type Post {
	id: ID!
	title: String!
}

type Query {
	singlePost(id: ID!): Post
}

schema {
	query: Query
	mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  name        = %q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_resolver" "test" {
  api_id           = "${aws_appsync_graphql_api.test.id}"
  field            = "singlePost"
  type             = "Query"
  data_source      = "${aws_appsync_datasource.test.name}"
  request_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF
  response_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, rName, rName)
}

func testAccAppsyncResolver_DataSource(rName, dataSource string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = <<EOF
type Mutation {
	putPost(id: ID!, title: String!): Post
}

type Post {
	id: ID!
	title: String!
}

type Query {
	singlePost(id: ID!): Post
}

schema {
	query: Query
	mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  name        = %q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_resolver" "test" {
  api_id           = "${aws_appsync_graphql_api.test.id}"
  field            = "singlePost"
  type             = "Query"
  data_source      = "${aws_appsync_datasource.test.name}"
  request_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF
  response_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, rName, dataSource)
}

func testAccAppsyncResolver_RequestTemplate(rName, resourcePath string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = <<EOF
type Mutation {
	putPost(id: ID!, title: String!): Post
}

type Post {
	id: ID!
	title: String!
}

type Query {
	singlePost(id: ID!): Post
}

schema {
	query: Query
	mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  name        = %q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_resolver" "test" {
  api_id           = "${aws_appsync_graphql_api.test.id}"
  field            = "singlePost"
  type             = "Query"
  data_source      = "${aws_appsync_datasource.test.name}"
  request_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": %q,
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF
  response_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, rName, rName, resourcePath)
}

func testAccAppsyncResolver_ResponseTemplate(rName string, statusCode int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = <<EOF
type Mutation {
	putPost(id: ID!, title: String!): Post
}

type Post {
	id: ID!
	title: String!
}

type Query {
	singlePost(id: ID!): Post
}

schema {
	query: Query
	mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  name        = %q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_resolver" "test" {
  api_id           = "${aws_appsync_graphql_api.test.id}"
  field            = "singlePost"
  type             = "Query"
  data_source      = "${aws_appsync_datasource.test.name}"
  request_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        ## you can forward the headers using the below utility
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF
  response_template = <<EOF
#if($ctx.result.statusCode == %d)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, rName, rName, statusCode)
}
