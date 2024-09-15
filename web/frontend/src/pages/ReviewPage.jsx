import React, {useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import '../index.css';
import QoveryLoader from "../components/QoveryLoader";

function ReviewPage({migrationData}) {
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState(null);
    const navigate = useNavigate();

    const API_HOST_URL = process.env.REACT_APP_API_HOST_URL || 'http://localhost:8080';

    console.log(migrationData);

    useEffect(() => {
        const generateFiles = async () => {
            setIsLoading(true);
            setError(null);
            try {
                const provider = migrationData.source.toLowerCase();
                const response = await fetch(`${API_HOST_URL}/api/migrate/${provider}`, {
                    method: 'POST', headers: {
                        'Content-Type': 'application/json',
                    }, body: JSON.stringify(migrationData),
                });

                if (!response.ok) {
                    const errorData = await response.json();
                    throw new Error(errorData.error || 'Failed to generate migration files');
                }

                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `${provider}-migration.zip`;
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
                document.body.removeChild(a);

                setIsLoading(false);
            } catch (error) {
                console.error('Error generating files:', error);
                setError(error.message || 'Failed to generate migration files. Please try again.');
                setIsLoading(false);
            }
        };
        generateFiles();
    }, [migrationData, API_HOST_URL]);

    if (isLoading) {
        return (<div className="container">
                <div className="migration-form">
                    <p className="subtitle">Generating migration files...</p>
                    <p className="info-message">This process may take a few minutes depending on the number of services
                        to migrate. Please do not reload the page.</p>
                    <QoveryLoader/>
                </div>
            </div>);
    }

    if (error) {
        return (<div className="container">
                <div className="migration-form">
                    <p className="subtitle">Error:</p>
                    <p className="error-message">{error}</p>
                    <button onClick={() => navigate('/')} className="btn btn-primary">
                        Start Over
                    </button>
                </div>
            </div>);
    }

    return (<div className="container">
            <section className="hero">
                <h2 className="subtitle">Migration Files Generated</h2>
                <div className="migration-form">
                    <p>Your migration files have been generated and downloaded as a zip archive.</p>
                    <p>Please check your downloads folder for the zip file.</p>
                    <div className="button-group">
                        <button onClick={() => navigate('/')} className="btn btn-secondary">
                            Start Over
                        </button>
                        <button onClick={() => navigate('/next-steps')} className="btn btn-primary">
                            Next Steps
                        </button>
                    </div>
                </div>
            </section>
        </div>);
}

export default ReviewPage;