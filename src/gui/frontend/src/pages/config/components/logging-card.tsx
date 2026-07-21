import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import Checkbox from "../../../components/checkbox";
import { useConfig } from "../../../hooks/useConfig";

export default function LoggingCard() {
  const { loggingEnabled } = useConfig(
    (s) => s.config,
  );
  const update = useConfig((s) => s.update);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Logging" icon="list_alt" />
      <span className="flex flex-row items-center justify-between border border-border rounded-md py-1.5 px-3">
        <div>
          <label className="text-base ">Arducopter logging <span className="text-dark-gray text-xs">(local only)</span></label>
          <p className="text-dark-gray text-sm">
            Enable detailed telemetry dumps
          </p>
        </div>
        <Checkbox
          onChange={(e) => update((s) => ({ ...s, loggingEnabled: e }))}
          value={loggingEnabled}
        />
      </span>
    </Card>
  );
}
