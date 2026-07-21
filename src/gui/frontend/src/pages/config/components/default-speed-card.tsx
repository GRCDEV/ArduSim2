import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FormField from "../../../components/form-field";
import { useConfig } from "../../../hooks/useConfig";

export default function DefaultSpeedCard() {
  const update = useConfig((s) => s.update);
  const { defaultUAVSpeed } = useConfig((s) => s.config);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Default UAV Speed" icon="speed" />
      <FormField
        type="number"
        min={0}
        hint="10"
        label="Speed"
        initValue={defaultUAVSpeed?.toString() ?? "10"}
        onChange={(e) => update((s) => ({ ...s, defaultUAVSpeed: Number(e) }))}
        suffix={<p>m/s</p>}
      />
    </Card>
  );
}
