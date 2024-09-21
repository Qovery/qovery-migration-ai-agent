import React from 'react';

export default function VideoSection() {
    return (
        <section className="w-full pb-12 md:pb-12 lg:pb-24 px-4">
            <div className="container mx-auto">
                <div className="relative aspect-video rounded-lg overflow-hidden shadow-2xl">
                    <div style={{position: 'relative', paddingBottom: '56.25%', height: 0}}>
                        <iframe
                            src="https://www.loom.com/embed/0045d92738f0445aac1cd01766dbbdee"
                            frameBorder="0"
                            allowFullScreen={true}
                            style={{position: 'absolute', top: 0, left: 0, width: '100%', height: '100%'}}
                            {...{
                                'allow': 'autoplay; fullscreen; picture-in-picture',
                                'data-loom-playback-speed-plugin': 'true',
                            } as React.IframeHTMLAttributes<HTMLIFrameElement>}
                        ></iframe>
                    </div>
                </div>
            </div>
        </section>
    );
}