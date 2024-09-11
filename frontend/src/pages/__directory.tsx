import { StatusPage } from "@/model/StatusPage";
import { Status } from "@/model/Status";
import axios from "@/utils/axios";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { NavLink } from "@/components/mainsite/NavLink";
import { Metadata } from "next";
import { Props } from "next/script";
import Head from "next/head";
import SmallStatus from "@/components/SmallStatus";

interface DirectoryProps {
  companyList: StatusPage[];
}

export async function getServerSideProps() {
  const response = await axios.get("/api/v1/statusPages");
  let companyList: StatusPage[] = response.data.statusPages;
  //   companyList.forEach((company) => {
  //     // Capitalize first letter of company name
  //     company.name = company.name.charAt(0).toUpperCase() + company.name.slice(1);
  //     let status = await getCurrentStatus(company.url);
  //   });

  for await (const company of companyList) {
    // Capitalize first letter of company name
    company.name = company.name.charAt(0).toUpperCase() + company.name.slice(1);
    let status: Status = await getCurrentStatus(company.url);
    company.status = status;
  }

  return {
    props: {
      companyList: response.data.statusPages,
    },
  };
}

export async function getCurrentStatus(pageUrl: String) {
  const currStatusResp = axios.get(
    "/api/v1/currentStatus?statusPageUrl=" + pageUrl
  );
  const status: Status = (await currStatusResp).data.status;
  return status;
}

export async function generateMetadata() {
  return {
    title: `Statusphere - Directory`,
  };
}

export function Directory({ companyList }: DirectoryProps) {
  return (
    <div className={"xxm-4"}>
      <Head>
        <title>Stav externích API</title>
        <meta name="description" content="Stav externích API" />
      </Head>
      <h3 className="scroll-m-20 text-2xl font-semibold tracking-tight">
        Externí služby
      </h3>

      <div className="m-4">
        <Table className={"bg-white"}>
          <TableHeader>
            <TableRow>
              <TableHead>Název</TableHead>
              <TableHead>Stav</TableHead>
              <TableHead>Detaily</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {companyList.map((company) => (
              <TableRow key={company.name + ""}>
                <TableCell key={company.name + "-name"}>
                  {company.name}
                </TableCell>
                <TableCell key={company.name + "-status"}>
                  <SmallStatus
                    displayName={company.name}
                    status={company.status}
                    lastCurrentlyScraped={company.lastCurrentlyScraped}
                    statusPageUrl={company.url}
                  />
                </TableCell>
                <TableCell key={company.name + "-link"}>
                  <NavLink href={"/status/" + company.name}>
                    Zobrazit detaily
                  </NavLink>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

export default Directory;
