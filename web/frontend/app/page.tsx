import LandingPage from './components/LandingPage'
import Head from "next/head";
import Header from "@/app/components/Header";
import Footer from "@/app/components/Footer";

export default function Home() {
    return (
        <>
            <Head>
                <title>Qovery AI Migration Agent</title>
                <meta name="description" content="Simplify your cloud migration with Qovery AI Migration Agent"/>
                <link rel="icon" href="/favicon.ico"/>
            </Head>
            <div className="min-h-screen font-[family-name:var(--font-geist-sans)]">
                <Header/>
                <LandingPage/>
                <Footer/>
            </div>
        </>
    )
}