output "backup_bucket_name" {
  description = "Name of the S3 bucket that stores task-data backups."
  value       = aws_s3_bucket.task_backups.bucket
}

output "backup_bucket_arn" {
  description = "ARN of the task-backups S3 bucket."
  value       = aws_s3_bucket.task_backups.arn
}

output "backup_bucket_region" {
  description = "Region in which the task-backups bucket was created."
  value       = aws_s3_bucket.task_backups.region
}
