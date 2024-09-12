"use client";

import { useEffect, useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { cs } from "date-fns/locale";

interface TimeAgoFormattedProps {
  time: string;
}

export function TimeAgoFormatted({ time }: TimeAgoFormattedProps) {
  const [formattedTime, setFormattedTime] = useState<string>(() =>
    formatDistanceToNow(new Date(time), { addSuffix: true, locale: cs })
  );

  useEffect(() => {
    const timer = setInterval(() => {
      setFormattedTime(
        formatDistanceToNow(new Date(time), { addSuffix: true, locale: cs })
      );
    }, 1000);

    return () => clearInterval(timer);
  }, [time]);

  return <span suppressHydrationWarning>{formattedTime}</span>;
}
