import { useState, useEffect } from "react";

export function useFetch<T>(
  fetcher: (signal: AbortSignal) => Promise<T>
): T | undefined {
  const [result, setResult] = useState<T>();

  useEffect(() => {
    const ac = new AbortController();
    fetcher(ac.signal).then((d) => setResult(d));
    return () => ac.abort();
  }, [fetcher]);

  return result;
}
