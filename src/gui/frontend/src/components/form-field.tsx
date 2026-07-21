/* eslint-disable react-hooks/set-state-in-effect */
import {
  useEffect,
  useMemo,
  useState,
  type ChangeEvent,
  type ReactNode,
} from "react";
import { debounce } from "../utils/debouncer";
import InfoTooltip from "./info-tooltip";

interface Props {
  hint?: string;
  label?: string;
  type?: React.HTMLInputTypeAttribute;
  initValue?: string;
  onChange?(value: string): void;
  suffix?: ReactNode;
  prefix?: ReactNode;
  min?: number;
  max?: number;
  enabled?: boolean;
  tooltip?: string;
}

export default function FormField({
  hint,
  initValue,
  onChange,
  label,
  prefix,
  suffix,
  type,
  min,
  max,
  enabled = true,
  tooltip,
}: Props) {
  const [localValue, setLocalValue] = useState(initValue ?? "");

  useEffect(() => {
    setLocalValue(initValue ?? "");
  }, [initValue]);

  const debouncedChange = useMemo(
    () => debounce((value: string) => onChange?.(value), 500),
    [onChange],
  );

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    const value = event.target.value;

    setLocalValue(value);
    debouncedChange(value);
  };

  return (
    <div>
      <label className="text-dark-gray flex items-center gap-1 mb-1">
        {label}
        {tooltip && <InfoTooltip text={tooltip} />}
      </label>
      <span className="flex flex-row items-center gap-2">
        {prefix ?? prefix}
        <input
          disabled={!enabled}
          type={type ?? "text"}
          className="rounded-md border border-border px-3 py-1.5 w-full 
           disabled:opacity-20 placeholder text-field"
          style={{}}
          placeholder={hint}
          value={localValue !== undefined ? String(localValue) : ""}
          min={min}
          max={max}
          onChange={handleChange}
        />
        {suffix ?? suffix}
      </span>
    </div>
  );
}
