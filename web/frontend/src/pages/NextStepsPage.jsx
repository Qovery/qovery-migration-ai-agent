import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../index.css';

function NextStepsPage() {
    const navigate = useNavigate();

    return (
        <div className="container">
            <section className="hero">
                <h2 className="subtitle left-aligned">Next Steps for Your Migration</h2>
                <div className="migration-form">
                    <ol className="next-steps-list">
                        <li>Install <a href="https://developer.hashicorp.com/terraform/install" className="inline-link">Terraform</a> on your local machine if you haven't already.</li>
                        <li>Open a terminal and navigate to the directory containing the downloaded files.</li>
                        <li>Run <code>terraform init</code> to initialize the Terraform working directory.</li>
                        <li>Run <code>terraform plan</code> to see the execution plan.</li>
                        <li>If the plan looks good, run <code>terraform apply</code> to create the infrastructure.</li>
                        <li>Use the Dockerfile to build and deploy your application to the new infrastructure.</li>
                    </ol>
                    <p className="additional-info">
                        For more detailed instructions or if you encounter any issues, please refer to our
                        <a href="https://hub.qovery.com" className="inline-link"> documentation</a> or
                        <a href="https://discuss.qovery.com" className="inline-link"> contact our support team</a>.
                    </p>
                    <div className="button-group">
                        <button onClick={() => navigate('/review')} className="btn btn-secondary">Back to Review</button>
                        <button onClick={() => navigate('/')} className="btn btn-primary">Start New Migration</button>
                    </div>
                </div>
            </section>
        </div>
    );
}

export default NextStepsPage;