import { useEffect, useState } from "react";

interface Props {
  value?: boolean;
  onChange?(value: boolean): void;
}

export default function Checkbox({ value, onChange }: Props) {
  const [val, setVal] = useState(value);

  useEffect(() => {
    setVal(value);
  }, [value]);

  return (
    <input
      checked={val}
      type="checkbox"
      value={`${val}`}
      className="custom-checkbox"
      onChange={(e) => {
        onChange?.(e.target.checked);
      }}
    />
  );
}
