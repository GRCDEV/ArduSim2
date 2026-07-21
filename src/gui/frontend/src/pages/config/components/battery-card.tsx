import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FormField from "../../../components/form-field";
import { useConfig } from "../../../hooks/useConfig";

export default function BatteryCard() {
  const { batteryCapacity } = useConfig(
    (s) => s.config,
  );
  const update = useConfig((s) => s.update);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Battery" icon="battery_charging_full" />
      <FormField
        label="Default UAV Battery"
        hint="5000"
        type="number"
        initValue={batteryCapacity?.toString()}
        onChange={(e) => update((s) => ({ ...s, batteryCapacity: Number(e) }))}
        enabled={true}
        suffix={<p>mAh</p>}
      />
    </Card>
  );
}
