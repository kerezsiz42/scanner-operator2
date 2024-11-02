import React, {
  createContext,
  PropsWithChildren,
  useCallback,
  useReducer,
} from "react";
import { useSubscriber } from "../hooks/useSubscriber";

export type ScanResult = { id: string };

const initialState = { isConnected: false, scanResults: [] as ScanResult[] };
type State = typeof initialState;
type Action =
  | { type: "add"; payload: ScanResult[] }
  | { type: "remove"; id: string }
  | { type: "connection_change"; isConnected: boolean };

function globalReducer(state: State, action: Action): State {
  switch (action.type) {
    case "connection_change":
      return { ...state, isConnected: action.isConnected };
    case "add":
      return {
        ...state,
        scanResults: [...state.scanResults, ...action.payload].sort((a, b) =>
          a.id.localeCompare(b.id)
        ),
      };
    case "remove":
      return {
        ...state,
        scanResults: state.scanResults.filter((s) => s.id !== action.id),
      };
  }
}

export const GlobalStateContext = createContext({
  state: initialState,
  dispatch: (action: Action) => {},
});

export type GlobalStateProviderProps = PropsWithChildren<{}>;

export const GlobalStateProvider = ({ children }: GlobalStateProviderProps) => {
  const [state, dispatch] = useReducer(globalReducer, initialState);

  const onMessage = useCallback((value: string) => {
    // TODO: fetch new scanResult
    const scanResult = { id: value };
    dispatch({ type: "add", payload: [scanResult] });
  }, []);

  const onConnection = useCallback(
    (isConnected: boolean) => {
      // TODO: fetch all scanResults when it gets online again
      dispatch({ type: "connection_change", isConnected });
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
