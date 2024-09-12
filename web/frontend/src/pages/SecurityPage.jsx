import '../index.css'; // Import the updated CSS file here

function SecurityPage() {
    return (<div>
        <h2 className="text-xl font-bold mb-4">Security Information</h2>
        <p className="mb-4">At Qovery, we take the security and privacy of your data seriously. Here's what you need
            to know about our Migration AI Agent:</p>
        <ul className="list-disc pl-6">
            <li className="mb-2">Your credentials are never stored on our servers. They are only used temporarily to
                generate the migration files.
            </li>
            <li className="mb-2">All communication between your browser and our servers is encrypted using HTTPS.
            </li>
            <li className="mb-2">The Migration AI Agent is fully open-source. You can audit the code or self-host
                the application if you prefer.
            </li>
            <li className="mb-2">We regularly update our dependencies and conduct security audits to ensure the
                safety of the application.
            </li>
        </ul>
        <p className="mt-4">If you have any security concerns or questions, please don't hesitate to contact our
            security team at security@qovery.com.</p>
    </div>);
}

export default SecurityPage;