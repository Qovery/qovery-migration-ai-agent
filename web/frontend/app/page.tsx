import LandingPage from './components/LandingPage'
import Header from "@/app/components/Header";
import Footer from "@/app/components/Footer";

export default function Home() {
    return (
        <>
            <div className="flex flex-col min-h-screen bg-gradient-to-b from-[#1e1e3f] to-[#2d2d5f] text-white">
                <Header/>
                <LandingPage/>
                <Footer/>
            </div>
        </>
    )
}