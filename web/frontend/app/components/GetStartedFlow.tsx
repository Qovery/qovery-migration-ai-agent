import {useEffect, useState} from "react"
import {Button} from "@/app/components/ui/button"
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/app/components/ui/select"
import {Input} from "@/app/components/ui/input"
import {Alert, AlertDescription, AlertTitle} from "@/app/components/ui/alert"
import {Loader2} from "lucide-react"
import {migratePaas} from "@/app/lib/api"
import Link from 'next/link'
import {
    SiFlyway,
    SiGooglecloud,
    SiHeroku,
    SiKubernetes,
    SiNetlify,
    SiPlatformdotsh,
    SiRailway,
    SiRender,
    SiScaleway,
    SiVercel
} from "react-icons/si";
import {Progress} from "@/app/components/ui/progress"
import {FaAws} from "react-icons/fa";
import {VscAzure} from "react-icons/vsc"; // Assuming you've saved the SVG components in a file named PlatformIcons.js

const paasOptions = [
    {value: "heroku", label: "Heroku", icon: <SiHeroku size={24}></SiHeroku>},
    {value: "render", label: "Render (coming soon)", disabled: true, icon: <SiRender size={24}></SiRender>},
    {value: "railway", label: "Railway (coming soon)", disabled: true, icon: <SiRailway size={24}></SiRailway>},
    {value: "fly", label: "Fly (coming soon)", disabled: true, icon: <SiFlyway size={24}></SiFlyway>},
    {value: "vercel", label: "Vercel (coming soon)", disabled: true, icon: <SiVercel size={24}></SiVercel>},
    {value: "netlify", label: "Netlify (coming soon)", disabled: true, icon: <SiNetlify size={24}></SiNetlify>},
    {
        value: "platformsh",
        label: "Platform.sh (coming soon)",
        disabled: true,
        icon: <SiPlatformdotsh size={24}></SiPlatformdotsh>
    },
]

const cloudOptions = [
    {value: "aws", label: "AWS", icon: <FaAws size={24}></FaAws>},
    {value: "gcp", label: "GCP", icon: <SiGooglecloud size={24}></SiGooglecloud>},
    {value: "azure", label: "Azure", icon: <VscAzure size={24}></VscAzure>},
    {value: "scaleway", label: "Scaleway", icon: <SiScaleway size={24}></SiScaleway>},
    {value: "kubernetes", label: "Kubernetes", icon: <SiKubernetes size={24}></SiKubernetes>},
]

const migrationSteps = [
    "Retrieving apps configuration details",
    "Extracting important information",
    "Generating Terraform files 1/2",
    "Generating Terraform files 2/2",
    "Validating Terraform files",
    "Generating Dockerfiles",
    "Validating Dockerfile files",
    "Estimating overall costs",
    "Creating zip archive"
]

