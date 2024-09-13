import React, { useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import HomePage from './pages/HomePage';
import CredentialsPage from './pages/CredentialsPage';
import ReviewPage from './pages/ReviewPage';
import NextStepsPage from './pages/NextStepsPage';
import SecurityPage from './pages/SecurityPage';
import Footer from "./components/Footer";

function App() {
    const [migrationData, setMigrationData] = useState({
        source: '', destination: '', credentials: {},
    });

    return (
        <Router>
            <div className="App">
                <header className="site-header">
                    <div className="container">
                        <div className="header-content">
                            <h1 className="site-title">Qovery Migration</h1>
                            <nav className="site-nav">
                                <button className="btn btn-secondary">Contact</button>
                                <button className="btn btn-primary">Sign Up</button>
                            </nav>
                        </div>
                    </div>
                </header>
                <main className="main-content">
                    <Routes>
                        <Route
                            path="/"
                            element={<HomePage setMigrationData={setMigrationData}/>}
                        />
                        <Route
                            path="/credentials"
                            element={<CredentialsPage migrationData={migrationData} setMigrationData={setMigrationData}/>}
                        />
                        <Route
                            path="/review"
                            element={<ReviewPage migrationData={migrationData}/>}
                        />
                        <Route
                            path="/next-steps"
                            element={<NextStepsPage/>}
                        />
                        <Route
                            path="/security"
                            element={<SecurityPage/>}
                        />
                    </Routes>
                </main>
                <Footer />
            </div>
        </Router>
    );
}

export default App;