import React from 'react';
import Link from 'next/link';
import {Button} from "@/app/components/ui/button";
import {ArrowRight, CheckCircle, Quote} from "lucide-react";
import {migrationData, PaaS, IaaS} from '@/app/data/migrationData';
import Header from "@/app/components/Header";
import Footer from "@/app/components/Footer";

export default function EnhancedMigrationPage({
                                                  params,
                                              }: {
    params: { paas: PaaS; iaas: IaaS<PaaS> }
}) {
    const {paas, iaas} = params;
    const migrationInfo = migrationData[paas]?.[iaas];

    return (
        <div className="flex flex-col min-h-screen bg-gradient-to-b from-[#1e1e3f] to-[#2d2d5f] text-white">
            <Header/>
            <main className="flex-1">
                <section className="w-full py-12 sm:py-16 md:py-24 px-4 text-center">
                    <div className="flex justify-center items-center space-x-4 mb-6">
                        <img src={`/images/${paas}-logo.png`} alt={`${paas} logo`} className="w-16 h-16"/>
                        <ArrowRight className="w-8 h-8 text-violet-400"/>
                        <img src={`/images/${iaas}-logo.png`} alt={`${iaas} logo`} className="w-16 h-16"/>
                    </div>
                    <h1 className="text-4xl sm:text-5xl md:text-6xl font-bold tracking-tighter mb-4">
                        Migrate from {paas} to {iaas}
                    </h1>
                    <p className="text-gray-300 text-sm sm:text-base md:text-lg mb-8 max-w-[90%] sm:max-w-[70%] mx-auto">
                        Let Qovery AI Cloud Migration Agent guide you through migrating your application
                        from {paas} to {iaas}. Our intelligent system will optimize your infrastructure
                        for the best performance and cost-efficiency.
                    </p>
                    <Link href="/get-started" passHref>
                        <Button
                            className="bg-violet-600 hover:bg-violet-700 text-white text-lg px-6 py-3 rounded-full transition-all duration-300 transform hover:scale-105">
                            Start Migration
                            <ArrowRight className="ml-2 h-5 w-5"/>
                        </Button>
                    </Link>
                </section>

                <section className="w-full py-16 sm:py-20 md:py-24 px-4 bg-gray-900">
                    <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tighter text-center mb-12">
                        {paas} to {iaas} Migration Steps
                    </h2>
                    <div className="max-w-4xl mx-auto">
                        {migrationInfo?.steps.map((step, index) => (
                            <div key={index} className="flex items-start mb-8">
                                <div className="flex-shrink-0 mr-4">
                                    <div
                                        className="w-10 h-10 bg-violet-600 rounded-full flex items-center justify-center text-white font-bold">
                                        {index + 1}
                                    </div>
                                </div>
                                <div>
                                    <h3 className="text-xl font-semibold mb-2">{step.title}</h3>
                                    <p className="text-gray-300">{step.description}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                </section>

                <section className="w-full py-16 sm:py-20 md:py-24 px-4">
                    <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tighter text-center mb-12">
                        Benefits of Migrating to {iaas}
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
                        {migrationInfo?.benefits.map((benefit, index) => (
                            <div key={index} className="bg-gray-800 rounded-lg p-6 shadow-lg">
                                <CheckCircle className="w-8 h-8 text-green-400 mb-4"/>
                                <h3 className="text-xl font-semibold mb-2">{benefit.title}</h3>
                                <p className="text-gray-300">{benefit.description}</p>
                            </div>
                        ))}
                    </div>
                </section>

                {migrationInfo?.customerQuote && (
                    <section className="w-full py-16 sm:py-20 md:py-24 px-4 bg-gray-800">
                        <div className="max-w-4xl mx-auto text-center">
                            <Quote className="w-16 h-16 text-violet-400 mx-auto mb-6"/>
                            <blockquote className="text-2xl font-light italic mb-6">
                                "{migrationInfo.customerQuote.text}"
                            </blockquote>
                            <div className="font-semibold">
                                {migrationInfo.customerQuote.author}
                            </div>
                            <div className="text-violet-400">
                                {migrationInfo.customerQuote.company}
                            </div>
                        </div>
                    </section>
                )}
            </main>
            <Footer/>
        </div>
    );
}