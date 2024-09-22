import Link from "next/link"

export default function Footer() {
    return (
        <footer className="w-full py-6 px-4 border-t border-gray-800">
            <div className="container mx-auto flex flex-col sm:flex-row justify-between items-center">
                <p className="text-sm text-gray-400">© 2024 Qovery Inc. All rights reserved.</p>
                <nav className="flex gap-6 mt-4 sm:mt-0">
                    <Link className="text-sm text-gray-400 hover:text-purple-400 transition-colors" href="https://www.qovery.com/terms">
                        Terms of Service
                    </Link>
                    <Link className="text-sm text-gray-400 hover:text-purple-400 transition-colors" href="https://www.qovery.com/private-policy">
                        Privacy Policy
                    </Link>
                </nav>
            </div>
        </footer>
    )
}