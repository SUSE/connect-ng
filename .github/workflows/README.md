# Notes About CI Workflow Security Restrictions

About security when running the tests and NOT exposing
the secrets to externals. Currently Github Actions does
NOT expose the secrets if the branch is coming from a forked
repository.

See: https://github.blog/2020-08-03-github-actions-improvements-for-fork-and-pull-request-workflows/
See: https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions

An alternate would be to set, pull_request_target but this takes the CI code
from master removing the ability to change the code in a PR easily.

Aditionally, since 2021 pull requests from new contributors will not
trigger workflows automatically but will wait for approval from somebody
with write access.
See: https://docs.github.com/en/actions/managing-workflow-runs/approving-workflow-runs-from-public-forks