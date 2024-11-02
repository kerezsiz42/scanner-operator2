import React, { useContext } from "react";
import { KubernetesLogo } from "./KubernetesLogo";
import { ConnectionStatusDisplay } from "./ConnectionStatusDisplay";
import { GlobalStateContext } from "./GlobalState";

export function Navbar() {
  const { state } = useContext(GlobalStateContext);

  return (
    <div className="bg-slate-300 border-gray-900 border-b">
      <div className="px-4 flex items-center justify-between">
        <div className="p-1 flex items-center">
          <KubernetesLogo />
          <h1 className="text-3xl font-semibold text-gray-900 px-2">
            Scanner Operator
          </h1>
        </div>
        <ConnectionStatusDisplay isConnected={state.isConnected} />
      </div>
    </div>
  );
}
