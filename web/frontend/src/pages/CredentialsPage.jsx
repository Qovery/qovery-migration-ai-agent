import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../index.css';

function CredentialsPage({ migrationData, setMigrationData }) {
    const navigate = useNavigate();

    const handleSubmit = (e) => {
        e.preventDefault();
        const data = {};

        const formData = new FormData(e.target);
        for (let [key, value] of formData.entries()) {
            data[key] = value;
        }

        setMigrationData(prev => ({ ...prev, ...data }));
        navigate('/review');
    };

    const renderInputs = () => {
        switch (migrationData.source) {
            case 'heroku':
                return (
                    <div className="form-group">
                        <label htmlFor="herokuApiKey">Heroku API Key:</label>
                        <input type="password" id="herokuApiKey" name="herokuApiKey" className="input-field" required />
                    </div>
                );
            default:
                return null;
        }
    };

    return (
        <div className="container">
            <p className="subtitle">Provide the necessary credentials for your {migrationData.source} to {migrationData.destination} migration.</p>

            <form onSubmit={handleSubmit} className="migration-form">
                <h2 className="section-title">Source: {migrationData.source}</h2>
                {renderInputs()}

                <p className="security-note">
                    🔒 Your credentials are securely processed and not stored. They are only used for this migration.
                </p>

                <div className="button-group">
                    <button type="button" className="btn btn-secondary" onClick={() => navigate('/')}>Back</button>
                    <button type="submit" className="btn btn-primary">Next</button>
                </div>
            </form>
        </div>
    );
}

export default CredentialsPage;