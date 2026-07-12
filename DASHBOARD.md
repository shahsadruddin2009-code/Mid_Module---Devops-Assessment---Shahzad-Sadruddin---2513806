# CI Dashboard

This file is an automated log of test runs. The CI workflow
([.github/workflows/ci.yml](.github/workflows/ci.yml)) appends one row to the
table below every time it runs on a push, then commits and pushes the update
back to the branch (the log commit itself is skipped by CI so it does not
trigger an infinite loop).

| Timestamp (UTC) | Commit | Branch | Result | Test summary |
|---|---|---|---|---|
