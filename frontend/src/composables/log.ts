"use strict";

// Output logging information to console.
export function log(level: "INFO" | "ERROR", content: any) {
  const time = new Date();
  const msg = "[" + time.toLocaleTimeString() + "] ";
  switch (level) {
    case "INFO":
      console.log(msg, content);
      break;
    case "ERROR":
      console.trace(msg, content);
  }
}
