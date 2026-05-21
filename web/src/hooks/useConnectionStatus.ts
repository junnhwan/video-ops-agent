import { useState, useEffect, useRef } from "react";

type Status = "online" | "offline" | "checking";

const POLL_INTERVAL = 15_000;
const MOCK_MODE = import.meta.env.VITE_USE_MOCK === "true";

export function useConnectionStatus() {
  const [status, setStatus] = useState<Status>(MOCK_MODE ? "online" : "checking");
  const timerRef = useRef<ReturnType<typeof setInterval>>(undefined);

  async function check() {
    if (MOCK_MODE) { setStatus("online"); return; }
    try {
      const base = import.meta.env.VITE_API_BASE_URL || "/api";
      const res = await fetch(`${base}/health`, { signal: AbortSignal.timeout(3000) });
      setStatus(res.ok ? "online" : "offline");
    } catch {
      setStatus("offline");
    }
  }

  useEffect(() => {
    check();
    timerRef.current = setInterval(check, POLL_INTERVAL);
    return () => clearInterval(timerRef.current);
  }, []);

  return status;
}
