import {Button} from "@/app/components/ui/button"
import {ArrowRight, Bot, CheckCircle, Cloud, Github, Shield, Zap} from "lucide-react"
import Link from "next/link"

export default function LandingPage() {
    return (
        <div className="flex flex-col min-h-screen bg-gradient-to-b from-[#1e1e3f] to-[#2d2d5f] text-white">
            <main className="flex-1">
                <section className="w-full py-12 md:py-24 lg:py-32 xl:py-32 px-4">
                    <div className="container mx-auto text-center">
                        <div className="inline-block bg-purple-800 bg-opacity-50 rounded-full px-4 py-2 mb-6">
                            <span
                                className="text-sm font-medium">Qovery AI Migration Agent - Simplify Your Migration</span>
                        </div>
                        <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl lg:text-7xl mb-6">
                            AI-Powered Migration
                        </h1>
                        <p className="mx-auto max-w-[700px] text-gray-300 text-lg md:text-xl mb-8">
                            Qovery AI Migration Agent eliminates your migration headaches. Migrate and maintain a secure
                            and compliant infrastructure in hours - not months!
                        </p>
                        <div className="flex flex-col sm:flex-row justify-center gap-4">
                            <Link href="/get-started" passHref>
                                <Button
                                    className="bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 px-6 rounded-full transition-colors">
                                    Get Started
                                </Button>
                            </Link>
                            <Link href="https://github.com/Qovery/qovery-migration-ai-agent">
                            <Button variant="outline"
                                    className="border-white text-white hover:bg-white hover:text-[#1e1e3f] font-medium py-2 px-6 rounded-full transition-colors flex items-center justify-center">
                                <Github className="mr-2 h-5 w-5"/>
                                Give a Star
                            </Button>
                            </Link>
                        </div>
                        <div className="text-sm text-gray-400 mt-4">
                            Support our open-source project and stay updated!
                        </div>
                    </div>
                </section>

                <section className="w-full pb-12 md:pb-12 lg:pb-24 px-4">
                    <div className="container mx-auto">
                        <div className="relative aspect-video rounded-lg overflow-hidden shadow-2xl">
                            <div style={{position: 'relative', paddingBottom: '56.25%', height: 0}}>
                                <iframe
                                    src="https://www.loom.com/embed/0045d92738f0445aac1cd01766dbbdee"
                                    frameBorder="0"
                                    webkitallowfullscreen="true"
                                    mozallowfullscreen="true"
                                    allowFullScreen={true}
                                    style={{position: 'absolute', top: 0, left: 0, width: '100%', height: '100%'}}
                                ></iframe>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="features" className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
                    <div className="container mx-auto">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl text-center mb-12">
                            Key Features
                        </h2>
                        <div className="grid gap-8 md:grid-cols-3">
                            <div className="flex flex-col items-center text-center">
                                <Cloud className="h-12 w-12 mb-4 text-purple-400"/>
                                <h3 className="text-xl font-bold mb-2">Cross-Platform Compatibility</h3>
                                <p className="text-gray-300">
                                    Effortlessly migrate AI models between major cloud providers and frameworks.
                                </p>
                            </div>
                            <div className="flex flex-col items-center text-center">
                                <Zap className="h-12 w-12 mb-4 text-purple-400"/>
                                <h3 className="text-xl font-bold mb-2">Automated Optimization</h3>
                                <p className="text-gray-300">
                                    Our agent automatically optimizes your models for the target platform, ensuring peak
                                    performance.
                                </p>
                            </div>
                            <div className="flex flex-col items-center text-center">
                                <Bot className="h-12 w-12 mb-4 text-purple-400"/>
                                <h3 className="text-xl font-bold mb-2">Intelligent Assistance</h3>
                                <p className="text-gray-300">
                                    Get AI-powered recommendations and troubleshooting throughout the migration process.
                                </p>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="how-it-works" className="w-full py-12 md:py-24 lg:py-32 px-4">
                    <div className="container mx-auto">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl text-center mb-12">
                            How It Works
                        </h2>
                        <div className="max-w-3xl mx-auto">
                            <ol className="relative border-l border-gray-700 space-y-10">
                                {[
                                    {
                                        title: "Input Source Platform Details",
                                        description: "Provide details about your current Heroku/Render deployment"
                                    },
                                    {
                                        title: "Select Destination Platform",
                                        description: "Choose your target platform (AWS/GCP/Azure/Scaleway)"
                                    },
                                    {
                                        title: "Analyze Application Structure",
                                        description: "Our AI agent examines your application's architecture"
                                    },
                                    {
                                        title: "Generate Configurations",
                                        description: "Automated creation of Terraform and Dockerfile configurations"
                                    },
                                    {
                                        title: "Estimate Migration Costs",
                                        description: "Get an overview of potential costs for your migration"
                                    },
                                    {
                                        title: "Download and Review",
                                        description: "Access and verify the generated migration files"
                                    }
                                ].map((step, index) => (
                                    <li key={index} className="ml-6">
                                        <div
                                            className="absolute w-3 h-3 bg-purple-600 rounded-full mt-1.5 -left-1.5 border border-gray-900"></div>
                                        <h3 className="text-xl font-semibold text-purple-400">{step.title}</h3>
                                        <p className="mb-4 text-gray-400">{step.description}</p>
                                    </li>
                                ))}
                            </ol>
                        </div>
                    </div>
                </section>

                <section id="security" className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
                    <div className="container mx-auto">
                        <div className="flex flex-col items-center space-y-6 text-center">
                            <Shield className="h-20 w-20 text-purple-400"/>
                            <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl">Security
                                First</h2>
                            <p className="max-w-[600px] text-gray-300 text-lg">
                                We prioritize the security of your data and infrastructure
                            </p>
                            <ul className="space-y-4 text-left">
                                {[
                                    "Our code is fully auditable and open-source",
                                    "We do not store any credentials",
                                    "All migration operations are performed locally on your machine"
                                ].map((item, index) => (
                                    <li key={index} className="flex items-center">
                                        <CheckCircle className="mr-3 h-6 w-6 text-purple-400 flex-shrink-0"/>
                                        <span className="text-gray-300">{item}</span>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    </div>
                </section>

                <section className="w-full py-12 md:py-24 lg:py-32 px-4">
                    <div className="container mx-auto text-center">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl mb-6">
                            Ready to Migrate in 5 Minutes?
                        </h2>
                        <p className="mx-auto max-w-[600px] text-gray-300 text-lg mb-8">
                            Start your journey to seamless application migration with Qovery AI Migration Agent.
                        </p>
                        <Link href="/get-started" passHref>
                            <Button
                                className="bg-purple-600 hover:bg-purple-700 text-white font-medium py-3 px-8 rounded-full transition-colors text-lg"
                            >
                                Get Started Now
                                <ArrowRight className="ml-2 h-5 w-5"/>
                            </Button>
                        </Link>
                    </div>
                </section>

                <section className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
                    <div className="container mx-auto text-center">
                        <h2 className="text-2xl font-semibold mb-8">Trusted by 200+ Organizations Worldwide</h2>
                        <div className="flex flex-wrap justify-center items-center gap-8">
                            {['Eurovision', 'Elevo', 'KirkpatrickPrice', 'Common', 'FlowBank'].map((company) => (
                                <div key={company}
                                     className="w-32 h-12 bg-gray-800 rounded flex items-center justify-center">
                                    <span className="text-gray-300 font-medium">{company}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>
            </main>
        </div>
    )
}