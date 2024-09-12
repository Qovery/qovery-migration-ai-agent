import '../index.css';  // Import the updated CSS file here

function NextStepsPage() {
    return (<div>
            <h2 className="text-xl font-bold mb-4">Next Steps</h2>
            <ol className="list-decimal pl-6">
                <li className="mb-2">Download the generated Terraform manifest and Dockerfile.</li>
                <li className="mb-2">Install Terraform on your local machine if you haven't already.</li>
                <li className="mb-2">Open a terminal and navigate to the directory containing the downloaded files.</li>
                <li className="mb-2">Run `terraform init` to initialize the Terraform working directory.</li>
                <li className="mb-2">Run `terraform plan` to see the execution plan.</li>
                <li className="mb-2">If the plan looks good, run `terraform apply` to create the infrastructure.</li>
                <li className="mb-2">Use the Dockerfile to build and deploy your application to the new
                    infrastructure.
                </li>
            </ol>
            <p className="mt-4">For more detailed instructions or if you encounter any issues, please refer to our
                documentation or contact our support team.</p>
        </div>);
}

export default NextStepsPage;