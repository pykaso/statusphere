import { Status } from "@/model/Status";
import { timeAgo } from "@/utils/datetime";

interface SmallStatusProps {
  displayName: string;
  status: string;
  lastCurrentlyScraped: string;
  statusPageUrl: string;
}

export default function SmallStatus(props: SmallStatusProps) {
  return (
    <div
      className={
        "w-36 text-center p-2 rounded-2xl" + " " + getStatusColor(props.status)
      }
    >
      <div className="text-l text-white font-semibold">
        {getStatusText(props.status)}
      </div>
    </div>
  );
}

export const getStatusColor = (status: string): string => {
  if (status.toUpperCase() === Status.UP) {
    return "bg-direct-green";
  }
  if (
    status.toUpperCase() === Status.DOWN ||
    status.toUpperCase() === Status.DEGRADED
  ) {
    return "bg-red-400";
  }
  return "bg-blue-400";
};

export const getStatusText = (status: string): string => {
  if (status.toUpperCase() === Status.UP) {
    return "OK";
  }
  if (
    status.toUpperCase() === Status.DOWN ||
    status.toUpperCase() === Status.DEGRADED
  ) {
    return "CHYBA";
  }
  return "NEZNÁMÝ";
};

const getFooter = (status: string, lastCurrentlyScraped: string): string => {
  if (status.toUpperCase() === Status.UNKNOWN) {
    return "Zatím nevíme zda služba funguje.";
  }
  return "Last checked "; //{timeAgo(lastCurrentlyScraped)} ago.";
};
