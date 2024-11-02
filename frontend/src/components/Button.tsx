import React, { PropsWithChildren } from "react";

export type ButtonProps = PropsWithChildren<{
  onClick: () => void;
  text: string;
  className?: string;
}>;

export function Button({ onClick, text, className }: ButtonProps) {
  return (
    <div className="px-1">
      <button
        type="button"
        className={`p-2 rounded border hover:underline font-bold text-black border-black ${
          className ?? className
        }`}
        onClick={onClick}
      >
        {text}
      </button>
    </div>
  );
}
