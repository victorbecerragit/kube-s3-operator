# Contributing to kube-s3-operator

Thank you for your interest in contributing to kube-s3-operator! 🎉

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce**
- **Expected vs actual behavior**
- **Environment details** (Kubernetes version, operator version, etc.)
- **Logs** (sanitized of sensitive information)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear use case description**
- **Proposed solution**
- **Alternative solutions considered**
- **Impact on existing functionality**

### Pull Requests

1. **Fork the repository** and create your branch from `main`:

git checkout -b feature/my-new-feature


2. **Make your changes** following our coding standards:
- Write clear, commented code
- Follow Go best practices
- Add tests for new functionality
- Update documentation as needed

3. **Test your changes**:

make test

make lint


4. **Commit your changes** using conventional commits:


git commit -m "feat: add support for bucket encryption"


5. **Push to your fork**:

git push origin feature/my-new-feature


6. **Open a Pull Request** with:
- Clear description of changes
- Link to related issues
- Screenshots/logs if applicable

## Development Setup

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- kind or similar local Kubernetes cluster
- AWS credentials for testing

### Building Locally


Clone the repository
git clone https://github.com/victorbecerragit/kube-s3-operator.git
cd kube-s3-operator

Install dependencies
go mod download

Build the operator
make build

Run tests
make test

Run locally (requires cluster)
make run


### Testing


Run unit tests
make test

Run integration tests
make test-integration

Run e2e tests
make test-e2e


## Coding Guidelines

### Go Style

- Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation changes
- `test:` adding or updating tests
- `refactor:` code refactoring
- `chore:` maintenance tasks

### Branch Naming

- `feature/` - for new features
- `fix/` - for bug fixes
- `docs/` - for documentation
- `refactor/` - for refactoring

## Release Process

Releases are automated through GitHub Actions. Maintainers will:

1. Create a release branch
2. Update CHANGELOG.md
3. Tag the release
4. GitHub Actions will build and publish artifacts

## Getting Help

- 💬 Join our [Discussions](https://github.com/victorbecerragit/kube-s3-operator/discussions)
- 🐛 Report issues via [GitHub Issues](https://github.com/victorbecerragit/kube-s3-operator/issues)
- 📧 Email maintainers (see README)

## Recognition

Contributors will be recognized in our README. We use [All Contributors](https://allcontributors.org/) specification.


