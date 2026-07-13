terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Remote state is recommended for real deployments. Uncomment and configure
  # the backend below once a state bucket and lock table exist.
  #
  # backend "s3" {
  #   bucket         = "my-terraform-state-bucket"
  #   key            = "task-manager/terraform.tfstate"
  #   region         = "eu-west-2"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }
}
