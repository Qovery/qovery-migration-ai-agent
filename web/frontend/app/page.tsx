import LandingPage from './components/LandingPage'
import Header from "@/app/components/Header";
import Footer from "@/app/components/Footer";

export default function Home() {
    return (
        <>
            <div className="min-h-screen font-[family-name:var(--font-geist-sans)]">
                <Header/>
                <LandingPage/>
                <Footer/>
            </div>
        </>
    )
}