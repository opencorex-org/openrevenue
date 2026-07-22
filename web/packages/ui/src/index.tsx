import type { ButtonHTMLAttributes, PropsWithChildren } from "react";
export function Button({ children, ...props }: PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>>) { return <button className="rounded bg-blue-700 px-4 py-2 font-medium text-white focus:outline-none focus:ring" {...props}>{children}</button>; }
