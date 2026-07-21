import type { CSSProperties } from "react";
import { cn } from "../utils/cn";

interface Props {
  label?: string;
  type?: "filled" | "outlined";
  className?: string;
  style?: CSSProperties;
  icon?: string;
  onClick?: () => void;
  enabled?: boolean;
}

export default function Button({
  label,
  type,
  icon,
  onClick,
  style,
  className,
  enabled = true,
}: Props) {
  return (
    <button
      style={style}
      disabled={!enabled}
      onClick={onClick}
      className={cn(
        "onDisable cursor-pointer p-1.5 hoverable-gray flex flex-row items-center jusitfy-center rounded-md transition-colors duration-200",
        {
          "text-onPrimary bg-primary hoverable-primary": type === "filled",
          "px-3 text-md font-semibold": label,
          "border border-dark-gray font-medium": type === "outlined",
          "gap-2": label && icon,
        },
        className,
      )}
    >
      <span className="material-symbols-rounded">{icon}</span>
      <p>{label}</p>
    </button>
  );
}
