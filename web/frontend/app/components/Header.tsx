'use client';

import Link from "next/link"
import {Button} from "@/app/components/ui/button"
import {Github, Menu, X} from "lucide-react"
import QoveryLogo from '@/images/qovery-logo-square.svg';
import {useState} from "react"

export default function Header() {
    const [isMenuOpen, setIsMenuOpen] = useState(false);

    return (
        <>
            <header className="px-4 lg:px-6 h-16 flex items-center justify-between">
                <Link className="flex items-center justify-center" href="/">
                    <QoveryLogo width={32} height={32} className="mr-2"/>
                    <span className="font-bold text-xl text-white">Qovery AI Cloud Migration</span>
                </Link>
                <nav className="hidden md:flex gap-6 items-center">
                    <Link className="text-sm font-medium hover:text-purple-400 transition-colors"
                          href="https://discuss.qovery.com">
                        Forum
                    </Link>
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
                </nav>
                <button className="md:hidden" onClick={() => setIsMenuOpen(!isMenuOpen)}>
                    {isMenuOpen ? <X className="text-white"/> : <Menu className="text-white"/>}
                </button>
            </header>
            {isMenuOpen && (
                <div className="md:hidden bg-[#1e1e3f] p-4">
                    <nav className="flex flex-col space-y-2">
                        <Link className="text-sm font-medium hover:text-purple-400 transition-colors"
                              href="https://discuss.qovery.com">
                            Forum
                        </Link>
                        <Link href="https://github.com/Qovery/qovery-migration-ai-agent">
                            <Button variant="outline"
                                    className="text-white border-white hover:bg-white hover:text-[#1e1e3f] flex items-center justify-center w-full">
                                <Github className="mr-2 h-4 w-4"/>
                                Give a Star
                            </Button>
                        </Link>
                        <Link href="https://console.qovery.com">
                            <Button className="bg-white text-[#1e1e3f] hover:bg-gray-200 w-full">
                                Sign Up
                            </Button>
                        </Link>
                    </nav>
                </div>
            )}
        </>
    )
}