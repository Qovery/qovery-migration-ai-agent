import {Button} from "@/app/components/ui/button"
import {ArrowRight, Bot, Cloud, Github, Shield, Zap} from "lucide-react"
import Link from "next/link"
import VideoSection from "@/app/components/VideoSection";

export default function LandingPage() {
    return (
        <div className="flex flex-col min-h-screen bg-gradient-to-b from-[#1e1e3f] to-[#2d2d5f] text-white">
            <main className="flex-1">
                <section className="w-full py-12 md:py-24 lg:py-32 xl:py-32 px-4">
                    <div className="container mx-auto text-center">
                        <div className="inline-block bg-violet-800 bg-opacity-50 rounded-full px-4 py-2 mb-6">
                            <span
                                className="text-sm font-medium">Qovery AI Cloud Migration Agent - Simplify Your Migration</span>
                        </div>
                        <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl lg:text-7xl mb-6">
                            AI-Powered Cloud Migration
                        </h1>
                        <p className="mx-auto max-w-[700px] text-gray-300 text-lg md:text-xl mb-8">
                            Qovery AI Cloud Migration Agent eliminates your cloud migration headaches. Migrate and
                            maintain a secure
                            and compliant cloud infrastructure in hours - not months!
                        </p>
                        <div className="flex flex-col sm:flex-row justify-center gap-4">
                            <Link href="/get-started" passHref>
                                <Button
                                    className="bg-violet-600 hover:bg-violet-700 text-white font-medium py-2 px-6 rounded-full transition-colors">
                                    Get Started (No Signup)
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

                <VideoSection/>

                <section className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
                    <div className="container mx-auto">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl text-center mb-12">
                            Seamless Migration Process
                        </h2>
                        <div className="flex flex-col md:flex-row items-center justify-between gap-8">
                            <div className="w-full md:w-1/4">
                                <h3 className="text-xl font-semibold mb-4 text-center">Compatible PaaS</h3>
                                <ul className="space-y-2 text-center">
                                    {['Heroku', 'Render', 'Vercel', 'Netlify', 'DigitalOcean App Platform'].map((platform) => (
                                        <li key={platform} className="bg-gray-800 rounded-lg py-2 px-4">{platform}</li>
                                    ))}
                                </ul>
                            </div>
                            <ArrowRight className="hidden md:block h-12 w-12 text-violet-400"/>
                            <div className="w-full md:w-2/4 flex flex-col items-center">
                                <div className="bg-violet-600 rounded-full p-6 mb-4">
                                    <Bot className="h-16 w-16 text-white"/>
                                </div>
                                <h3 className="text-2xl font-semibold mb-2 text-center">Qovery AI Cloud Migration
                                    Agent</h3>
                                <p className="text-center text-gray-300 mb-4">
                                    Automates migration by generating Terraform files and leveraging Qovery's platform
                                </p>
                                <ul className="space-y-2 text-center">
                                    <li className="bg-gray-800 rounded-lg py-2 px-4">Analyzes source infrastructure</li>
                                    <li className="bg-gray-800 rounded-lg py-2 px-4">Generates Terraform
                                        configurations
                                    </li>
                                    <li className="bg-gray-800 rounded-lg py-2 px-4">Optimizes for target platform</li>
                                </ul>
                            </div>
                            <ArrowRight className="hidden md:block h-12 w-12 text-violet-400"/>
                            <div className="w-full md:w-1/4">
                                <h3 className="text-xl font-semibold mb-4 text-center">Target IaaS</h3>
                                <ul className="space-y-2 text-center">
                                    {['AWS', 'Google Cloud Platform', 'Microsoft Azure', 'Scaleway'].map((platform) => (
                                        <li key={platform} className="bg-gray-800 rounded-lg py-2 px-4">{platform}</li>
                                    ))}
                                    <li key="kubernetes" className="bg-blue-800 rounded-lg py-2 px-4">Kubernetes</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="features" className="w-full py-12 md:py-24 lg:py-32 px-4">
                    <div className="container mx-auto">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl text-center mb-12">
                            Key Features
                        </h2>
                        <div className="grid gap-8 md:grid-cols-4">
                            <div className="flex flex-col items-center text-center">
                                <Cloud className="h-12 w-12 mb-4 text-violet-400"/>
                                <h3 className="text-xl font-bold mb-2">Cross-Platform Compatibility</h3>
                                <p className="text-gray-300">
                                    Effortlessly migrate AI models between major cloud providers and frameworks.
                                </p>
                            </div>
                            <div className="flex flex-col items-center text-center">
                                <Zap className="h-12 w-12 mb-4 text-violet-400"/>
                                <h3 className="text-xl font-bold mb-2">Automated Optimization</h3>
                                <p className="text-gray-300">
                                    Our agent automatically optimizes your models for the target platform, ensuring peak
                                    performance.
                                </p>
                            </div>
                            <div className="flex flex-col items-center text-center">
                                <Bot className="h-12 w-12 mb-4 text-violet-400"/>
                                <h3 className="text-xl font-bold mb-2">Intelligent Assistance</h3>
                                <p className="text-gray-300">
                                    Get AI-powered recommendations and troubleshooting throughout the migration process.
                                </p>
                            </div>
                            <div className="flex flex-col items-center text-center">
                                <Shield className="h-12 w-12 mb-4 text-violet-400"/>
                                <h3 className="text-xl font-bold mb-2">Security First</h3>
                                <p className="text-gray-300">
                                    We prioritize the security of your data and infrastructure with open-source,
                                    credential-free, and local operations.
                                </p>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="how-it-works" className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
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
                                            className="absolute w-3 h-3 bg-violet-600 rounded-full mt-1.5 -left-1.5 border border-gray-900"></div>
                                        <h3 className="text-xl font-semibold text-violet-400">{step.title}</h3>
                                        <p className="mb-4 text-gray-400">{step.description}</p>
                                    </li>
                                ))}
                            </ol>
                        </div>
                    </div>
                </section>

                <section className="w-full py-12 md:py-24 lg:py-32 px-4">
                    <div className="container mx-auto text-center">
                        <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl mb-6">
                            Ready to Migrate in 5 Minutes?
                        </h2>
                        <p className="mx-auto max-w-[600px] text-gray-300 text-lg mb-8">
                            Start your journey to seamless application migration with Qovery AI Cloud Migration Agent.
                        </p>
                        <Link href="/get-started" passHref>
                            <Button
                                className="bg-violet-600 hover:bg-violet-700 text-white font-medium py-3 px-8 rounded-full transition-colors text-lg">
                                Get Started Now
                                <ArrowRight className="ml-2 h-5 w-5"/>
                            </Button>
                        </Link>
                    </div>
                </section>

                <section className="w-full py-12 md:py-24 lg:py-32 px-4 bg-gray-900">
                    <div className="container mx-auto text-center">
                        <h2 className="text-2xl font-semibold mb-8">Qovery is Trusted by 200+ Organizations
                            Worldwide</h2>
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