export default function GetStartedFlow() {
    const [step, setStep] = useState(1)
    const [selectedPaas, setSelectedPaas] = useState(paasOptions[0].value)
    const [selectedCloud, setSelectedCloud] = useState(cloudOptions[0].value)
    const [herokuApiKey, setHerokuApiKey] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [downloadUrl, setDownloadUrl] = useState("")
    const [error, setError] = useState("")
    const [migrationProgress, setMigrationProgress] = useState(0)
    const [currentMigrationStep, setCurrentMigrationStep] = useState(0)

    const ArrowDown = () => (
        <svg className="w-8 h-8 mx-auto my-4" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 4v16m0 0l-6-6m6 6l6-6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"
                  strokeLinejoin="round"/>
        </svg>
    );

    useEffect(() => {
        let interval: ReturnType<typeof setInterval> | undefined;
        if (isLoading && currentMigrationStep < migrationSteps.length) {
            interval = setInterval(() => {
                setCurrentMigrationStep(prev => {
                    if (prev < migrationSteps.length - 1) {
                        return prev + 1
                    }
                    if (interval !== undefined) {
                        clearInterval(interval)
                    }
                    return prev
                })
                setMigrationProgress(prev => Math.min(prev + 100 / migrationSteps.length, 100))
            }, 15000) // Change step every 15 seconds -- this is a heuristic value
        }
        return () => {
            if (interval !== undefined) {
                clearInterval(interval)
            }
        }
    }, [isLoading, currentMigrationStep])

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (step === 1) {
            setStep(2)
        } else if (step === 2) {
            setIsLoading(true)
            setError("")
            try {
                const result = await migratePaas({
                    source: selectedPaas,
                    destination: selectedCloud,
                    herokuApiKey,
                })
                setDownloadUrl(result.downloadUrl)
            } catch (error: unknown) {
                console.error("Migration failed:", error)
                let errorMessage = "An unexpected error occurred. Please try again."
                if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response) {
                    try {
                        const errorData = JSON.parse(error.response.data as string)
                        errorMessage = errorData.error || errorMessage
                    } catch (jsonError) {
                        console.error("Error parsing JSON:", jsonError)
                    }
                }
                setError(errorMessage)
            } finally {
                setIsLoading(false)
                setMigrationProgress(0)
                setCurrentMigrationStep(0)
            }
        }
    }

    const handlePrevious = () => {
        if (step > 1) {
            setStep(step - 1)
            setError("")
        }
    }

    return (
        <div className="max-w-md mx-auto p-6 bg-white rounded-lg">
            <h2 className="text-2xl font-bold mb-4 text-gray-800">Get Started with Qovery AI Cloud Migration</h2>

            {error && (
                <Alert variant="destructive" className="mb-4">
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <form onSubmit={handleSubmit}>
                {step === 1 && (
                    <div>
                        <div className="mb-4">
                            <label className="block mb-2 text-sm font-medium text-gray-700">Select Source PaaS
                                Platform</label>
                            <Select value={selectedPaas} onValueChange={setSelectedPaas} required>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Select Source PaaS"/>
                                </SelectTrigger>
                                <SelectContent>
                                    {paasOptions.map((option) => (
                                        <SelectItem key={option.value} value={option.value} disabled={option.disabled}>
                                            <div className="flex items-center">
                                                {option.icon}
                                                <span className="ml-2">{option.label}</span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <ArrowDown/>

                        <div>
                            <label className="block mb-2 text-sm font-medium text-gray-700">Select Target Cloud
                                Platform</label>
                            <Select value={selectedCloud} onValueChange={setSelectedCloud} required>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Select Target Cloud Platform"/>
                                </SelectTrigger>
                                <SelectContent>
                                    {cloudOptions.map((option) => (
                                        <SelectItem key={option.value} value={option.value}>
                                            <div className="flex items-center">
                                                {option.icon}
                                                <span className="ml-2">{option.label}</span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <p className="mt-4 text-sm text-gray-600">
                            Select the PaaS platform you want to migrate from and the cloud provider you want to migrate
                            your stack to.
                        </p>
                    </div>
                )}

                {step === 2 && (
                    <div>
                        <label className="block mb-2 text-sm font-medium text-gray-700">Enter Heroku API Key</label>
                        <Input
                            type="password"
                            value={herokuApiKey}
                            onChange={(e) => setHerokuApiKey(e.target.value)}
                            placeholder="Enter your Heroku API Key"
                            className="w-full mb-4"
                            required
                            disabled={isLoading}
                        />
                        <Alert>
                            <AlertTitle>🛡️ Your data is safe with us</AlertTitle>
                            <AlertDescription>
                                We only access your data in read-only mode. Our code is open-source and can be reviewed
                                on{" "}
                                <a
                                    href="https://github.com/Qovery/qovery-migration-ai-agent"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-blue-600 hover:underline"
                                >
                                    GitHub
                                </a>
                                .
                            </AlertDescription>
                        </Alert>
                        {!isLoading && (
                            <p className="mt-4 text-sm text-gray-600">
                                After clicking "Next", the generation of the migration files process will begin. This
                                can take up to 5 minutes
                                to complete.
                            </p>
                        )}
                        {isLoading && (
                            <div className="mt-4">
                                <h3 className="text-lg font-semibold mb-2 text-gray-800">Migration Progress</h3>
                                <Progress value={migrationProgress} className="w-full mb-4"/>
                                <p className="text-sm text-gray-600">
                                    {migrationSteps[currentMigrationStep]}...
                                </p>
                            </div>
                        )}
                        {downloadUrl && (
                            <div className="mt-4">
                                <p className="mb-2 text-gray-600">
                                    Your migration files have been generated successfully!
                                </p>
                                <Link href={downloadUrl} passHref>
                                    <Button className="w-full bg-violet-600 hover:bg-violet-700 text-white">
                                        Download Migration Files
                                    </Button>
                                </Link>
                                <p className="mt-4 text-sm text-gray-600">
                                    Please review the instructions in the README.md file inside the archive for next
                                    steps on reviewing and executing the Terraform files.
                                </p>
                            </div>
                        )}
                    </div>
                )}

                {(step === 1 || (step === 2 && !isLoading && !downloadUrl)) && (
                    <Button type="submit" className="w-full mt-4 bg-violet-600 hover:bg-violet-700 text-white"
                            disabled={isLoading}>
                        {isLoading ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                                Generating Migration Files...
                            </>
                        ) : (
                            "Next"
                        )}
                    </Button>
                )}
            </form>

            {step === 2 && !isLoading && !downloadUrl && (
                <Button variant="ghost" className="w-full mt-4" onClick={handlePrevious}>
                    Previous
                </Button>
            )}

            {step === 1 && (
                <Link href="/" passHref>
                    <Button variant="outline" className="w-full mt-4">
                        Go Back Home
                    </Button>
                </Link>
            )}
        </div>
    )
}