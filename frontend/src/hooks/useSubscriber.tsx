import { useEffect } from "react";
import { Subscriber } from "../subscriber";

export type UseSubscriberProps = {
  onMessage: (value: string) => void;
  onConnection: (value: boolean) => void;
};

export function useSubscriber({ onMessage, onConnection }: UseSubscriberProps) {
  useEffect(() => {
    const ac = new AbortController();
    const s = new Subscriber("/subscribe", { signal: ac.signal });

    s.addEventListener("message", (e: CustomEventInit) => {
      onMessage(e.detail as string);
    });

    s.addEventListener("connection", (e: CustomEventInit) => {
      onConnection(e.detail as boolean);
    });

    return () => ac.abort();
  }, [onMessage, onConnection]);
}
