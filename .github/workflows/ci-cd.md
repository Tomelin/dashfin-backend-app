# CI/CD Workflow Documentation (`ci-cd.yml`)

## 1. Overview

This document describes the CI/CD (Continuous Integration/Continuous Deployment) workflow defined in `.github/workflows/ci-cd.yml`. The workflow automates the process of linting, testing, building, and deploying the Go application. It also includes steps for AI code review, security scanning, documentation generation, and notifications.

**Workflow Triggers:**

The workflow is triggered on:
-   **Push** events to the `main` branch.
-   **Pull Request** events targeting the `main` branch (excluding drafts for AI reviews and security scans where specified).

## 2. Workflow Stages

The workflow consists of the following sequential and parallel jobs:

### 2.1. `lint`
-   **Name:** Lint Code
-   **Description:** Checks the Go source code for stylistic issues and potential errors using `golangci-lint`.
-   **Trigger:** Runs on every push and pull request to `main`.
-   **Details:** Uses the `golangci/golangci-lint-action@v4`.

### 2.2. `test`
-   **Name:** Unit Tests & Coverage
-   **Depends On:** `lint`
-   **Description:** Runs unit tests, checks for race conditions, and generates a code coverage report.
-   **Trigger:** Runs after `lint` succeeds.
-   **Details:**
    -   Executes `go test -v -race -coverprofile=coverage.out ./...`.
    -   Generates an HTML coverage report: `go tool cover -html=coverage.out -o coverage.html`.
    -   Uploads `coverage.out` and `coverage.html` as workflow artifacts named `coverage-report`.

### 2.3. `build`
-   **Name:** Build Application and Docker Image
-   **Depends On:** `test`
-   **Description:** Compiles the Go application and then builds a Docker image. If the event is a push to `main`, the Docker image is pushed to the configured container registry. For Pull Requests, this job should be configured to load the built Docker image into the local Docker daemon (e.g., using `load: true` in the `docker/build-push-action`) to make it available for the `security-scan` job. The image is not pushed for PRs.
-   **Trigger:** Runs after `test` succeeds.
-   **Details:**
    -   Compiles the Go application: `go build -v -o myapp ./...` (outputs a binary named `myapp`).
    -   **On push to `main` only:**
        -   Logs into Docker Hub (or your configured registry) using `secrets.DOCKER_USERNAME` and `secrets.DOCKER_PASSWORD`.
        -   Extracts image metadata (tags like commit SHA, `latest` for main branch pushes) using `docker/metadata-action`.
        -   Builds the Docker image from the `Dockerfile` in the repository root.
        -   Pushes the image to `${env.DOCKER_REGISTRY}/${env.DOCKER_NAMESPACE}/myapp` (e.g., `docker.io/yourdockerhubusername/myappname`). This needs to be configured via environment variables in the workflow file.

### 2.4. `generate-docs`
-   **Name:** Generate Go Documentation
-   **Depends On:** `test`
-   **Description:** Generates HTML documentation for the Go source code using `godoc`.
-   **Trigger:** Runs after `test` succeeds (can run in parallel with `build` and AI review jobs).
-   **Details:**
    -   Installs `godoc` tool.
    -   Runs `godoc -html ./... > godocs.html`.
    -   Uploads `godocs.html` as a workflow artifact named `godoc-html`.

### 2.5. `gemini-review`
-   **Name:** Gemini Code Review
-   **Depends On:** `test`
-   **Description:** Uses Google's Gemini model to analyze changed Go files and provide feedback. Comments are posted directly on the Pull Request. This job only runs on non-draft Pull Request events.
-   **Trigger:** Runs on non-draft Pull Requests to `main`, after `test` succeeds.
-   **Details:**
    -   Requires `GEMINI_API_KEY` secret.
    -   Uses the `.github/scripts/analyze-gemini-go.js` script.

### 2.6. `claude-review`
-   **Name:** Claude Code Review
-   **Depends On:** `test`
-   **Description:** Uses Anthropic's Claude model to analyze changed Go files and provide feedback. Comments are posted directly on the Pull Request. This job only runs on non-draft Pull Request events.
-   **Trigger:** Runs on non-draft Pull Requests to `main`, after `test` succeeds (in parallel with `gemini-review`).
-   **Details:**
    -   Requires `CLAUDE_API_KEY` secret.
    -   Uses the `.github/scripts/analyze-claude-go.js` script.

