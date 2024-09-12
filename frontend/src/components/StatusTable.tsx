"use client";

import { StatusPage } from "@/model/StatusPage";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { NavLink } from "@/components/mainsite/NavLink";
import SmallStatus from "@/components/SmallStatus";
import axios from "@/utils/axios";
import { Status } from "@/model/Status";
import { TimeAgoFormatted } from "./TimeAgoFormatted";

interface StatusTableProps {
  companyList: StatusPage[];
}

export default function StatusTable({ companyList }: StatusTableProps) {
  return (
    <div className="m-4">
      <Table className={"bg-white"}>
        <TableHeader>
          <TableRow>
            <TableHead>Název</TableHead>
            <TableHead>Stav</TableHead>
            <TableHead>Poslední kontrola před</TableHead>
            <TableHead>Detaily</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {companyList.map((company) => (
            <TableRow key={company.name + ""}>
              <TableCell key={company.name + "-name"}>{company.name}</TableCell>
              <TableCell key={company.name + "-status"}>
                <SmallStatus
                  displayName={company.name}
                  status={company.status}
                  lastCurrentlyScraped={company.lastCurrentlyScraped}
                  statusPageUrl={company.url}
                />
              </TableCell>
              <TableCell>
                <TimeAgoFormatted time={company.lastCurrentlyScraped} />
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
  );
}
