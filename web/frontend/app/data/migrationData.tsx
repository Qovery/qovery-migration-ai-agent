interface MigrationStep {
    title: string;
    description: string;
}

interface MigrationBenefit {
    title: string;
    description: string;
}

interface CustomerQuote {
    text: string;
    author: string;
    company: string;
}

interface IaaSData {
    steps: MigrationStep[];
    benefits: MigrationBenefit[];
    customerQuote?: CustomerQuote;
}

interface PaaSData {
    [iaas: string]: IaaSData;
}

interface MigrationData {
    [paas: string]: PaaSData;
}

export const migrationData: MigrationData = {
    heroku: {
        aws: {
            steps: [
                {
                    title: "Export Heroku Configuration",
                    description: "Extract your Heroku app's configuration and environment variables using the Heroku CLI."
                },
                {
                    title: "Set up AWS Infrastructure",
                    description: "Use Terraform or AWS CloudFormation to create the necessary AWS resources based on your Heroku setup."
                },
                {
                    title: "Adapt Database",
                    description: "Migrate your Heroku PostgreSQL to Amazon RDS, ensuring data integrity and minimal downtime."
                },
                {
                    title: "Configure CI/CD",
                    description: "Set up AWS CodePipeline or another CI/CD tool to replicate your Heroku deployment process."
                },
                {
                    title: "Deploy to AWS",
                    description: "Push your application code to the newly created AWS infrastructure, such as Elastic Beanstalk or ECS."
                }
            ],
            benefits: [
                {
                    title: "Greater Infrastructure Control",
                    description: "Gain fine-grained control over your infrastructure configuration and scaling options."
                },
                {
                    title: "Cost Optimization",
                    description: "Potential for significant cost savings, especially for larger applications with predictable workloads."
                },
                {
                    title: "Ecosystem Integration",
                    description: "Seamlessly integrate with a wide range of AWS services for enhanced functionality."
                },
                {
                    title: "Improved Scalability",
                    description: "Leverage AWS's robust scaling capabilities to handle varying loads more efficiently."
                }
            ],
            customerQuote: {
                text: "Migrating from Heroku to AWS with Qovery's AI agent was surprisingly smooth. We saw immediate improvements in performance and significant cost savings.",
                author: "Jane Doe",
                company: "TechStartup Inc."
            }
        },
        gcp: {
            steps: [
                {
                    title: "Export Heroku Configuration",
                    description: "Extract your Heroku app's configuration and environment variables using the Heroku CLI."
                },
                {
                    title: "Set up GCP Project",
                    description: "Create a new GCP project and enable necessary APIs for your application's requirements."
                },
                {
                    title: "Configure Cloud SQL",
                    description: "Set up a Cloud SQL instance to replace your Heroku PostgreSQL database."
                },
                {
                    title: "Adapt Application",
                    description: "Modify your application code to work with GCP services and environment variables."
                },
                {
                    title: "Deploy to GCP",
                    description: "Use Google Cloud Run or App Engine to deploy your application, setting up continuous deployment."
                }
            ],
            benefits: [
                {
                    title: "Advanced Analytics",
                    description: "Leverage GCP's powerful data analytics and machine learning capabilities."
                },
                {
                    title: "Global Network",
                    description: "Utilize Google's global network for improved performance and reduced latency."
                },
                {
                    title: "Kubernetes Integration",
                    description: "Easily migrate to a Kubernetes-based infrastructure with Google Kubernetes Engine if needed."
                },
                {
                    title: "Flexible Scaling",
                    description: "Take advantage of GCP's auto-scaling features for optimal resource utilization."
                }
            ],
            // No customer quote for this migration path
        },
    },
    render: {
        aws: {
            steps: [
                {
                    title: "Export Render Configuration",
                    description: "Document your Render service configurations, including environment variables and build settings."
                },
                {
                    title: "Set up AWS Infrastructure",
                    description: "Use AWS Elastic Beanstalk or ECS to create a similar environment to your Render setup."
                },
                {
                    title: "Migrate Database",
                    description: "If using Render's managed PostgreSQL, migrate it to Amazon RDS for PostgreSQL."
                },
                {
                    title: "Adapt Deployment Process",
                    description: "Set up AWS CodePipeline to mirror Render's Git-based deployment process."
                },
                {
                    title: "Update DNS",
                    description: "Update your DNS settings to point to your new AWS resources instead of Render."
                }
            ],
            benefits: [
                {
                    title: "Extensive Service Integration",
                    description: "Access a wide range of AWS services to extend your application's capabilities."
                },
                {
                    title: "Customizable Infrastructure",
                    description: "Gain more control over your infrastructure setup and configuration."
                },
                {
                    title: "Advanced Monitoring",
                    description: "Utilize AWS CloudWatch for comprehensive monitoring and alerting."
                },
                {
                    title: "Global Reach",
                    description: "Leverage AWS's global infrastructure for improved performance and redundancy."
                }
            ],
            customerQuote: {
                text: "Moving from Render to AWS opened up a world of possibilities for our application. Qovery's AI agent made the transition seamless and worry-free.",
                author: "John Smith",
                company: "DataDriven Solutions"
            }
        },
        azure: {
            steps: [
                {
                    title: "Document Render Setup",
                    description: "Catalog your Render services, noting configurations and environment variables."
                },
                {
                    title: "Create Azure Resources",
                    description: "Set up Azure App Service or Azure Kubernetes Service to host your application."
                },
                {
                    title: "Migrate Database",
                    description: "Transfer your database to Azure Database for PostgreSQL if you're using Render's PostgreSQL."
                },
                {
                    title: "Configure CI/CD",
                    description: "Implement Azure DevOps or GitHub Actions for continuous deployment, mirroring Render's setup."
                },
                {
                    title: "Finalize Migration",
                    description: "Update your application's configuration to use Azure services and deploy your code."
                }
            ],
            benefits: [
                {
                    title: "Hybrid Cloud Capabilities",
                    description: "Easily integrate with on-premises infrastructure using Azure's hybrid cloud features."
                },
                {
                    title: ".NET Integration",
                    description: "Benefit from seamless integration with .NET technologies if part of your stack."
                },
                {
                    title: "Compliance and Security",
                    description: "Leverage Azure's comprehensive compliance offerings and advanced security features."
                },
                {
                    title: "Scalable Databases",
                    description: "Take advantage of Azure's highly scalable database solutions for growing applications."
                }
            ],
            // No customer quote for this migration path
        },
    },
};

// Type for accessing migration data
export type PaaS = keyof typeof migrationData;
export type IaaS<T extends PaaS> = keyof typeof migrationData[T];