'use client';

import React, { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';

const VideoSection = () => {
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const timer = setTimeout(() => {
            setIsLoading(false);
        }, 1500); // Simulating a 1.5-second load time

        return () => clearTimeout(timer);
    }, []);

    return (
        <section className="w-full pb-12 md:pb-12 lg:pb-24 px-4">
            <div className="container mx-auto">
                <div className="relative aspect-video rounded-lg overflow-hidden shadow-2xl">
                    {isLoading ? (
                        <div className="absolute inset-0 flex items-center justify-center bg-violet-200">
                            <Loader2 className="w-12 h-12 text-violet-600 animate-spin" />
                        </div>
                    ) : (
                        <div style={{position: 'relative', paddingBottom: '56.25%', height: 0}}>
                            <iframe
                                src="https://www.loom.com/embed/0045d92738f0445aac1cd01766dbbdee"
                                frameBorder="0"
                                allowFullScreen={true}
                                style={{position: 'absolute', top: 0, left: 0, width: '100%', height: '100%'}}
                                allow="autoplay; fullscreen; picture-in-picture"
                                data-loom-playback-speed-plugin="true"
                            />
                        </div>
                    )}
                </div>
            </div>
        </section>
    );
};

export default VideoSection;