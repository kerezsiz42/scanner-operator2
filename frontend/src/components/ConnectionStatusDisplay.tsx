import React, { PropsWithChildren } from "react";

export type ConnectionStatusDisplay = PropsWithChildren<{
  isConnected: boolean;
}>;

export function ConnectionStatusDisplay({ isConnected }) {
  return (
    <div className="flex p-1 items-center">
      <div className="px-2 font-bold">{isConnected ? "Online" : "Offline"}</div>
      <div
        className={`rounded-[50%] w-4 h-4 ${
          isConnected ? "bg-green-500" : "bg-red-500"
        }`}
      ></div>
    </div>
  );
}
