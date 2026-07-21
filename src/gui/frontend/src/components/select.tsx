export interface SelectableValue<T> {
  value: T;
  label: string;
}

interface Props<T> {
  label?: string;
  initValue?: T;
  onChange?(value?: T): void;
  enabled?: boolean;
  options: SelectableValue<T>[];
}

export default function Select<T>({
  initValue,
  onChange,
  label,
  options,
  enabled = true,
}: Props<T>) {
  return (
    <div className="flex gap-2 items-center">
      <label className="text-dark-gray">{label}</label>
      <select
        disabled={!enabled}
        value={initValue !== undefined ? String(initValue) : ""}
        onChange={(e) => onChange?.(e.target.value as T)}
      >
        {options.map((e) => (
          <option key={String(e.value)} value={e.value as string}>
            {e.label}
          </option>
        ))}
      </select>
    </div>
  );
}
