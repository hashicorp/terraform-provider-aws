# Provide AWS Credentials
provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "us-east-1"
}

# Create S3 Full Access Policy
resource "aws_iam_policy" "s3_policy" {
  name        = "s3-policy"
  description = "Policy for allowing all S3 Actions."

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": "*"
        }
    ]
}
EOF
}

# Create API Gateway Role
resource "aws_iam_role" "s3_api_gateyway_role" {
  name = "s3-api-gateyway-role"

  # Create Trust Policy for API Gateway
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
} 
  EOF
}

# Attach S3 Access Policy to the API Gateway Role
resource "aws_iam_role_policy_attachment" "s3_policy_attach" {
  role       = "${aws_iam_role.s3_api_gateyway_role.name}"
  policy_arn = "${aws_iam_policy.s3_policy.arn}"
}

# resource "aws_api_gateway_rest_api" "MyDemoAPI" {
#   name        = "MyDemoAPI"
#   description = "This is my API for demonstration purposes"
# }


# resource "aws_api_gateway_resource" "MyDemoResource" {
#   rest_api_id = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
#   parent_id   = "${aws_api_gateway_rest_api.MyDemoAPI.root_resource_id}"
#   path_part   = "mydemoresource"
# }


# resource "aws_api_gateway_method" "MyDemoMethod" {
#   rest_api_id   = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
#   resource_id   = "${aws_api_gateway_resource.MyDemoResource.id}"
#   http_method   = "GET"
#   authorization = "NONE"
# }


# resource "aws_api_gateway_integration" "MyDemoIntegration" {
#   rest_api_id          = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
#   resource_id          = "${aws_api_gateway_resource.MyDemoResource.id}"
#   http_method          = "${aws_api_gateway_method.MyDemoMethod.http_method}"
#   type                 = "MOCK"
#   cache_key_parameters = ["method.request.path.param"]
#   cache_namespace      = "foobar"
#   timeout_milliseconds = 29000


#   request_parameters = {
#     "integration.request.header.X-Authorization" = "'static'"
#   }


#   # Transforms the incoming XML request to JSON
#   request_templates {
#     "application/xml" = <<EOF
# {
#    "body" : $input.json('$')
# }
# EOF
#   }
# }

