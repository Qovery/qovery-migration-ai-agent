'use client'

import {useState} from "react"
import {Button} from "@/app/components/ui/button"
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/app/components/ui/select"
import {Input} from "@/app/components/ui/input"
import {Alert, AlertDescription, AlertTitle} from "@/app/components/ui/alert"
import {Loader2} from "lucide-react"
import {migratePaas} from "@/app/lib/api"
import Link from 'next/link'
import {useRouter} from 'next/navigation'

const paasOptions = [
    {value: "heroku", label: "Heroku"},
    {value: "render", label: "Render (coming soon)", disabled: true},
    {value: "railway", label: "Railway (coming soon)", disabled: true},
    {value: "fly", label: "Fly (coming soon)", disabled: true},
    {value: "vercel", label: "Vercel (coming soon)", disabled: true},
    {value: "netlify", label: "Netlify (coming soon)", disabled: true},
    {value: "platform", label: "Platform (coming soon)", disabled: true},
]

const cloudOptions = [
    {value: "aws", label: "AWS"},
    {value: "gcp", label: "GCP"},
    {value: "azure", label: "Azure"},
    {value: "scaleway", label: "Scaleway"},
]

export default function GetStartedFlow() {
    const [step, setStep] = useState(1)
    const [selectedPaas, setSelectedPaas] = useState("")
    const [selectedCloud, setSelectedCloud] = useState("")
    const [apiKey, setApiKey] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [downloadUrl, setDownloadUrl] = useState("")
    const router = useRouter()

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        if (step < 3) {
            setStep(step + 1)
        } else if (step === 3) {
            setIsLoading(true)
            try {
                const result = await migratePaas({
                    source: selectedPaas,
                    destination: selectedCloud,
                    apiKey,
                })
                setDownloadUrl(result.downloadUrl)
                setStep(4)
            } catch (error) {
                console.error("Migration failed:", error)
                // Handle error (show alert, etc.)
            } finally {
                setIsLoading(false)
            }
        }
    }

    return (
        <div className="max-w-md mx-auto p-6 bg-white rounded-lg">
            <h2 className="text-2xl font-bold mb-4 text-gray-800">Get Started with Qovery Migration</h2>
            <form onSubmit={handleSubmit}>
                {step === 1 && (
                    <div>
                        <label className="block mb-2 text-sm font-medium text-gray-700">Select PaaS Platform</label>
                        <Select value={selectedPaas} onValueChange={setSelectedPaas}>
                            <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select PaaS"/>
                            </SelectTrigger>
                            <SelectContent>
                                {paasOptions.map((option) => (
                                    <SelectItem key={option.value} value={option.value} disabled={option.disabled}>
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                )}

                {step === 2 && (
                    <div>
                        <label className="block mb-2 text-sm font-medium text-gray-700">Select Cloud Platform</label>
                        <Select value={selectedCloud} onValueChange={setSelectedCloud}>
                            <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select Cloud Platform"/>
                            </SelectTrigger>
                            <SelectContent>
                                {cloudOptions.map((option) => (
                                    <SelectItem key={option.value} value={option.value}>
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                )}

                {step === 3 && (
                    <div>
                        <label className="block mb-2 text-sm font-medium text-gray-700">Enter API Key</label>
                        <Input
                            type="password"
                            value={apiKey}
                            onChange={(e) => setApiKey(e.target.value)}
                            placeholder="Enter your API Key"
                            className="w-full mb-4"
                        />
                        <Alert>
                            <AlertTitle>Your data is safe with us</AlertTitle>
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
                    </div>
                )}

                {step === 4 && (
                    <div>
                        <h3 className="text-xl font-semibold mb-4 text-gray-800">Migration Complete!</h3>
                        <p className="mb-4 text-gray-600">
                            Your migration files have been generated successfully. Click the button below to download
                            the archive.
                        </p>
                        {downloadUrl ? (
                            <Link href={downloadUrl} passHref>
                                <Button className="w-full bg-purple-600 hover:bg-purple-700 text-white">
                                    Download Migration Files
                                </Button>
                            </Link>
                        ) : (
                            <Button
                                className="w-full bg-purple-600 hover:bg-purple-700 text-white"
                                disabled
                            >
                                Download Migration Files
                            </Button>
                        )}
                        <p className="mt-4 text-sm text-gray-600">
                            Please review the instructions in the README.md file inside the archive for next steps on
                            reviewing and
                            executing the Terraform files.
                        </p>
                    </div>
                )}

                {step < 4 && (
                    <Button type="submit" className="w-full mt-4 bg-purple-600 hover:bg-purple-700 text-white"
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

            <Button variant="ghost" className="w-full mt-4" onClick={() => router.push('/')}>
                Back to Home
            </Button>
        </div>
    )
}