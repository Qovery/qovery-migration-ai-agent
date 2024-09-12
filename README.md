# Qovery AI Migration Agent

Qovery AI Migration Agent is a command-line tool designed to facilitate the migration of applications from various platforms to Qovery. Currently, it supports migrating Heroku applications to AWS, GCP, or Scaleway using Qovery.

## Why this tool?

Migrating applications from one platform to another can be a time-consuming, error-prone process and super costly $$$. Qovery Migration Agent aims to simplify this process by automating the generation of Terraform configurations, Dockerfiles, and other necessary files for deploying applications on Qovery.

## Features

- Migrate Heroku applications to Qovery
- Generate Terraform configurations for Qovery deployments
- Create Dockerfiles for migrated applications
- Support for multiple cloud providers (AWS, GCP, Scaleway)

## Prerequisites

- Go 1.16 or later
- Heroku API Key
- Claude API Key
- Qovery API Key

## Installation

1. Clone the repository:
```
git clone https://github.com/yourusername/qovery-migration-agent.git
```

2. Change to the project directory:
```
cd qovery-migration-agent
```

3. Build the project:
```
go build -o qovery-migration-agent
```

## Usage

1. Set up your environment variables in a `.env` file (or export them in your shell):
```
HEROKU_API_KEY=your_heroku_api_key
CLAUDE_API_KEY=your_claude_api_key
QOVERY_API_KEY=your_qovery_api_key
```

2. Run the migration command:
```
./qovery-migration-agent migrate --from heroku --to aws --output /path/to/output
```

Replace `aws` with `gcp` or `scaleway` as needed.

3. You can now deploy the generated Terraform configurations to Qovery.

```bash
cd /path/to/output

terraform init
terraform apply
```

## How it works

The migration agent uses the Heroku (or other provider) API to fetch information about the application to be migrated. It then generates Terraform configurations for deploying the application on Qovery. The generated Terraform configurations include the necessary resources such as the application, environment, database, and other services.

```mermaid
graph TD
    A[Start] --> B[CLI fetches app data from Heroku API]
    B --> C[CLI filters out sensitive data]
    C --> D[CLI sends non-sensitive app data to Claude AI API]
    D --> E[Claude AI generates Dockerfiles]
    E --> F[CLI receives Dockerfiles]
    F --> G[CLI sends non-sensitive app data to Claude AI API]
    G --> H[Claude AI generates Qovery Terraform files]
    H --> I[CLI receives Terraform files]
    I --> J[CLI reintegrates sensitive data]
    J --> K[CLI generates output for user review]
    K --> L[User reviews and uses output]
    L --> M[End]

    subgraph "Data Sources"
        N[Heroku API]
        O[GitHub Examples]
    end

    subgraph "Security Measure"
        P[Sensitive data kept locally]
    end

    B -.-> N
    H -.-> O
    C -.-> P
    J -.-> P
```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.