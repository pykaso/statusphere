import { Status } from "./Status";

export interface StatusPage {
  name: string;
  url: string;
  lastHistoricallyScraped: string;
  lastCurrentlyScraped: string;
  isIndexed: boolean;
  status: string;
}