### 2.7. `ai-review-summary`
-   **Name:** AI Review Summary
-   **Depends On:** `gemini-review`, `claude-review`
-   **Description:** Posts a summary comment to the Pull Request based on the outcomes of the Gemini and Claude AI review jobs.
-   **Trigger:** Runs `if: always()` on non-draft Pull Requests to `main`, after both AI review jobs have finished.
-   **Details:**
    -   Uses the `.github/scripts/post-summary.js` script.

### 2.8. `security-scan`
-   **Name:** Security Vulnerability Scan (Trivy)
-   **Depends On:** `build`
-   **Description:** Scans the built Docker image for known vulnerabilities using Aqua Security's Trivy.
-   **Trigger:** Runs after `build` succeeds, on pushes to `main` (where image is pushed) and non-draft Pull Requests to `main` (where image is assumed to be built locally by the `build` job).
-   **Details:**
    -   Scans the image `${env.DOCKER_REGISTRY}/${env.DOCKER_NAMESPACE}/myapp:${{ github.sha }}`.
    -   Fails if `CRITICAL` or `HIGH` severity vulnerabilities are found (`exit-code: '1'`).
    -   Generates a `trivy-results.sarif` file, which is uploaded as an artifact named `trivy-scan-results`.
    -   The SARIF file is also uploaded to the GitHub Security tab for review under the category `myapp-docker-image`.

### 2.9. `deploy-dev`
-   **Name:** Deploy to Development
-   **Depends On:** `build` (Note: For stricter control, could depend on `security-scan` if successful scan is a prerequisite for any deployment). Current setup runs it in parallel to security scan for faster feedback, but a failed security scan would still fail the overall workflow.
-   **Description:** Deploys the application (using the Docker image) to the development environment. This stage contains placeholder scripts and needs to be configured with actual deployment commands.
-   **Trigger:** Runs after `build` succeeds, **only on push to `main`**.
-   **GitHub Environment:** Uses the `development` environment.
-   **Details:**
    -   Placeholder scripts demonstrate deploying the image `${env.DOCKER_REGISTRY }/${{ env.DOCKER_NAMESPACE }}/myapp:${{ github.sha }}`. The image name components (`DOCKER_REGISTRY`, `DOCKER_NAMESPACE`, `myapp`) should match your configuration.
    -   Requires secrets like `DEV_SERVER_HOST`, `DEV_SERVER_USER`, etc., for accessing the development environment.

### 2.10. `deploy-prod`
-   **Name:** Deploy to Production
-   **Depends On:** `deploy-dev`
-   **Description:** Deploys the application to the production environment. This stage is critical and uses GitHub Environments, which can be configured to require manual approval.
-   **Trigger:** Runs after `deploy-dev` succeeds, **only on push to `main`**.
-   **GitHub Environment:** Uses the `production` environment (e.g., `http://your-app-production-url.com`). This URL is a placeholder and should be updated. This environment should be configured with protection rules (like required reviewers) in GitHub settings.
-   **Details:**
    -   Placeholder scripts demonstrate deploying the image.
    -   Requires secrets like `PROD_SERVER_HOST`, `PROD_SERVER_USER`, etc.
    -   **Caution:** Actual deployment scripts must be robust and may include health checks.

### 2.11. `notify`
-   **Name:** Send Workflow Notification
-   **Depends On:** `lint`, `test`, `build`, `security-scan`, `generate-docs`, `gemini-review`, `claude-review`, `ai-review-summary`, `deploy-dev`, `deploy-prod`
-   **Description:** Sends a notification summarizing the outcome of the workflow.
-   **Trigger:** Runs `if: always()`, meaning it executes even if preceding jobs fail, to report the overall status.
-   **Details:**
    -   Collects results from all dependent jobs.
    -   Calculates an overall workflow status (SUCCESS, FAILURE, CANCELLED).
    -   Placeholder for sending a message to a platform like Slack or Microsoft Teams. Requires secrets like `SLACK_BOT_TOKEN` and a channel ID.

## 3. Configuration

### 3.1. GitHub Secrets

The following secrets must be configured in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

