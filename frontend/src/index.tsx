import React, { useCallback, useEffect } from "react";
import ReactDOM from "react-dom/client";
import { useFetch } from "./useFetch";
import { Subscriber } from "./subscriber";

import type { components } from "./oapi.gen";

function App() {
  const getHello = useCallback(
    (signal: AbortSignal) =>
      fetch("/hello", { signal }).then(
        (d) => d.json() as Promise<components["schemas"]["Hello"]>
      ),
    []
  );

  const data = useFetch(getHello);

  useEffect(() => {
    const ac = new AbortController();
    const s = new Subscriber("/subscribe", { signal: ac.signal });
    s.addEventListener("message", (e: CustomEventInit) => {
      console.log(e.detail);
    });

    return () => ac.abort();
  }, []);

  return (
    <div>
      <h1 className="text-3xl font-bold underline">Scanner Operator</h1>
      <p>{JSON.stringify(data)}</p>
    </div>
  );
}

const element = document.getElementById("app")!;
const root = ReactDOM.createRoot(element);
root.render(<App />);
