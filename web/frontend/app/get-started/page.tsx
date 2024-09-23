'use client'

import GetStartedFlow from '../components/GetStartedFlow'
import Header from "@/app/components/Header";
import Footer from "@/app/components/Footer";

export default function GetStartedPage() {
    return (
        <>
            <div className="flex flex-col min-h-screen bg-gradient-to-b from-[#1e1e3f] to-[#2d2d5f] text-white">
                <Header/>
                <div
                    className="min-h-screen text-white flex items-center justify-center">
                    <div className="bg-white p-6 rounded-lg shadow-xl max-w-md w-full text-black">
                        <GetStartedFlow/>
                    </div>
                </div>
                <Footer/>
            </div>
        </>
    )
}