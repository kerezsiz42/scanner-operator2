import React, {
  createContext,
  PropsWithChildren,
  useCallback,
  useReducer,
} from "react";
import { useSubscriber } from "../hooks/useSubscriber";
import { components } from "../oapi.gen";

type ScanResult = components["schemas"]["ScanResult"];

const initialState = {
  isConnected: false,
  scanResults: [] as ScanResult[],
};
type State = typeof initialState;
type Action =
  | { type: "add"; payload: ScanResult }
  | { type: "remove"; payload: ScanResult["imageId"] }
  | { type: "connection_gained"; payload: ScanResult[] }
  | { type: "connection_lost" };

function globalReducer(state: State, action: Action): State {
  switch (action.type) {
    case "connection_lost": {
      return { ...state, isConnected: false };
    }
    case "connection_gained": {
      return {
        ...state,
        isConnected: true,
        scanResults: action.payload.sort((a, b) =>
          a.imageId.localeCompare(b.imageId)
        ),
      };
    }
    case "add": {
      const index = state.scanResults.findIndex(
        (item) => item.imageId === action.payload.imageId
      );

      let scanResults: ScanResult[] = [];
      if (index !== -1) {
        scanResults = state.scanResults.map((s, i) =>
          i === index ? action.payload : s
        );
      } else {
        scanResults = [...state.scanResults, action.payload];
      }

      return {
        ...state,
        scanResults: scanResults.sort((a, b) =>
          a.imageId.localeCompare(b.imageId)
        ),
      };
    }
    case "remove": {
      return {
        ...state,
        scanResults: state.scanResults.filter(
          (s) => s.imageId !== action.payload
        ),
      };
    }
  }
}

export const GlobalStateContext = createContext({
  state: initialState,
  dispatch: (action: Action) => {},
});

export type GlobalStateProviderProps = PropsWithChildren<{}>;

export const GlobalStateProvider = ({ children }: GlobalStateProviderProps) => {
  const [state, dispatch] = useReducer(globalReducer, initialState);

  const onMessage = useCallback(
    async (imageId: string) => {
      const res = await fetch(`/scan-results/${encodeURIComponent(imageId)}`);
      if (!res.ok) {
        return;
      }

      const scanResult = (await res.json()) as ScanResult;
      dispatch({ type: "add", payload: scanResult });
    },
    [dispatch]
  );

  const onConnection = useCallback(
    async (isConnected: boolean) => {
      if (isConnected) {
        const res = await fetch(`/scan-results`);
        if (!res.ok) {
          return;
        }

        const scanResults = (await res.json()) as ScanResult[];
        dispatch({ type: "connection_gained", payload: scanResults });
      } else {
        dispatch({ type: "connection_lost" });
      }
    },
    [dispatch]
  );

  useSubscriber({ onMessage, onConnection });

  return (
    <GlobalStateContext.Provider value={{ state, dispatch }}>
      {children}
    </GlobalStateContext.Provider>
  );
};
