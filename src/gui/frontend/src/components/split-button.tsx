import { useEffect, useRef, useState } from "react";
import { cn } from "../utils/cn";
import type { SelectableValue } from "./select";

interface Props {
  label: string;
  icon: string;
  enabled?: boolean;
  options: SelectableValue<string>[];
  color?: string;
  onSelectOption?(value: string): void;
  onClick?(): void;
}

export default function SplitButton({
  label,
  icon,
  options,
  enabled = true,
  color,
  onSelectOption,
  onClick,
}: Props) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  return (
    <div
      className={cn(
        `border-2 rounded-md transition-colors 
        duration-200 cursor-pointer w-min flex items-stretch relative hoverable-gray`,
        {
          "pointer-events-none opacity-50 cursor-auto": !enabled,
        },
      )}
      style={{
        // backgroundColor: `color-mix(in srgb, ${color ?? "var(--color-primary)"} 50%, transparent)`,
        borderColor: color ?? "var(--color-primary)",
      }}
    >
      <button
        className="flex items-center pr-3 pl-1.5 py-1.5 gap-2
        font-medium cursor-pointer hover:bg-cwhite/10"
        onClick={onClick}
        style={{
          color: color ?? "var(--color-primary)",
        }}
      >
        <span className="material-symbols-rounded">{icon}</span>
        <label className="cursor-pointer">{label}</label>
      </button>
      <button
        className="flex items-center hover:bg-cwhite/10 cursor-pointer border-l-2"
        onClick={() => setIsOpen((s) => !s)}
        style={{
          color: color ?? "var(--color-primary)",
          borderColor: color ?? "var(--color-primary)",
        }}
      >
        <span className="material-symbols-rounded">
          {isOpen ? "keyboard_arrow_up" : "keyboard_arrow_down"}
        </span>
      </button>
      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute top-full left-0 rounded-md border border-border bg-cwhite 
          flex flex-col gap-1 mt-1 w-full shadow-sm p-1"
        >
          {options.map((e) => (
            <button
              onClick={() => {
                setIsOpen(false);
                onSelectOption?.(e.value);
              }}
              className="px-2 py-1 hover:bg-gray rounded-md text-left overflow-clip"
            >
              {e.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
