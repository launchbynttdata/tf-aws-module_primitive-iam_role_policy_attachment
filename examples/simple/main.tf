# Basic usage of the aws_iam_role_policy_attachment module

# data structure for iam_role:

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    actions   = ["ec2:Describe*"]
    resources = ["arn:aws:ec2:*:*:*"]
    effect    = "Allow"
  }
}

resource "aws_iam_policy" "example" {
  name        = "example-policy"
  description = "An example policy"
  policy      = data.aws_iam_policy_document.example.json
}

module "aws_iam_role_policy_attachment" {
  source = "../../"

  role_name  = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example.arn
}
