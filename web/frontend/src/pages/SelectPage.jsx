import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../index.css';

// SVG icons for cloud providers
const icons = {
    heroku: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 256 400"><path fill="#430098" d="M223.371 384.038H32.629a31.978 31.978 0 0 1-31.937-31.937V31.937A31.978 31.978 0 0 1 32.629 0h190.742a31.978 31.978 0 0 1 31.937 31.937v320.164a31.978 31.978 0 0 1-31.937 31.937zM251.41 32.457c0-15.066-12.228-27.294-27.294-27.294H32.614C17.548 5.163 5.32 17.39 5.32 32.457v319.625c0 15.067 12.228 27.294 27.294 27.294H224.11c15.066 0 27.294-12.227 27.294-27.294V32.457h.006z"/><path fill="#430098" d="M160.54 347.418V276.47s7.956-29.61-94.027-29.61v100.56h41.488v-71.182s46.413-1.803 46.413 36.177v35.005h41.487v-35.005h-35.36zm-23.35-170.66v41.487h-52.54v-41.487h52.54zm-52.54-23.357h52.54V70.336c-16.565 0-52.54 16.15-52.54 83.065z"/></svg>,
    render: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="#46E3B7" d="M96 226.25v-37.5c0-12.69 10.31-23 23-23h354c12.69 0 23 10.31 23 23v37.5c0 12.69-10.31 23-23 23H119c-12.69 0-23-10.31-23-23zM96 323.25v-37.5c0-12.69 10.31-23 23-23h354c12.69 0 23 10.31 23 23v37.5c0 12.69-10.31 23-23 23H119c-12.69 0-23-10.31-23-23z"/></svg>,
    aws: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="#FF9900" d="M261.3 308.7c-25.2 0-48.2-4.1-67.6-12.3-19.5-8.2-35.9-19.6-48.7-34.2-12.8-14.6-22.8-31.8-29.8-51.4-7-19.6-10.5-41.1-10.5-64.3 0-23.2 3.5-44.7 10.5-64.3 7-19.6 17-36.8 29.8-51.4 12.8-14.6 29.3-26 49.3-34.2 20-8.2 42.7-12.3 67.9-12.3 25.2 0 48.2 4.1 67.6 12.3 19.5 8.2 35.9 19.6 48.7 34.2 12.8 14.6 22.8 31.8 29.8 51.4 7 19.6 10.5 41.1 10.5 64.3 0 23.2-3.5 44.7-10.5 64.3-7 19.6-17 36.8-29.8 51.4-12.8 14.6-29.3 26-49.3 34.2-20 8.2-42.7 12.3-67.9 12.3zm0-45.6c15.2 0 28.9-2.5 41.1-7.4 12.2-4.9 22.6-11.8 31.2-20.6 8.6-8.8 15.2-19.3 19.8-31.5 4.6-12.2 6.9-25.7 6.9-40.5 0-14.8-2.3-28.3-6.9-40.5-4.6-12.2-11.2-22.7-19.8-31.5-8.6-8.8-19-15.7-31.2-20.6-12.2-4.9-25.9-7.4-41.1-7.4-15.2 0-28.9 2.5-41.1 7.4-12.2 4.9-22.6 11.8-31.2 20.6-8.6 8.8-15.2 19.3-19.8 31.5-4.6 12.2-6.9 25.7-6.9 40.5 0 14.8 2.3 28.3 6.9 40.5 4.6 12.2 11.2 22.7 19.8 31.5 8.6 8.8 19 15.7 31.2 20.6 12.2 4.9 25.9 7.4 41.1 7.4z"/></svg>,
    gcp: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="#4285F4" d="M386.154 254.863L512 179.37v153.186l-125.846 75.494V254.863zm-127.433 75.494L384 405.85v-76.493l-125.279-74.494v75.494zm-128.567 0L256 405.85v-76.493l-125.846-74.494v75.494zM0 179.37l125.846 75.493v153.187L0 332.556V179.37zM256 0l126.154 75.493-126.154 75.494L129.846 75.493 256 0z"/></svg>,
    scaleway: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="#4F0599" d="M256 0C114.615 0 0 114.615 0 256s114.615 256 256 256 256-114.615 256-256S397.385 0 256 0zm0 448c-106.039 0-192-85.961-192-192S149.961 64 256 64s192 85.961 192 192-85.961 192-192 192z"/></svg>,
    azure: <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="#0089D6" d="M255.992 31.707L110.095 234.943l168.324 235.272h-74.345L31.789 234.944 178.615 31.707h77.377zm37.218 0l72.331 128.223L480.211 480.21H406.62l-98.476-172.732-69.278-124.915h54.344z"/></svg>
};

function SelectPage({ setMigrationData }) {
    const navigate = useNavigate();

    const handleSubmit = (e) => {
        e.preventDefault();
        const source = e.target.source.value;
        const destination = e.target.destination.value;
        setMigrationData(prev => ({ ...prev, source, destination }));
        navigate('/credentials');
    };

    return (
        <div className="container">
            <section className="hero">
                <p className="subtitle">Seamlessly migrate your infrastructure with Qovery's AI-powered platform in 30 Seconds</p>

                <form onSubmit={handleSubmit} className="migration-form">
                    <h2 className="section-title">Choose Your Migration Source and Destination</h2>

                    <div className="form-group">
                        <label htmlFor="source">Source:</label>
                        <select name="source" id="source" className="select-input">
                            <option value="heroku">{icons.heroku} Heroku</option>
                            <option value="render">{icons.render} Render</option>
                        </select>
                    </div>
                    <div className="form-group">
                        <label htmlFor="destination">Destination:</label>
                        <select name="destination" id="destination" className="select-input">
                            <option value="aws">{icons.aws} AWS</option>
                            <option value="gcp">{icons.gcp} GCP</option>
                            <option value="scaleway">{icons.scaleway} Scaleway</option>
                            <option value="azure">{icons.azure} Azure</option>
                        </select>
                    </div>
                    <button type="submit" className="btn btn-primary">Start Migration</button>
                </form>
            </section>
        </div>
    );
}

export default SelectPage;