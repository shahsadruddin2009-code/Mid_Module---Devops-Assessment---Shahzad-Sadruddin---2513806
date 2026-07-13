# Infrastructure as Code (Terraform)

This directory provisions the cloud infrastructure that supports the Task Manager
application. It sits **alongside** the CI pipeline in
[`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — the workflow runs and
tests the application, while Terraform manages the infrastructure it depends on.

## What it creates

| Resource | Purpose |
| --- | --- |
| S3 bucket | Encrypted, versioned, private storage for backups of the app's `tasks.json` data. |
| Bucket security controls | Public access block, ownership enforcement and SSE-S3 encryption. |
| Lifecycle policy | Expires old backup versions after `backup_retention_days` and aborts stale multipart uploads. |

All resources follow security best practices: no public access, encryption at
rest enabled, and versioning turned on so backups can be restored.

## Layout

```
infra/
  versions.tf              provider and Terraform version constraints
  variables.tf             input variables with validation
  main.tf                  provider config and resource definitions
  outputs.tf               useful values exported after apply
  terraform.tfvars.example sample variable values (copy to terraform.tfvars)
```

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.6.0
- AWS credentials configured (e.g. `aws configure`, an SSO profile, or the
  `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` environment variables).

## Usage

```bash
cd infra

# Provide your own values
cp terraform.tfvars.example terraform.tfvars

# Standard Terraform workflow
terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply
```

To remove everything again:

```bash
terraform destroy
```

## Notes

- `terraform.tfvars` is git-ignored so environment-specific settings never get
  committed. Commit `terraform.tfvars.example` instead.
- For team use, enable the S3 remote backend block in `versions.tf` so state is
  shared and locked rather than kept on one machine.
