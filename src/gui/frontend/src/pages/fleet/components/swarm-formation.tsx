import { useEffect, useMemo, useState } from "react";
import { GetKmlFirstCoordinate } from "../../../../wailsjs/go/main/App";
import type { domain } from "../../../../wailsjs/go/models";
import CardTitle from "../../../components/card-title";
import FormField from "../../../components/form-field";
import Select from "../../../components/select";
import { useSwarms, type UAV } from "../../../hooks/useSwarms";
import { useServices } from "../../../hooks/useServices";
import { cn } from "../../../utils/cn";

const FORMATIONS = [
  {
    value: "LINEAR",
    icon: "linear_scale",
  },
  {
    value: "MATRIX",
    icon: "grid_view",
  },
  {
    value: "CIRCLE",
    icon: "circle",
  },
  {
    value: "RANDOM",
    icon: "scatter_plot",
  },
];

export default function SwarmFormation() {
  const [kmlCoords, setKmlCoords] = useState<domain.Coordinate | undefined>(undefined);

  const activeSwarmIdx = useSwarms((s) => s.activeSwarmIdx);
  const swarms = useSwarms((s) => s.swarms);
  const updateSwarm = useSwarms((s) => s.updateSwarm);
  
  const swarm = swarms[activeSwarmIdx];
  const {
    formationCenterLat,
    formationCenterLon,
    formationSpacing,
    formationCenterMode,
  } = swarm || {};

  const services = useServices((s) => s.services);

  const kmlFiles = useMemo<{ filename: string; path: string }[]>(
    () => swarm ? getKmlFiles(services, swarm.uavs) : [],
    [swarm, services],
  );

  useEffect(() => {
    useServices.getState().loadServices();
  }, []);

  useEffect(() => {
    if (!formationCenterMode) {
      setKmlCoords(undefined);
      return;
    }
    GetKmlFirstCoordinate(formationCenterMode).then(setKmlCoords);
  }, [formationCenterMode]);

  if (!swarm) return null;

  const handleSelectCoordsSrc = async (value: string | undefined) => {
    if (!value) {
      setKmlCoords(undefined);
      updateSwarm(activeSwarmIdx, { formationCenterMode: value });
      return;
    }

    const coords = await GetKmlFirstCoordinate(value);
    setKmlCoords(coords);
    updateSwarm(activeSwarmIdx, {
      formationCenterMode: value,
      formationCenterLat: coords.lat,
      formationCenterLon: coords.lon,
    });
  };

  return (
    <aside className="border-l border-border min-w-70 max-w-90 flex-1/4 bg-cwhite">
      <span className="border-b border-border flex flex-row items-center justify-between px-3 py-3 bg-gray">
        <CardTitle label={`Swarm ${swarm.id}`} icon="grid_3x3" />
      </span>
      <div className="p-2 flex flex-col gap-2">
        <div>
          <label className="text-dark-gray flex items-center gap-1 mb-1">
            Ground Formation
          </label>
          <FormationModeSelection />
        </div>
        <FormField
          initValue={`${formationSpacing}`}
          type="number"
          min={0}
          label="Formation Spacing"
          suffix={<p>m</p>}
          onChange={(e) =>
            updateSwarm(activeSwarmIdx, { formationSpacing: Number(e) })
          }
        />
        <Select<string>
          key={kmlFiles.map((e) => e.path).join(",")}
          options={[
            { label: "Custom", value: "" },
            ...kmlFiles.map((e) => ({ label: e.filename, value: e.path })),
          ]}
          label="Get coords from"
          initValue={formationCenterMode ?? ""}
          onChange={handleSelectCoordsSrc}
        />
        <span className="flex gap-2">
          <FormField
            initValue={`${kmlCoords?.lat ?? formationCenterLat}`}
            type="number"
            min={0}
            label="Latitude (deg)"
            enabled={kmlCoords === undefined}
            onChange={(e) =>
              updateSwarm(activeSwarmIdx, { formationCenterLat: Number(e) })
            }
          />
          <FormField
            initValue={`${kmlCoords?.lon ?? formationCenterLon}`}
            type="number"
            min={0}
            label="Longitude (deg)"
            enabled={kmlCoords === undefined}
            onChange={(e) =>
              updateSwarm(activeSwarmIdx, { formationCenterLon: Number(e) })
            }
          />
        </span>
      </div>
    </aside>
  );
}

function FormationModeSelection() {
  const activeSwarmIdx = useSwarms((s) => s.activeSwarmIdx);
  const swarms = useSwarms((s) => s.swarms);
  const updateSwarm = useSwarms((s) => s.updateSwarm);
  
  const swarm = swarms[activeSwarmIdx];
  const groundFormation = swarm?.groundFormation;

  return (
    <div className="grid grid-cols-2 grid-rows-2 gap-2">
      {FORMATIONS.map((e) => (
        <div
          key={e.value}
          onClick={() => updateSwarm(activeSwarmIdx, { groundFormation: e.value })}
          className={cn(
            `border border-border rounded-md p-3 overflow-clip justify-center
              bg-cwhite shadow-xs cursor-pointer hoverable-gray flex flex-col items-center `,
            {
              "bg-primary text-onPrimary hoverable-primary":
                e.value === groundFormation,
            },
          )}
        >
          <span className="material-symbols-rounded">{e.icon}</span>
          <p>{e.value}</p>
        </div>
      ))}
    </div>
  );
}

function getKmlFiles(services: domain.ServiceType[], uavs: UAV[]) {
  const kmlFiles: { filename: string; path: string }[] = [];

  const servicesWithKmlFiles = services.filter((e) =>
    e.schemaRaw.includes('"format"') && e.schemaRaw.includes("kml"),
  );

  for (const uav of uavs) {
    const servicesToSearch = uav.services.filter((e) =>
      servicesWithKmlFiles.find((v) => v.id === e.serviceId),
    );
    for (const serv of servicesToSearch) {
      for (const prop of Object.values(serv.config)) {
        if (typeof prop === "string" && prop.includes(".kml")) {
          kmlFiles.push({
            path: prop as string,
            filename: (prop as string).split(/[/\\]/).pop() ?? "",
          });
        }
      }
    }
  }
  return kmlFiles;
}
