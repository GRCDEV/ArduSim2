import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FormField from "../../../components/form-field";
import TabSelector from "../../../components/tab-selector";
import { useConfig } from "../../../hooks/useConfig";
import { cn } from "../../../utils/cn";

const NETSIM_MODES = [
  { value: "realistic", label: "Realistic", component: <div /> },
  { value: "fixed_range", label: "Fixed range", component: <FixedRangeForm /> },
  { value: "unrestricted", label: "No restrictions", component: <div /> },
];

export default function NetsimCard() {
  const netsimMode = useConfig((s) => s.config.netsimMode ?? "realistic");
  const netsimInstances = useConfig((s) => s.config.netsimInstances);
  const update = useConfig((s) => s.update);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Netsim Configuration" icon="wifi_tethering" />
      <TabSelector
        initValue={netsimMode}
        onChange={(e) =>
          update((s) => ({
            ...s,
            netsimMode: e.value,
            netsimMaxRangeM:
              e.value === "fixed_range" ? s.netsimMaxRangeM : null,
          }))
        }
        options={NETSIM_MODES}
      />
      <FormField
        label="Netsim Instances"
        hint="1"
        initValue={String(netsimInstances ?? 1)}
        onChange={(e) =>
          update((s) => ({
            ...s,
            netsimInstances: Math.max(1, parseInt(e) || 1),
          }))
        }
      />
    </Card>
  );
}

function FixedRangeForm() {
  const netsimMaxRangeM = useConfig((s) => s.config.netsimMaxRangeM);
  const update = useConfig((s) => s.update);

  return (
    <div className={cn("mt-3")}>
      <FormField
        label="Maximum range"
        hint="1350"
        type="number"
        min={1}
        initValue={netsimMaxRangeM != null ? String(netsimMaxRangeM) : ""}
        suffix={<p className="w-6 text-dark-gray text-sm">m</p>}
        onChange={(e) => {
          const parsed = parseFloat(e);
          update((s) => ({
            ...s,
            netsimMaxRangeM: e === "" || isNaN(parsed) ? null : parsed,
          }));
        }}
      />
      <p className="text-dark-gray text-xs mt-1">
        Leave empty to use netsim's default config.json value (1350 m).
      </p>
    </div>
  );
}
