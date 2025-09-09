MailHog CI/CD Pipeline
=====================

Modern continuous integration and deployment pipeline using GitHub Actions.

## Pipeline Overview

The CI/CD pipeline runs on every push and pull request to the `master` branch, providing comprehensive testing and build verification.

### Pipeline Jobs

#### 1. Test Job
- **Runs on**: Ubuntu Latest  
- **Go Versions**: 1.24, 1.25 (matrix)
- **Features**:
  - Go module caching for faster builds
  - Unit test execution with coverage
  - Coverage upload to Codecov

```yaml
strategy:
  matrix:
    go-version: ['1.24', '1.25']
```

#### 2. Build Job  
- **Runs after**: Test job passes
- **Purpose**: Verify application builds correctly
- **Artifact**: Uploads `mailhog` binary for integration testing

#### 3. Integration Test Job
- **Runs after**: Build job passes  
- **Tests**:
  - Health endpoints (`/health`, `/ready`, `/metrics`)
  - SMTP server connectivity (port 1025)
  - Web UI accessibility (port 8025)

#### 4. Docker Job
- **Runs on**: Master branch pushes only
- **Features**: Docker image building with layer caching

## Health Endpoint Testing

The pipeline specifically tests our new health endpoints:

```bash
# Health endpoint validation
curl -f -s http://localhost:8025/health | jq '.'

# Readiness endpoint validation  
curl -f -s http://localhost:8025/ready | jq '.'

# Metrics endpoint validation
curl -f -s http://localhost:8025/metrics | head -10
```

## Running Locally

### Prerequisites
- Go 1.24+ installed
- `make` available
- `jq` for JSON validation (optional)

### Commands
```bash
# Run unit tests
make test

# Run tests with coverage
go test -cover ./api/

# Build application
go build -o mailhog .

# Run integration tests manually
./mailhog &
sleep 3
curl http://localhost:8025/health
curl http://localhost:8025/ready
curl http://localhost:8025/metrics
pkill mailhog
```

## Migration from Travis CI

This project has migrated from Travis CI to GitHub Actions for better integration and modern tooling:

### Before (Travis CI)
- Go 1.6 (deprecated)
- No health endpoint testing
- Basic build verification only

### After (GitHub Actions)  
- Go 1.24 & 1.25 (latest)
- Full health endpoint integration testing
- Comprehensive build and test matrix
- Docker image building
- Coverage reporting

## Pipeline Configuration

The complete pipeline is defined in `.github/workflows/ci.yml` and includes:

- **Dependency Caching**: Speeds up builds
- **Matrix Testing**: Tests multiple Go versions
- **Artifact Management**: Builds and shares binaries between jobs
- **Integration Testing**: Real application testing
- **Conditional Docker**: Builds images only on master

## Monitoring

The pipeline provides several monitoring capabilities:

1. **Test Coverage**: Uploaded to Codecov
2. **Build Status**: Visible in GitHub PRs
3. **Integration Results**: Health endpoint validation
4. **Artifact Storage**: Binary availability for download

## Security

- No secrets required for basic pipeline
- Docker builds use GitHub's container registry
- All tests run in isolated environments
- No external dependencies in testing