-   `DOCKER_USERNAME`: Username for your Docker registry (e.g., Docker Hub).
-   `DOCKER_PASSWORD`: Password or access token for your Docker registry.
-   `GEMINI_API_KEY`: API key for Google Gemini AI model, used for code review.
-   `CLAUDE_API_KEY`: API key for Anthropic Claude AI model, used for code review.
-   `DEV_SERVER_HOST`: Hostname/IP of the development server.
-   `DEV_SERVER_USER`: Username for deploying to the development server.
-   `SSH_PRIVATE_KEY_DEV` (Optional): SSH private key for development server access.
-   `KUBE_CONFIG_DEV` (Optional): Kubeconfig data for Kubernetes deployment to dev.
-   `PROD_SERVER_HOST`: Hostname/IP of the production server.
-   `PROD_SERVER_USER`: Username for deploying to the production server.
-   `SSH_PRIVATE_KEY_PROD` (Optional): SSH private key for production server access.
-   `KUBE_CONFIG_PROD` (Optional): Kubeconfig data for Kubernetes deployment to prod.
-   `SLACK_BOT_TOKEN` (Optional): Slack bot token for notifications.
-   Any other secrets required by your specific deployment scripts.

### 3.2. Global Environment Variables (in `ci-cd.yml`)

The workflow uses global `env` variables that might need review:

-   `GO_VERSION`: Currently set in the workflow (e.g., '1.21'). Adjust if your project needs a different Go version.
-   `DOCKER_REGISTRY`: Currently set in the workflow (e.g., 'docker.io'). Change if using GitHub Container Registry (ghcr.io) or other.
-   `DOCKER_NAMESPACE`: Currently set in the workflow (e.g., 'yourusername'). This is your Docker Hub username or organization and needs to be configured. The image name (e.g., `myapp`) is also set in the workflow and should be reviewed.

### 3.3. GitHub Environments

-   **`development`**: Used by the `deploy-dev` job. Can be configured with environment-specific secrets and variables. The example URL (`http://dev.example.com`) in the workflow is a placeholder.
-   **`production`**: Used by the `deploy-prod` job. The example URL (`http://your-production-app-url.com`) in the workflow is a placeholder. **It is highly recommended to configure this environment with protection rules**, such as:
    -   **Required reviewers:** Specify users or teams that must approve deployments to production.
    -   **Wait timer:** A delay before deployment.
    -   This provides a manual gate before production deployment.

### 3.4. `Dockerfile`

Ensure a `Dockerfile` is present at the root of your repository. The `build` job uses it to create the Docker image.

## 4. Monitoring Pipeline Status

-   **GitHub Actions Tab:** The status of each workflow run can be monitored directly in your GitHub repository under the "Actions" tab.
-   **Notifications:** If configured, notifications (e.g., Slack messages) will provide summaries of workflow completion and status.
-   **Artifacts:** Build artifacts, coverage reports, GoDocs HTML files, and Trivy SARIF reports are uploaded and can be downloaded from the workflow run summary page for a limited time (default 7 days for docs/coverage/scan results).

## 5. Related Practices

### 5.1. Observability (Metrics & Logs)

-   The CI/CD pipeline deploys the application, but it does not directly handle runtime observability.
-   Your Go application **must be instrumented** using libraries like OpenTelemetry or specific SDKs for your chosen observability platform (e.g., Prometheus, Grafana, Datadog, New Relic).
-   The application should send metrics (e.g., request rates, error rates, latency) and structured logs to your observability system.

### 5.2. Pull Request Reviews & Branch Protection

-   **Code Reviews:** All code changes should go through pull requests and be reviewed by team members before merging into `main`. This includes manual reviews in addition to the automated AI reviews.
-   **Branch Protection Rules:** Configure branch protection rules for your `main` branch in GitHub repository settings (`Settings > Branches > Add rule`):
    -   **Require status checks to pass before merging:** Select the jobs from this CI/CD workflow (e.g., `lint`, `test`, `build`, `security-scan`) as required checks. This ensures that PRs can only be merged if the CI pipeline succeeds.
    -   **Require pull request reviews before merging.**
    -   Optionally, require up-to-date branches, signed commits, etc.
-   The AI review jobs integrated into this workflow (`gemini-review`, `claude-review`) contribute to code quality by providing automated suggestions on pull requests.

This comprehensive CI/CD setup, combined with good development and review practices, helps ensure code quality and reliable deployments.
