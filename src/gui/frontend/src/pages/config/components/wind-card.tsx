import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import Checkbox from "../../../components/checkbox";
import FormField from "../../../components/form-field";
import { useConfig } from "../../../hooks/useConfig";
import { cn } from "../../../utils/cn";

export default function WindCard() {
  const { windDirection, windEnabled, windSpeed } = useConfig((s) => s.config);
  const update = useConfig((s) => s.update);

  return (
    <Card
      className={cn("space-y-2", {
        "opacity-40": !windEnabled,
      })}
    >
      <span className="flex flex-row items-start justify-between">
        <CardTitle icon="air" label="Wind System" />
        <Checkbox
          onChange={(e) => update((s) => ({ ...s, windEnabled: e }))}
          value={windEnabled}
        />
      </span>
      <div className="flex flex-row gap-5 items-top">
        <Compass direction={windDirection ?? 0} />
        <div className="w-full space-y-2">
          <FormField
            hint="0"
            enabled={windEnabled}
            label="Direction"
            initValue={windDirection?.toString()}
            suffix={<p className="w-10">DEG</p>}
            type="number"
            min={0}
            max={360}
            onChange={(e) =>
              update((s) => ({ ...s, windDirection: Number(e) }))
            }
          />
          <FormField
            hint="0"
            label="Wind speed"
            enabled={windEnabled}
            initValue={windSpeed?.toString()}
            suffix={<p className="w-10">M/S</p>}
            onChange={(e) => update((s) => ({ ...s, windSpeed: Number(e) }))}
            type="number"
            min={0}
          />
        </div>
      </div>
    </Card>
  );
}

function Compass({ direction }: { direction: number }) {
  return (
    <div className="rounded-full border border-border bg-background aspect-square w-30 h-30 p-1">
      <div className="relative w-full h-full font-bold">
        <CompassArrow direction={direction} />
        <p className="absolute inset-x-0 top-0 text-center">N</p>
        <p className="absolute inset-x-0 bottom-0 text-center">S</p>
        <p className="absolute inset-y-0 left-0 flex items-center p-1">W</p>
        <p className="absolute inset-y-0 right-0 flex items-center p-1">E</p>
      </div>
    </div>
  );
}

function CompassArrow({ direction }: { direction?: number }) {
  return (
    <div
      className="absolute inset-0 flex flex-col items-center justify-center gap-2 transition-all duration-200"
      style={{
        transform: `rotate(${direction}deg)`,
      }}
    >
      <div className="bg-primary rounded-full w-1 flex-1 flex justify-center">
        <div className="bg-cwhite border border-border rounded-md w-2 h-6" />
      </div>
      <div className="bg-dark-gray rounded-full aspect-square w-2" />
      <div className="flex-1" />
    </div>
  );
}
