/* eslint-disable react-hooks/exhaustive-deps */

import { useEffect, useState, type ReactNode } from "react";
import Button from "./button";

interface TabOption {
  label: string;
  component: ReactNode;
  value: string;
}

interface Props {
  options: TabOption[];
  initValue?: string;
  onChange?(option: TabOption): void;
}

export default function TabSelector({ options, initValue, onChange }: Props) {
  const [idx, setIdx] = useState(0);

  useEffect(() => {
    if (!initValue) return;
    const initIndex = options.findIndex((e) => e.value === initValue);
    if (initIndex != -1) setIdx(initIndex);
  }, [initValue]);

  return (
    <div>
      <span className="flex gap-1 border border-border rounded-sm p-1">
        {options.map((e, i) => (
          <Button
            key={`selector${i}`}
            label={e.label}
            className="flex-1 items-center justify-center"
            type={idx === i ? "filled" : undefined}
            onClick={() => {
              setIdx(i);
              onChange?.(e);
            }}
          />
        ))}
      </span>
      {options[idx].component}
    </div>
  );
}
