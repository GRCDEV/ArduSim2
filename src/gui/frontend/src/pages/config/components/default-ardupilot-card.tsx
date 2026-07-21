import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FilePickerField from "../../../components/file-picker-field";
import { useConfig } from "../../../hooks/useConfig";
import { SelectArduPilotInstance } from "../../../../wailsjs/go/main/App";

export default function DefaultArduPilotCard() {
  const update = useConfig((s) => s.update);
  const { defaultArduPilotInstance } = useConfig((s) => s.config);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Default ArduPilot Instance" icon="memory" />
      <FilePickerField
        label="Instance Path"
        hint="Default SITL instance will be used if empty"
        initValue={defaultArduPilotInstance ?? ""}
        onChange={(val) => update((s) => ({ ...s, defaultArduPilotInstance: val }))}
        onBrowse={SelectArduPilotInstance}
      />
    </Card>
  );
}
