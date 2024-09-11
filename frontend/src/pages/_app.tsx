import "@/index.css";
import { AppProps } from "next/app";
import Layout from "@/components/Layout";
import Head from "next/head";

export default function MyApp({ Component, pageProps }: AppProps) {
  return (
    <div className="test">
      <Head>
        <meta property="og:locale" content="en" />
      </Head>
      <Layout>
        <Component {...pageProps} />
      </Layout>
    </div>
  );
}
