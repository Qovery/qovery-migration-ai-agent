import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '../index.css';

function ReviewPage({ migrationData }) {
    const [files, setFiles] = useState({ terraform: '', dockerfile: '' });
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        const generateFiles = async () => {
            setIsLoading(true);
            setError(null);
            try {
                const response = await fetch('http://localhost:8080/api/migrate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(migrationData),
                });
                if (!response.ok) throw new Error('Failed to generate migration files');
                setFiles({
                    terraform: await fetchFile('terraform.tf'),
                    dockerfile: await fetchFile('Dockerfile'),
                });
            } catch (error) {
                console.error('Error generating files:', error);
                setError('Failed to generate migration files. Please try again.');
            } finally {
                setIsLoading(false);
            }
        };
        generateFiles();
    }, [migrationData]);

    const fetchFile = async (filename) => {
        const response = await fetch(`http://localhost:8080/api/download/${filename}`);
        return await response.text();
    };

    const handleDownload = (content, filename) => {
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
    };

    if (isLoading) {
        return (
            <div className="container">
                <div className="migration-form">
                    <p className="subtitle">Generating migration files...</p>
                    {/* You can add a loading spinner here */}
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="container">
                <div className="migration-form">
                    <p className="subtitle">{error}</p>
                    <button onClick={() => navigate('/')} className="btn btn-primary">
                        Start Over
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="container">
            <section className="hero">
                <h2 className="subtitle">Review Generated Files</h2>
                <div className="migration-form">
                    <div className="form-group">
                        <h3 className="section-title">Terraform Manifest:</h3>
                        <pre className="code-preview">{files.terraform}</pre>
                        <button onClick={() => handleDownload(files.terraform, 'terraform.tf')} className="btn btn-secondary">
                            Download Terraform
                        </button>
                    </div>
                    <div className="form-group">
                        <h3 className="section-title">Dockerfile:</h3>
                        <pre className="code-preview">{files.dockerfile}</pre>
                        <button onClick={() => handleDownload(files.dockerfile, 'Dockerfile')} className="btn btn-secondary">
                            Download Dockerfile
                        </button>
                    </div>
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
        </div>
    );
}

export default ReviewPage;