import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { StatusPage } from "@/model/StatusPage";
import { calculateDuration, convertToSimpleDate } from "@/utils/datetime";
import { ReadMore } from "@/components/ReadMore";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/components/ui/lib/utils";
import { useEffect, useState } from "react";

interface OutagesProps {
  statusPageDetails: StatusPage;
  incidents: Incident[];
}

function getBadgeColour(impact: string) {
  switch (impact) {
    case "none":
      return "bg-direct-green";
    case "minor":
      return "bg-yellow-400";
    case "major":
      return "bg-red-400";
    case "critical":
      return "bg-red-700";
    case "maintenance":
      return "bg-blue-400";
    default:
      return "bg-gray-400";
  }
}

export function Outages(props: OutagesProps) {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    setIsMobile(window.innerWidth <= 640);
  }, []);

  if (!props.statusPageDetails.isIndexed) {
    return (
      <div>
        <Card className={"bg-white"}>
          <CardHeader className={"items-left"}>
            <CardTitle
              className={"scroll-m-20 text-xl font-semibold tracking-tight"}
            >
              Incidents are not currently indexed for{" "}
              {props.statusPageDetails.name}
            </CardTitle>
            <CardDescription className={"leading-7 [&:not(:first-child)]:mt-6"}>
              You can view the official status page at:{" "}
              <a href={props.statusPageDetails.url}>
                {" "}
                {props.statusPageDetails.name} status page
              </a>
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (props.incidents === undefined || props.incidents.length === 0) {
    return (
      <div className={"mt-4"}>
        <h2 className="scroll-m-20 border-b pb-2 text-2xl font-semibold first:mt-0">
          Incidenty
        </h2>
        <div>Žádné incidenty neevidujeme</div>
      </div>
    );
  }

  return (
    <div className={"mt-4"}>
      <h2 className="scroll-m-20 border-b pb-2 text-2xl font-semibold first:mt-0">
        Incidenty
      </h2>
      <Table className={"bg-white"}>
        <TableCaption>
          {" "}
          Zdroj:
          <a href={props.statusPageDetails.url}>
            {" "}
            Oficiální {props.statusPageDetails.name} status stánka nebo API
          </a>
        </TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className="w-48">Začátek (UTC)</TableHead>
            <TableHead className="w-36">Závažnost</TableHead>
            <TableHead className="w-36 text-left">Trvání</TableHead>
            {!isMobile && <TableHead>Popis</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {props.incidents.map((incident) => (
            <TableRow>
              <TableCell>{convertToSimpleDate(incident.startTime)}</TableCell>
              <TableCell>
                <Badge className={getBadgeColour(incident.impact)}>
                  {incident.impact}
                </Badge>
              </TableCell>
              <TableCell>
                {calculateDuration(incident.startTime, incident.endTime)}
              </TableCell>
              {!isMobile && (
                <TableCell className="text-left">
                  <ReadMore text={incident.description} />
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
