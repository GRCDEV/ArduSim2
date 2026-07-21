import { useEffect, useState } from "react";
import { useShallow } from "zustand/shallow";
import Card from "../../../components/card";
import { getUavColor } from "../../../constants/uav-colors";
import { useMap } from "../../../hooks/useMap";
import { useSimulationSession } from "../../../hooks/useSimulationSession";
import { useSwarms } from "../../../hooks/useSwarms";
import { useTelemetry, type TelemetryData } from "../../../hooks/useTelemetry";
import { cn } from "../../../utils/cn";
import { formatTime } from "../../../utils/format-time";

export default function UavTelemetryDisplay() {
  const swarms = useSwarms((s) => s.swarms);
  const fleetUavs = swarms.flatMap(s => s.uavs);
  const getUavs = useTelemetry((s) => s.interpolatedUavs);
  const [setupTime, simulationTime] = useSimulationSession(
    useShallow((s) => [s.setupTime, s.simulationTime]),
  );
  const toggleFollowTarget = useMap((s) => s.toggleFollowTarget);

  // Pull fresh data from the telemetry closure at ~10fps.
  // interpolatedUavs() reads rawData.current which updates without Zustand set(),
  // so we need our own loop to detect changes and trigger React re-renders.
  const [snapshot, setSnapshot] = useState<Record<string, TelemetryData>>({});
  useEffect(() => {
    let rafId: number;
    let lastUpdate = 0;
    const tick = (time: number) => {
      if (time - lastUpdate > 100) {
        setSnapshot({ ...getUavs() });
        lastUpdate = time;
      }
      rafId = requestAnimationFrame(tick);
    };
    rafId = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafId);
  }, [getUavs]);

  return (
    <div
      className="h-full bg-gray w-1/3 max-w-70 border-l 
      border-border shadow-sm flex flex-col"
    >
      <div className="flex-1 overflow-y-auto p-3 flex flex-col gap-2">
        {fleetUavs.map((e, i) => (
          <TelemetryCard
            key={i}
            uav_id={e.id}
            data={snapshot[e.id]}
            onClick={() => toggleFollowTarget(e.id)}
          />
        ))}
      </div>
      <div className="flex flex-col gap-2 p-3">
        <TimeCard time={setupTime} label="Setup" />
        <TimeCard time={setupTime + simulationTime} label="Total" />
      </div>
    </div>
  );
}

function TimeCard({ time, label }: { time: number; label: string }) {
  return (
    <Card className="text-dark-gray font-medium flex gap-2 justify-between px-3 py-1.5">
      {label} <p className="text-cblack font-normal">{formatTime(time)}</p>
    </Card>
  );
}

interface TelemetryCardProps {
  uav_id: string;
  data?: TelemetryData;
  onClick?(): void;
}

function TelemetryCard({ uav_id, data, onClick }: TelemetryCardProps) {
  const p = data;
  const color = getUavColor(Number(uav_id));
  const speed = p
    ? Math.sqrt(p.speed.vx ** 2 + p.speed.vy ** 2 + p.speed.vz ** 2)
    : 0;
  return (
    <div onClick={() => (p ? onClick?.() : null)}>
      <Card
        className={cn(
          `flex flex-col gap-1.5 cursor-pointer hoverable-opacity transition-all duration-200`,
          { "opacity-50 pointer-events-none": !p },
        )}
      >
        <header className="flex gap-2 text-lg font-medium items-cente">
          <span
            className="material-symbols-rounded pt-0.5"
            style={{ fontVariationSettings: "'FILL' 100", color: color }}
          >
            drone
          </span>
          <p>Uav {uav_id}</p>
        </header>
        <main className="flex flex-col gap-1.5">
          <span className="flex gap-2">
            <ParamCard
              label="LAT"
              value={p?.position.lat.toFixed(6) ?? "0"}
              unit="º"
            />
            <ParamCard
              label="LON"
              value={p?.position.lon.toFixed(6) ?? "0"}
              unit="º"
            />
          </span>
          <span className="flex gap-2">
            <ParamCard
              label="ALT"
              value={p?.position.alt.toFixed(1) ?? "0"}
              unit="m"
            />
            <ParamCard label="SPD" value={speed.toFixed(1)} unit="m/s" />
            <BatteryCard value={p?.battery ?? 0} color={color} />
          </span>
        </main>
        <footer>
          <p className="text-sm leading-4 mt-2 text-dark-gray w-full ">
            {p?.flight_mode ?? ""}
          </p>
        </footer>
      </Card>
    </div>
  );
}

function BatteryCard({ value, color }: { value: number; color: string }) {
  return (
    <div className="flex-1">
      <p className="text-dark-gray font-medium text-sm">BAT</p>
      <span className="flex gap-2 items-center">
        <p className="leading-4">{value}%</p>
        <div className="h-1 w-full bg-gray rounded-full overflow-clip">
          <div
            style={{
              width: `${value}%`,
              backgroundColor: color,
              height: "100%",
            }}
          />
        </div>
      </span>
    </div>
  );
}

function ParamCard({
  label,
  value,
  unit,
}: {
  label: string;
  value: string;
  unit: string;
}) {
  return (
    <div className="flex-1">
      <p className="text-dark-gray font-medium text-sm">{label}</p>
      <p className="leading-4">
        {value}
        <span className="text-dark-gray"> {unit}</span>
      </p>
    </div>
  );
}
