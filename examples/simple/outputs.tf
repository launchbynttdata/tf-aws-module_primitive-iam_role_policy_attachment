output "role_name" {
  description = "The name of the IAM role to which the policy is attached."
  value       = aws_iam_role.example.name
}
output "policy_arn" {
  description = "The ARN of the IAM policy that is attached."
  value       = aws_iam_policy.example.arn
}
