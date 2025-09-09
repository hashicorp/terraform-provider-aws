list "aws_batch_job_queue" "queues" {
  provider = aws
  # config {
  #     region = "us-west-2"
  # }
}

list "aws_cloudwatch_log_group" "log_groups" {
  provider = aws
  # config {
  #     region = "us-west-2"
  # }
}

list "aws_instance" "instances" {
  provider = aws
  # config {
  #     region = "us-west-2"
  # }
}
