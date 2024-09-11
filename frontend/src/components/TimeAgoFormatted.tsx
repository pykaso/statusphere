"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/router";
import { timeAgo } from "@/utils/datetime";
interface TimeAgoFormattedProps {
  time: string;
}

export function TimeAgoFormatted(props: TimeAgoFormattedProps) {
  const [isClient, setIsClient] = useState(false);

  // Nastavíme isClient na true po hydrataci (když běží na klientu)
  useEffect(() => {
    setIsClient(true);
  }, []);

  // Pokud není na klientu, nezobrazuj časově závislé informace
  if (!isClient) {
    return null; // Nebude vykresleno na serveru
  }

  return <span>{timeAgo(props.time)}</span>;
}
