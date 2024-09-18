import React, {useEffect} from 'react';

const ScriptInjector = () => {
    useEffect(() => {
        // Filter environment variables that end with _JS_SCRIPT
        const scriptEnvVars = Object.keys(process.env).filter(key =>
            key.endsWith('_JS_SCRIPT')
        );

        // Inject each script
        scriptEnvVars.forEach(key => {
            const scriptContent = process.env[key];
            if (scriptContent) {
                const script = document.createElement('script');
                script.innerHTML = scriptContent;
                document.head.appendChild(script);
            }
        });
    }, []);

    return null;
};

export default ScriptInjector;