# Qovery Migration AI Agent - Web

This project is an AI-powered tool to help migrate stacks from PaaS providers (Heroku, Render) to IaaS providers (AWS, GCP, Scaleway, Azure) using Qovery.

## Features

- User-friendly web interface for selecting source and destination platforms
- Secure handling of credentials
- Automated generation of Terraform manifests and Dockerfiles
- Downloadable migration files
- Step-by-step guidance for executing the migration

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Node.js (for local development)
- Go (for local development)

### Running the Application

1. Clone the repository:
```
git clone https://github.com/yourusername/qovery-migration-ai-agent.git
cd qovery-migration-ai-agent
```

2. Start the application using Docker Compose:
```
docker-compose up --build
```

3. Open your browser and navigate to `http://localhost:3000` to access the application.

### Development

For local development:

1. Start the backend:
```
cd backend
go run main.go
```

2. Start the frontend:
```
cd frontend
npm install
npm start
```