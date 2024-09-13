# Qovery AI Migration Agent

Qovery AI Migration Agent is an app designed to facilitate the migration of applications from various platforms to Qovery. Currently, it supports migrating Heroku applications to AWS, GCP, or Scaleway using Qovery.

## Why this tool?

Migrating applications from one platform to another can be a time-consuming, error-prone process and super costly $$$. Qovery Migration Agent aims to simplify this process by automating the generation of Terraform configurations, Dockerfiles, and other necessary files for deploying applications on Qovery.

## Features

> Note: This project is still in development and may not support all features yet.

- Migrate Heroku/Render applications to AWS, GCP, Azure or Scaleway via Qovery
- Generate Terraform configurations for Qovery deployments
- Create Dockerfiles for migrated applications

## Structure

The project is structured as follows:

- [CLI](cli): Contains the command-line interface for the migration agent (can be used on your local machine)
- [Web](web): Contains the web interface for the migration agent (can be deployed on a server)

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
    J --> K[Validate Terraform manifest]
    K --> L{Is manifest valid?}
    L -->|Yes| M[CLI generates output for user review]
    L -->|No| N[Auto-remediation process]
    N --> O[Claude AI fixes Terraform errors]
    O --> K
    M --> P[User reviews and uses output]
    P --> Q[End]

    subgraph "Data Sources"
        R[Heroku API]
        S[GitHub Examples]
    end

    subgraph "Security Measure"
        T[Sensitive data kept locally]
    end

    B -.-> R
    H -.-> S
    C -.-> T
    J -.-> T
```

## Security

- This application does not store any user credentials.
- All code is open-source and can be audited.
- For more information, see the Security page in the application.

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for more details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions, please file an issue on the GitHub repository or contact support@qovery.com.
