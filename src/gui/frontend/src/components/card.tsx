import { type ReactNode } from "react";
import { cn } from "../utils/cn";

interface Props {
  children?: ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

export default function Card({ children, className, style }: Props) {
  return (
    <div
      style={style}
      className={cn(
        "border border-border rounded-md p-3 bg-cwhite shadow-xs",
        className,
      )}
    >
      {children}
    </div>
  );
}
