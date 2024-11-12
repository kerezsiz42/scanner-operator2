import React, { useCallback, useContext } from "react";
import { Button } from "./Button";
import { GlobalStateContext } from "./GlobalState";

export function ScanResultList() {
  const { state, dispatch } = useContext(GlobalStateContext);

  const deleteScanResult = useCallback(
    async (imageId: string) => {
      const res = await fetch(`/scan-results/${encodeURIComponent(imageId)}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        return;
      }

      dispatch({ type: "remove", payload: imageId });
    },
    [dispatch]
  );

  return (
    <div className="flex-1">
      <main className="max-w-[850px] mx-auto p-3">
        {state.scanResults.map((s) => (
          <div
            key={s.imageId}
            className="p-2 my-4 border rounded border-gray-900 text-gray-900 bg-slate-200 flex flex-col"
          >
            <div>{s.imageId}</div>
            <div className="flex justify-end">
              <Button
                onClick={() =>
                  window.open(
                    `/scan-results/${encodeURIComponent(s.imageId)}`,
                    "_blank"
                  )
                }
                text="Show Details"
                className="text-yellow-500 border-yellow-500"
              />
              <Button
                onClick={() => deleteScanResult(s.imageId)}
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
