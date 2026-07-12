"""Initialise (if needed) a standalone git repo for this project folder and
push its contents to the GitHub repository:

    https://github.com/shahsadruddin2009-code/Mid_Module---Devops-Assessment---Shahzad-Sadruddin---2513806

This script only ever operates on THIS folder. It does not touch, walk into,
or modify any parent/ancestor git repository (e.g. one rooted higher up in
your home directory).

Usage:
    python push_to_github.py [-m "commit message"] [-b branch] [--remote-url URL]

Requirements:
    - git must be installed and on PATH.
    - You must be authenticated with GitHub for pushes (e.g. via a stored
      credential, SSH key, or the Git Credential Manager prompting for a
      username/PAT).
"""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

REPO_DIR = Path(__file__).resolve().parent
DEFAULT_REMOTE_URL = (
    "https://github.com/shahsadruddin2009-code/"
    "Mid_Module---Devops-Assessment---Shahzad-Sadruddin---2513806.git"
)
DEFAULT_BRANCH = "main"
DEFAULT_COMMIT_MESSAGE = "Update project files"


def run(args: list[str]) -> subprocess.CompletedProcess:
    """Run a git command inside REPO_DIR, raising on failure."""
    print(f"$ git {' '.join(args)}")
    result = subprocess.run(
        ["git", *args],
        cwd=REPO_DIR,
        text=True,
        capture_output=True,
    )
    if result.stdout:
        print(result.stdout.rstrip())
    if result.returncode != 0:
        if result.stderr:
            print(result.stderr.rstrip(), file=sys.stderr)
        raise SystemExit(f"git {args[0]} failed with exit code {result.returncode}")
    if result.stderr:
        # git prints informational messages (e.g. branch creation) to stderr
        print(result.stderr.rstrip())
    return result


def is_own_git_repo() -> bool:
    """True only if REPO_DIR itself (not an ancestor) is a git repo root."""
    return (REPO_DIR / ".git").is_dir()


def has_commits() -> bool:
    result = subprocess.run(
        ["git", "rev-parse", "--verify", "HEAD"],
        cwd=REPO_DIR,
        text=True,
        capture_output=True,
    )
    return result.returncode == 0


def ensure_remote(remote_url: str) -> None:
    result = subprocess.run(
        ["git", "remote", "get-url", "origin"],
        cwd=REPO_DIR,
        text=True,
        capture_output=True,
    )
    if result.returncode == 0:
        current_url = result.stdout.strip()
        if current_url != remote_url:
            run(["remote", "set-url", "origin", remote_url])
    else:
        run(["remote", "add", "origin", remote_url])


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("-m", "--message", default=DEFAULT_COMMIT_MESSAGE, help="Commit message")
    parser.add_argument("-b", "--branch", default=DEFAULT_BRANCH, help="Branch to push to")
    parser.add_argument("--remote-url", default=DEFAULT_REMOTE_URL, help="Git remote URL to push to")
    args = parser.parse_args()

    if not is_own_git_repo():
        print(f"Initialising a new git repository in: {REPO_DIR}")
        run(["init", "-b", args.branch])
    else:
        print(f"Using existing git repository in: {REPO_DIR}")

    ensure_remote(args.remote_url)

    run(["add", "-A"])

    status = subprocess.run(
        ["git", "diff", "--cached", "--quiet"],
        cwd=REPO_DIR,
    )
    if status.returncode == 0 and has_commits():
        print("Nothing to commit; working tree is clean.")
    else:
        run(["commit", "-m", args.message])

    run(["push", "-u", "origin", args.branch])
    print("Done. Files pushed to:", args.remote_url)


if __name__ == "__main__":
    main()
