import React, {useEffect} from 'react';
import {Link} from 'react-router-dom';
import mermaid from 'mermaid';
import './LandingPage.css';
import '../mermaid.css';

const LandingPage = () => {
    useEffect(() => {
        mermaid.initialize({startOnLoad: true, theme: 'dark'});
    }, []);

    return (<div className="landing-page">
        <div className="hero">
            <h1>Qovery AI Migration Agent</h1>
            <p className="subtitle">
                Simplify your application migration process with Qovery AI Migration Agent. Our tool is designed to
                facilitate the migration of applications from various platforms to AWS, GCP, Scaleway and Azure via
                &nbsp;<a href="https://www.qovery.com" target="_blank" rel="noopener noreferrer"
                         className="inline-link">Qovery</a>.
            </p>
            <div className="hero-buttons">
                <Link to="/select" className="cta-button">Get Started</Link>
                <a href="https://github.com/Qovery/qovery-migration-ai-agent"
                   className="github-button"
                   target="_blank"
                   rel="noopener noreferrer">
                    View on GitHub
                </a>
            </div>
        </div>

        <section className="why-choose">
            <h2>Why Choose Qovery Migration Agent?</h2>
            <p>
                Migrating applications from one platform to another can be a time-consuming, error-prone process and
                extremely costly. Qovery Migration Agent aims to simplify this process by automating the generation
                of Terraform configurations, Dockerfiles, and other necessary files for deploying applications on
                Qovery.
            </p>
        </section>

        <section className="features">
            <h2>Features</h2>
            <ul>
                <li>Generate Terraform files to migrate fullstack apps from Heroku to AWS, GCP, Azure, or Scaleway
                    via Qovery
                </li>
                <li>Estimate migration costs</li>
                <li>Automate the creation of Dockerfiles and other necessary configuration files</li>
                <li>Provide step-by-step guidance throughout the migration process</li>
            </ul>
        </section>

        <section className="how-it-works">
            <h2>How It Works</h2>
            <div>
                <div className="mermaid">
                    {`
                        graph TD
                            A[Start] --> B[Input Source Platform Details Heroku/Render]
                            B --> C[Select Destination Platform AWS/GCP/Azure/Scaleway]
                            C --> D[Analyze Application Structure]
                            D --> E[Generate Terraform and Dockerfile Configurations]
                            E --> F[Estimate Migration Costs]
                            F --> G[Generate Zip File with Migration Files]
                            G --> H[You Can Download and Review the Migration Files]
                            H --> I[End]
                        `}
                </div>
                <p className="diagram-explanation">
                    This diagram illustrates the step-by-step process of migrating your application using Qovery AI
                    Migration Agent. From inputting your source platform details to verifying the deployment, our
                    tool guides you
                    through each stage, ensuring a smooth and efficient migration experience.
                </p>
            </div>
        </section>

        <section className="security">
            <h2>Security</h2>
            <div className="security-note">
                <p>We prioritize the security of your data and infrastructure:</p>
                <ul>
                    <li>Our code is fully auditable and open-source</li>
                    <li>We do not store any credentials</li>
                    <li>All migration operations are performed locally on your machine - not on the server</li>
                </ul>
            </div>
        </section>

        <section className="cta">
            <h2>Ready to Simplify Your Migration?</h2>
            <Link to="/select" className="cta-button">Get Started</Link>
        </section>
    </div>);
};

export default LandingPage;