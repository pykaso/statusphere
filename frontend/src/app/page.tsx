import Header from "@/components/header";
import axios from "@/utils/axios";
import { StatusPage } from "@/model/StatusPage";
import { Status } from "@/model/Status";
import StatusTable from "@/components/StatusTable";
import Head from "next/head";

export const revalidate = 0;

export function generateMetadata() {
  return {
    title: `Stav externích služeb`,
  };
}

async function getStatuses(): Promise<StatusPage[]> {
  const response = await axios.get("/api/v1/statusPages");
  let companyList: StatusPage[] = response.data.statusPages;

  for await (const company of companyList) {
    // Capitalize first letter of company name
    company.name = company.name.charAt(0).toUpperCase() + company.name.slice(1);
    let status: Status = await getCurrentStatus(company.url);
    company.status = status;
  }
  return companyList;
}

async function getCurrentStatus(pageUrl: String) {
  const currStatusResp = axios.get(
    "/api/v1/currentStatus?statusPageUrl=" + pageUrl
  );
  const status: Status = (await currStatusResp).data.status;
  return status;
}

export default async function Home() {
  const companyList = await getStatuses();
  return (
    <>
      <Header />
      <StatusTable companyList={companyList} />
    </>
  );
}
