/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useMemo, useState } from "react";
import Button from "./button";
import Checkbox from "./checkbox";
import FilePickerField from "./file-picker-field";
import FormField from "./form-field";
import Select from "./select";

interface Props {
  schemaRaw: string;
  values: Record<string, any>;
  onChange?(key: string, value: any): void;
}

export default function DynamicForm({ schemaRaw, values, onChange }: Props) {
  const [currentValues, setCurrentValues] = useState(values);
  const schema: { properties: Record<string, any> } = useMemo(
    () => JSON.parse(schemaRaw),
    [schemaRaw],
  );

  useEffect(() => {
    if (!schema.properties) return;

    for (const [k, v] of Object.entries(schema.properties)) {
      if (v.default === undefined || values[k] !== undefined) continue;
      onChange?.(k, v.default);
      setCurrentValues((s) => ({ ...s, [k]: v.default }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [schema]);

  if (!schema.properties) return null;

  const handleOnChange = (key: string, value: any) => {
    onChange?.(key, value);
    setCurrentValues((s) => ({
      ...s,
      [key]: value,
    }));
  };

  return (
    <div className="flex flex-col gap-2">
      {Object.entries(schema.properties).map(([key, prop]) => {
        if (!prop.title) return null;
        if (prop.enum) {
          return (
            <Select
              key={key}
              initValue={currentValues[key] ?? prop.default}
              options={prop.enum.map((e: any) => ({ label: e, value: e }))}
              label={prop.title || key}
              onChange={(e) => handleOnChange(key, e)}
            />
          );
        }

        if (prop.type === "string" && prop.format) {
          return (
            <FilePickerField
              key={key}
              label={prop.title || key}
              initValue={currentValues[key] ?? prop.default}
              hint={`PATH/TO/FILE.${prop.format}`}
              onChange={(e) => handleOnChange(key, e)}
            />
          );
        }

        if (prop.type === "boolean") {
          return (
            <span className="flex gap-2">
              <Checkbox
                key={key}
                value={currentValues[key] ?? prop.default}
                onChange={(e) => handleOnChange(key, e)}
              />
              <p className="text-dark-gray">{prop.title || key}</p>
            </span>
          );
        }

        if (prop.type === "array") {
          const listValues = Array.isArray(currentValues[key]) ? currentValues[key] : (prop.default || []);
          const isNumeric = prop.items?.type === "number" || prop.items?.type === "integer";
          return (
            <div key={key} className="flex flex-col gap-2">
              <p className="text-dark-gray">{prop.title || key}</p>
              {listValues.map((val: any, idx: number) => (
                <div key={idx} className="flex gap-2 items-center">
                  <div className="flex-1">
                    <FormField
                      type={isNumeric ? "number" : "text"}
                      initValue={val}
                      onChange={(e) => {
                        const newArray = [...listValues];
                        newArray[idx] = isNumeric ? Number(e) : e;
                        handleOnChange(key, newArray);
                      }}
                    />
                  </div>
                  <Button
                    icon="delete"
                    onClick={() => {
                      const newArray = listValues.filter((_: any, i: any) => i !== idx);
                      handleOnChange(key, newArray);
                    }}
                  />
                </div>
              ))}
              <Button
                label="Add Item"
                icon="add"
                onClick={() => handleOnChange(key, [...listValues, ""])}
              />
            </div>
          );
        }

        const isNumeric = prop.type === "number" || prop.type === "integer";
        return (
          <FormField
            key={key}
            type={isNumeric ? "number" : "text"}
            initValue={currentValues[key] ?? prop.default}
            label={prop.title || key}
            onChange={(e) => handleOnChange(key, isNumeric ? Number(e) : e)}
          />
        );
      })}
    </div>
  );
}
