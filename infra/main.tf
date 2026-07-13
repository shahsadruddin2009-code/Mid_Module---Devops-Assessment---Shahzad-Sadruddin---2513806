provider "aws" {
  region = var.aws_region
}

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = merge(
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
      Application = "task-manager-cli"
    },
    var.tags,
  )
}

# Unique suffix so the globally-unique bucket name does not collide.
resource "random_id" "suffix" {
  byte_length = 4
}

# ---------------------------------------------------------------------------
# S3 bucket used to store backups of the application's tasks.json data.
# Hardened following AWS security best practices: private, encrypted,
# versioned, with public access fully blocked.
# ---------------------------------------------------------------------------
resource "aws_s3_bucket" "task_backups" {
  bucket = "${local.name_prefix}-backups-${random_id.suffix.hex}"
  tags   = local.common_tags
}

resource "aws_s3_bucket_ownership_controls" "task_backups" {
  bucket = aws_s3_bucket.task_backups.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_public_access_block" "task_backups" {
  bucket = aws_s3_bucket.task_backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "task_backups" {
  bucket = aws_s3_bucket.task_backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "task_backups" {
  bucket = aws_s3_bucket.task_backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "task_backups" {
  bucket = aws_s3_bucket.task_backups.id

  # Ensure a versioning rule exists before lifecycle rules are applied.
  depends_on = [aws_s3_bucket_versioning.task_backups]

  rule {
    id     = "expire-old-backups"
    status = "Enabled"

    filter {
      prefix = ""
    }

    noncurrent_version_expiration {
      noncurrent_days = var.backup_retention_days
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
