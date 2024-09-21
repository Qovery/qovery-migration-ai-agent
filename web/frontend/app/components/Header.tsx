import Link from "next/link"
import Image from "next/image"
import {Button} from "@/app/components/ui/button"
import {Github} from "lucide-react"
import QoveryLogo from '/images/qovery-logo-square.svg';

export default function Header() {
    return (
        <header className="px-4 lg:px-6 h-16 flex items-center">
            <Link className="flex items-center justify-center" href="/">
                <QoveryLogo width={32} height={32} className="mr-2"/>
                <span className="font-bold text-xl text-white">Qovery AI Cloud Migration</span>
            </Link>
            <nav className="ml-auto flex gap-6">
                <Link className="text-sm font-medium hover:text-purple-400 transition-colors" href="https://discuss.qovery.com">
                    Forum
                </Link>
            </nav>
            <div className="ml-6 flex gap-4">
                <Link href="https://github.com/Qovery/qovery-migration-ai-agent">
                    <Button variant="outline"
                            className="text-white border-white hover:bg-white hover:text-[#1e1e3f] flex items-center">
                        <Github className="mr-2 h-4 w-4"/>
                        Give a Star
                    </Button>
                </Link>
                <Link href="https://console.qovery.com">
                    <Button className="bg-white text-[#1e1e3f] hover:bg-gray-200">
                        Sign Up
                    </Button>
                </Link>
            </div>
        </header>
    )
}