resource "aws_iam_role" "ecs_execution_role" {
  name = "tf-example-ecs-task-execution-role"

  assume_role_policy = <<DEFINITION
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
DEFINITION
}

resource "aws_iam_role_policy" "ecs_execution_role_policy" {
  name = "tf-example-ecs-execution-role-policy"
  role = "${aws_iam_role.ecs_execution_role.id}"

  policy = <<DEFINITION
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
DEFINITION
}
