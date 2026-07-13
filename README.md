# Task Manager — CSO7024 Mid-Module Assessment starter

This is the starter application for the **CSO7024 DevOps** mid-module assessment.
It is a small, working command-line task manager written in Python. Your job in
the assessment is to extend it with one new feature while managing your work with
Git and a simple Continuous Integration (CI) workflow. The full assessment brief
on Canvas explains how everything fits together; this file covers the application
itself and the feature you must add.

## What the application does

The task manager keeps a list of tasks. Each task has an `id`, a `title` and a
`done` flag. From the command line you can add tasks, list them, mark one as done,
and remove one. Tasks are stored in a `tasks.json` file in the current directory.

## Before you start

You need **Python 3.10 or newer** and **Git** installed. Working in a virtual
environment is recommended but not required.

## Install

```bash
python -m venv .venv
source .venv/bin/activate        # on Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

## Run the application

```bash
python -m taskmanager.cli add "Write the report"
python -m taskmanager.cli add "Email the team"
python -m taskmanager.cli list
python -m taskmanager.cli done 1
python -m taskmanager.cli remove 2
```

## Run the tests

```bash
pytest
```

All tests should pass before you make any changes. Confirm this first, so you
know your environment is set up correctly.

## Project layout

```
taskmanager/
  core.py     the task operations (add, complete, remove, load, save)
  cli.py      the command-line interface
tests/
  test_core.py   tests for the existing operations
infra/        Terraform (Infrastructure as Code) — see below
requirements.txt
```

## Continuous Integration and Infrastructure as Code

This project uses two separate DevOps tools that do different jobs:

| | [`.github/workflows/ci.yml`](.github/workflows/ci.yml) | [`infra/`](infra) |
| --- | --- | --- |
| Tool | GitHub Actions (YAML) | Terraform (HCL) |
| Job | **CI automation** — runs the Python and Go tests on every push and appends a row to [`DASHBOARD.md`](DASHBOARD.md) | **Infrastructure as Code** — provisions the AWS resources the app depends on |
| Runs on | GitHub's runners, automatically on push/PR | Your machine (or a CI job), on demand via `terraform apply` |

They are complementary, not interchangeable: GitHub Actions cannot execute a
Terraform configuration as a workflow, and Terraform does not run tests.

### What the Terraform config creates

The [`infra/`](infra) directory provisions a private, encrypted, versioned S3
bucket for backups of the app's `tasks.json` data, following security best
practices (public access fully blocked, SSE-S3 encryption, versioning, and a
lifecycle policy that expires old backup versions).

```
infra/
  versions.tf              provider and Terraform version constraints
  variables.tf              input variables with validation
  main.tf                   provider config and resource definitions
  outputs.tf                useful values exported after apply
  terraform.tfvars.example  sample variable values (copy to terraform.tfvars)
```

**Prerequisites:** [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.6.0 and AWS credentials configured (e.g. `aws configure`, an SSO profile, or the `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` environment variables).

**Usage:**

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars   # provide your own values

terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply
```

To remove everything again: `terraform destroy`.

Notes:
- `terraform.tfvars` is git-ignored so environment-specific settings are never
  committed; commit `terraform.tfvars.example` instead.
- For team use, enable the S3 remote backend block in `infra/versions.tf` so
  state is shared and locked rather than kept on one machine.

## Your task: add a task priority feature

Add the ability to give each task a **priority** and to list only the tasks that
match a chosen priority. Implement it exactly as specified below, because the
function names and behaviour are what the marker will look for.

**1. Extend `add_task` in `taskmanager/core.py`.**
Change its signature to:

```python
def add_task(tasks: list[dict], title: str, priority: str = "medium") -> list[dict]:
```

- `priority` must be one of `"high"`, `"medium"` or `"low"`.
- If any other value is given, raise a `ValueError`.
- Store the chosen priority on the new task under a `"priority"` key, alongside
  the existing `id`, `title` and `done` fields.
- When `priority` is not supplied it defaults to `"medium"`, so the existing
  tests, which call `add_task(tasks, "Title")`, must still pass.

**2. Add a new function `tasks_with_priority` in `taskmanager/core.py`:**

```python
def tasks_with_priority(tasks: list[dict], priority: str) -> list[dict]:
```

- Return a new list containing only the tasks whose `"priority"` equals the
  given `priority`, in their original order.
- Do not modify the input list.

**3. Update the command-line interface in `taskmanager/cli.py`.**

- Give the `add` command an optional `--priority` argument (one of `high`,
  `medium`, `low`, defaulting to `medium`) and pass it through to `add_task`.
- Give the `list` command an optional `--priority` argument that, when supplied,
  shows only the matching tasks (use `tasks_with_priority`).

**4. Add at least one new automated test** for the feature, in its own test file
(for example `tests/test_priority.py`). Cover both adding a task with a priority
and filtering by priority, and check that an invalid priority raises `ValueError`.

Keep your change small and focused, and make sure the existing tests in
`tests/test_core.py` still pass.
