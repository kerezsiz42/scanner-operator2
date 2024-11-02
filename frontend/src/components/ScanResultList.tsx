import React, { useContext } from "react";
import { Button } from "./Button";
import { GlobalStateContext } from "./GlobalState";

export function ScanResultList() {
  const { state, dispatch } = useContext(GlobalStateContext);

  async function deleteScanResult(id: string) {
    dispatch({ type: "remove", id });
    // TODO: implement
    // fetch("/scan", { method: "DELETE" }).then((d) => {
    //   if (d.ok) {
    //     dispatch({ type: "remove", id });
    //   }
    // });
  }

  return (
    <div className="flex-1">
      <main className="max-w-[850px] mx-auto p-3">
        {state.scanResults.map((s) => (
          <div
            key={s.id}
            className="p-2 my-4 border rounded border-gray-900 text-gray-900 bg-slate-200 flex flex-col"
          >
            <div>{s.id}</div>
            <div className="flex justify-end">
              <Button
                onClick={() => console.log("hello")}
                text="Show Details"
                className="text-yellow-500 border-yellow-500"
              />
              <Button
                onClick={() => deleteScanResult(s.id)}
                text="Remove"
                className="text-red-500 border-red-500"
              />
            </div>
          </div>
        ))}
      </main>
    </div>
  );
}
