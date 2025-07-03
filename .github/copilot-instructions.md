# OctoCAT Supply Chain management Application

# This file contains instructions for GitHub Copilot to assist in generating code and commit messages.

## GitHub Repo information
This repo is hosted in GitHub:
- **Repository Owner**: octodemo
- **Repository URL**: https://github.com/your-username/octocat-supply-chain-management
- **Repository Name**: octocat-supply-chain-management
- **Repository Description**: A supply chain management application built with Go, integrating OpenAI and Gemini APIs for intelligent decision-making.
- **Repository Topics**: go, openai, gemini, supply-chain-management, ai-integration, automation, microservices

## Architecture Overview
The complete Architecture is described in the [ARCHITECTURE](../docs/ARCHITECTURE.md) file, which outlines the components and their interactions within the application.

# Build and Run Instructions

## Build Instructions
Refer to [Build Instructions](../docs/BUILD.md) for detailed steps on how to build the application.

Every time you change the code, make sure that the code compiles by running:
```bash
go build ./...
```

## Run Instructions
To run the application, use the following command:
```bash
go run main.go
```

# This will start the application and you can access it at `http://localhost:8080`.


# Commit Message Generation Instructions

## Conventional Commit Messages
The project follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for commit messages. This helps in maintaining a clear and consistent history of changes.

**EXAMPLE** commit message format:

```markdown
feat: add new payment method

- Added support for credit card payments
- Updated checkout flow to include new payment method
- Refactored payment processing code for better readability

Closes #123
```

## Commit Message Generation Instructions
When generating commit messages, follow these guidelines:
1. **Be Detailed**: Provide a comprehensive description of the changes made, including the purpose and impact of the changes.
2. **Use Emojis**: Add relevant emojis to make the commit message more engaging and visually appealing.
3. **Friendly Tone**: Use a friendly and approachable tone in the commit message.
4. **Structure**: Use a structured format for the commit message, including:
   - A brief summary of the changes
   - Detailed description of the changes
   - Any relevant links or references to issues or documentation
5. **Examples**: Include examples of how the changes can be tested or used.


## Example Commit Message
```markdown
feat: add new payment method

- Added support for credit card payments
- Updated checkout flow to include new payment method
- Refactored payment processing code for better readability

Closes #123 
```