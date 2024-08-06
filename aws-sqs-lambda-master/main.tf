// IAM Role for Lambda Function
resource "aws_iam_role" "default" {
  name               = var.service_name
  description        = "IAM Role for ${var.service_name}"
  assume_role_policy = file("iam/${var.service_name}_role.json")
}

resource "aws_iam_policy" "default" {
  name        = var.service_name
  description = "IAM Policy for ${var.service_name}"
  policy      = file("iam/${var.service_name}_policy.json")
}

resource "aws_iam_role_policy_attachment" "default" {
  role       = aws_iam_role.default.name
  policy_arn = aws_iam_policy.default.arn
}

// Lambda Function Resources
resource "aws_cloudwatch_log_group" "default" {
  name              = "/aws/lambda/${var.service_name}"
  retention_in_days = 7
}

data "archive_file" "default" {
  type        = "zip"
  source_dir  = "src"
  output_path = var.output_path
}

resource "aws_lambda_function" "default" {
  filename         = var.output_path
  function_name    = var.service_name
  role             = aws_iam_role.default.arn
  handler          = "lambda_function.handler"
  source_code_hash = data.archive_file.default.output_base64sha256
  runtime          = "nodejs20.x"
  timeout = 120
}

// TODO SNS Resources
/*
resource "aws_sns_topic" "default" {
  name            = var.service_name
  delivery_policy = file("sns_delivery_policy.json")
}

resource "aws_sns_topic_subscription" "default" {
  topic_arn = aws_sns_topic.default.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.default.arn
}
*/

// SQS
resource "aws_sqs_queue" "default" {
  name                       = "${var.service_name}_sqs"
  visibility_timeout_seconds = 12

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.default_deadletter.arn
    maxReceiveCount     = 4
  })
}

// Lambda SQS mapping
resource "aws_lambda_event_source_mapping" "default" {
  event_source_arn = aws_sqs_queue.default.arn
  function_name    = aws_lambda_function.default.arn
}

// Dead Letter Queue
resource "aws_sqs_queue" "default_deadletter" {
  name = "${var.service_name}_sqs_dlq"
}

resource "aws_sqs_queue_redrive_allow_policy" "terraform_queue_redrive_allow_policy" {
  queue_url = aws_sqs_queue.default_deadletter.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.default.arn]
  })
}
