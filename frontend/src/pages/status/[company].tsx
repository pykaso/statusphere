import axios from "@/utils/axios";
import { StatusPage } from "@/model/StatusPage";
import { Status } from "@/model/Status";
import { CurrentStatus } from "@/components/CurrentStatus";
import { Outages } from "@/components/Outages";
import { RecommendCompany } from "@/components/RecommendCompany";
import Head from "next/head";
import React from "react";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { CopyBlock, dracula } from "react-code-blocks";

interface CompanyStatusPageProps {
  statusPageDetails: StatusPage;
  currStatus: Status;
  companyName: string;
  outages: Incident[];
  isError?: boolean;
  apiCallCurrentStatus: string;
  apiCallIncidents: string;
}

export async function getServerSideProps(context: any) {
  const companyName = context.params.company;
  try {
    const statusPageResp = await axios.get(
      "/api/v1/statusPage?statusPageName=" + companyName
    );
    const statusPageDetails: StatusPage = statusPageResp.data.statusPage;
    const currStatusResp = axios.get(
      "/api/v1/currentStatus?statusPageUrl=" + statusPageDetails.url
    );
    const outagesResp = axios.get(
      "/api/v1/incidents?limit=50&statusPageUrl=" + statusPageDetails.url
    );

    const currStatus: Status = (await currStatusResp).data.status;
    const outages = (await outagesResp).data.incidents;

    return {
      props: {
        statusPageDetails: statusPageDetails,
        currStatus: currStatus,
        companyName: companyName,
        outages: outages,
      },
    };
  } catch (e) {
    return {
      props: {
        isError: true,
        companyName: companyName,
      },
    };
  }
}

export default function CompanyStatusPage({
  statusPageDetails,
  currStatus,
  companyName,
  outages,
  isError,
  apiCallCurrentStatus,
  apiCallIncidents,
}: CompanyStatusPageProps) {
  return (
    <div>
      <div className={"mt-4 flex justify-center w-full z-10"}>
        <div className={"flex justify-left"}>
          <h1 className={"text-2xl"}>{companyName}</h1>
        </div>
      </div>

      <div className={"mt-4 flex justify-center w-full z-10"}>
        <Head>
          <title>{statusPageDetails.name} - Stav služby</title>
          <meta
            name="description"
            content={`Current status of ${statusPageDetails.name}. Is ${statusPageDetails.name} down?`}
          />
        </Head>

        <div className={"w-[90vw] lg:w-[80vw] space-y-8 flex justify-center"}>
          <div>
            {isError ? (
              <RecommendCompany input={companyName} />
            ) : (
              <>
                <CurrentStatus
                  displayName={statusPageDetails.name}
                  status={currStatus}
                  lastCurrentlyScraped={statusPageDetails.lastCurrentlyScraped}
                  statusPageUrl={statusPageDetails.url}
                />
                <Outages
                  statusPageDetails={statusPageDetails}
                  incidents={outages}
                />
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
