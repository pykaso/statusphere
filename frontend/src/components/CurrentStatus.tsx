import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Status } from "@/model/Status";
import { TimeAgoFormatted } from "./TimeAgoFormatted";
import { getStatusColor, getStatusText } from "./SmallStatus";

interface CurrentStatusProps {
  displayName: string;
  status: string;
  lastCurrentlyScraped: string;
  statusPageUrl: string;
}

export function CurrentStatus(props: CurrentStatusProps) {
  if (props.status === Status.UNKNOWN) {
    return <div></div>;
  }

  return (
    <Card className={getStatusColor(props.status)}>
      <CardHeader className={"items-left"}>
        <CardTitle className="scroll-m-20 border-b pb-2 text-2xl font-semibold tracking-tight first:mt-0">
          Aktuální stav: {getStatusText(props.status)}
        </CardTitle>
      </CardHeader>
      <CardFooter>
        <p className="leading-7 [&:not(:first-child)]:mt-6">
          Poslední kontrola{" "}
          <a href={props.statusPageUrl}>
            {" "}
            oficiální {props.displayName} status page
          </a>{" "}
          proběhla před <TimeAgoFormatted time={props.lastCurrentlyScraped} />.
        </p>
      </CardFooter>
    </Card>
  );
}
