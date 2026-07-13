variable "aws_region" {
  description = "AWS region in which to create resources."
  type        = string
  default     = "eu-west-2"
}

variable "project_name" {
  description = "Short project identifier used to name and tag resources."
  type        = string
  default     = "task-manager"

  validation {
    condition     = can(regex("^[a-z0-9-]{3,32}$", var.project_name))
    error_message = "project_name must be 3-32 chars: lowercase letters, digits and hyphens only."
  }
}

variable "environment" {
  description = "Deployment environment (e.g. dev, staging, prod)."
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "backup_retention_days" {
  description = "Number of days to retain non-current task-backup versions before deletion."
  type        = number
  default     = 30

  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 365
    error_message = "backup_retention_days must be between 1 and 365."
  }
}

variable "tags" {
  description = "Additional tags applied to every resource."
  type        = map(string)
  default     = {}
}
