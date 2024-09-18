import React, {useState} from 'react';
import {BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import LandingPage from './pages/LandingPage'; // Add this import
import SelectPage from './pages/SelectPage';
import CredentialsPage from './pages/CredentialsPage';
import ReviewPage from './pages/ReviewPage';
import NextStepsPage from './pages/NextStepsPage';
import SecurityPage from './pages/SecurityPage';
import Footer from "./components/Footer";
import ScriptInjector from "./components/ScriptInjector";

function App() {
    const [migrationData, setMigrationData] = useState({
        source: '', destination: '',
    });

    return (<Router>
        <div className="App">
            <ScriptInjector/>
            <header className="site-header">
                <div className="container">
                    <div className="header-content">
                        <h1 className="site-title">Qovery Migration</h1>
                        <nav className="site-nav">
                            <a href="https://github.com/Qovery/qovery-migration-ai-agent"
                               className="btn btn-secondary"
                               target="_blank"
                               rel="noopener noreferrer">
                                GitHub
                            </a>
                            <a href="https://console.qovery.com"
                               className="btn btn-primary"
                               target="_blank"
                               rel="noopener noreferrer">
                                Sign Up
                            </a>
                        </nav>
                    </div>
                </div>
            </header>
            <main className="main-content">
                <Routes>
                    <Route
                        path="/"
                        element={<LandingPage/>}
                    />
                    <Route
                        path="/select"
                        element={<SelectPage setMigrationData={setMigrationData}/>}
                    />
                    <Route
                        path="/credentials"
                        element={<CredentialsPage migrationData={migrationData}
                                                  setMigrationData={setMigrationData}/>}
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
            <Footer/>
        </div>
    </Router>);
}

export default App;