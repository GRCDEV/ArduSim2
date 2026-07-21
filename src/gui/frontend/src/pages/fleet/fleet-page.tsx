import { useState } from "react";
import { SelectArduPilotInstance } from "../../../wailsjs/go/main/App";
import AddNewServiceDialog from "../../components/add-new-service-dialog";
import Button from "../../components/button";
import Checkbox from "../../components/checkbox";
import FilePickerField from "../../components/file-picker-field";
import FormField from "../../components/form-field";
import ServiceCard from "../../components/service-card";
import { useConfig } from "../../hooks/useConfig";
import { useSwarms } from "../../hooks/useSwarms";
import { useServices } from "../../hooks/useServices";
import SwarmFormation from "./components/swarm-formation";
import UavsList from "./components/uavs-list";

export default function FleetPage() {
  const swarms = useSwarms((s) => s.swarms);
  const deleteService = useSwarms((s) => s.deleteService);
  const addService = useSwarms((s) => s.addService);
  const updateService = useSwarms((s) => s.updateService);
  const updateUav = useSwarms((s) => s.updateUav);
  const deleteUav = useSwarms((s) => s.deleteUav);
  const activeSwarmIdx = useSwarms((s) => s.activeSwarmIdx);
  const activeUavIdx = useSwarms((s) => s.activeUavIdx);
  const cloneUav = useSwarms((s) => s.cloneUav);
  const { defaultUAVSpeed, defaultArduPilotInstance, defaultMixer, defaultController, batteryCapacity } = useConfig((s) => s.config);
  const { services, mixers, controllers } = useServices();

  const [serviceIdx, setServiceIdx] = useState<number | undefined>(undefined);
  const [baseServiceType, setBaseServiceType] = useState<"mixer" | "controller" | undefined>(undefined);
  const [isConfigCollapsed, setIsConfigCollapsed] = useState(true);

  const swarm = swarms[activeSwarmIdx];
  const uav = activeUavIdx !== null && swarm ? swarm.uavs[activeUavIdx] : null;

  return (
    <div className="flex h-full">
      <UavsList />
      <div className="flex-1/2 px-3 pt-3 flex flex-col gap-2 pb-5 overflow-y-auto">
        {!uav ? (
          <div className="flex flex-col items-center justify-center h-full text-dark-gray opacity-50">
            <span className="material-symbols-rounded text-6xl mb-4">flight</span>
            <p className="text-xl font-medium">Select a UAV to configure it</p>
          </div>
        ) : (
          <>
            <span className="flex justify-between items-center">
              <h3 className="font-bold text-4xl">Uav {uav.id}</h3>
              <span className="flex gap-2">
                <Button icon="content_copy" type="outlined" onClick={cloneUav} />
                <Button icon="delete" type="outlined" onClick={deleteUav} />
              </span>
            </span>
            <div className="flex flex-col gap-4 mt-4 border border-border p-4 rounded-md shadow-sm">
              <span 
                className="flex justify-between items-center cursor-pointer select-none"
                onClick={() => setIsConfigCollapsed(!isConfigCollapsed)}
              >
                <h4 className="font-semibold text-xl text-dark-gray">UAV Configuration</h4>
                <span className="material-symbols-outlined text-dark-gray">
                  {isConfigCollapsed ? "expand_more" : "expand_less"}
                </span>
              </span>
              {!isConfigCollapsed && (
                <div className="flex flex-col gap-2">
                <span className="flex items-end gap-2">
                  <div className="flex-1">
                    <FormField
                      type="number"
                      label="Speed (m/s)"
                      enabled={uav.speed !== null && uav.speed !== undefined}
                      initValue={`${uav.speed ?? defaultUAVSpeed ?? 10}`}
                      onChange={(val) => updateUav(activeUavIdx!, { speed: Number(val) })}
                    />
                  </div>
                  <span className="flex items-center gap-2 mb-2">
                    <Checkbox
                      value={uav.speed === null || uav.speed === undefined}
                      onChange={(auto) => updateUav(activeUavIdx!, { speed: auto ? undefined : (defaultUAVSpeed ?? 10) })}
                    />
                    <label className="text-dark-gray text-sm">Auto</label>
                  </span>
                </span>

                <span className="flex items-end gap-2">
                  <div className="flex-1">
                    <FormField
                      type="number"
                      label="Battery (mAh)"
                      enabled={uav.batteryCapacity !== null && uav.batteryCapacity !== undefined}
                      initValue={`${uav.batteryCapacity ?? batteryCapacity ?? 5000}`}
                      onChange={(val) => updateUav(activeUavIdx!, { batteryCapacity: Number(val) })}
                    />
                  </div>
                  <span className="flex items-center gap-2 mb-2">
                    <Checkbox
                      value={uav.batteryCapacity === null || uav.batteryCapacity === undefined}
                      onChange={(auto) => updateUav(activeUavIdx!, { batteryCapacity: auto ? undefined : (batteryCapacity ?? 5000) })}
                    />
                    <label className="text-dark-gray text-sm">Auto</label>
                  </span>
                </span>

                <div className="flex flex-col gap-2 mt-2">
                  <label className="text-dark-gray">Home Location Override</label>
                  <div className="flex gap-2">
                    <div className="flex-1">
                      <FormField
                        type="number"
                        label="Latitude"
                        enabled={uav.homeOverride !== null && uav.homeOverride !== undefined}
                        initValue={`${uav.homeOverride?.lat ?? 0}`}
                        onChange={(val) => updateUav(activeUavIdx!, { homeOverride: { lat: Number(val), lon: uav.homeOverride?.lon ?? 0, alt: uav.homeOverride?.alt ?? 0 } as any })}
                      />
                    </div>
                    <div className="flex-1">
                      <FormField
                        type="number"
                        label="Longitude"
                        enabled={uav.homeOverride !== null && uav.homeOverride !== undefined}
                        initValue={`${uav.homeOverride?.lon ?? 0}`}
                        onChange={(val) => updateUav(activeUavIdx!, { homeOverride: { lat: uav.homeOverride?.lat ?? 0, lon: Number(val), alt: uav.homeOverride?.alt ?? 0 } as any })}
                      />
                    </div>
                    <span className="flex items-center gap-2 mb-2 self-end">
                      <Checkbox
                        value={uav.homeOverride === null || uav.homeOverride === undefined}
                        onChange={(auto) => updateUav(activeUavIdx!, { homeOverride: auto ? undefined : ({ lat: 0, lon: 0, alt: 0 } as any) })}
                      />
                      <label className="text-dark-gray text-sm">Auto</label>
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-2 mt-2">
                  <span className="flex items-end gap-2">
                    <div className="flex-1">
                      <div className={uav.arduPilotInstance === null || uav.arduPilotInstance === undefined ? "opacity-50 pointer-events-none" : ""}>
                        <FilePickerField
                          label="ArduPilot Instance Override"
                          hint="Default SITL instance will be used if empty"
                          initValue={uav.arduPilotInstance ?? defaultArduPilotInstance ?? ""}
                          onChange={(val) => updateUav(activeUavIdx!, { arduPilotInstance: val })}
                          onBrowse={SelectArduPilotInstance}
                        />
                      </div>
                    </div>
                    <span className="flex items-center gap-2 mb-2">
                      <Checkbox
                        value={uav.arduPilotInstance === null || uav.arduPilotInstance === undefined}
                        onChange={(auto) => updateUav(activeUavIdx!, { arduPilotInstance: auto ? undefined : (defaultArduPilotInstance ?? "") })}
                      />
                      <label className="text-dark-gray text-sm">Auto</label>
                    </span>
                  </span>
                </div>

                <div className="flex flex-col gap-2 mt-2">
                  <span className="flex items-end gap-2">
                    <div className="flex-1">
                        <FormField
                          label="Node Label (Optional)"
                          hint="e.g. node-1"
                          tooltip="If specified, adds a nodeSelector to this UAV's deployment in Kubernetes"
                          initValue={uav.nodeLabel ?? ""}
                          onChange={(val) => updateUav(activeUavIdx!, { nodeLabel: val !== "" ? val : undefined })}
                        />
                    </div>
                  </span>
                </div>
              </div>
              )}
            </div>
            <span className="flex justify-between items-center mt-5">
              <h5 className="font-semibold text-2xl text-dark-gray">
                Base Services
              </h5>
            </span>
            <ServiceCard
              service={uav.mixer ?? defaultMixer ?? { serviceTitle: "Mixer (Default)" } as any}
              isModified={uav.mixer !== null && uav.mixer !== undefined}
              onSelect={() => setBaseServiceType("mixer")}
            />
            <ServiceCard
              service={uav.controller ?? defaultController ?? { serviceTitle: "Controller (Default)" } as any}
              isModified={uav.controller !== null && uav.controller !== undefined}
              onSelect={() => setBaseServiceType("controller")}
            />

            <span className="flex justify-between items-center mt-5">
              <h5 className="font-semibold text-2xl text-dark-gray">
                Active Services
              </h5>
              <Button
                type="filled"
                icon="add"
                label="Add Service"
                onClick={() => setServiceIdx(-1)}
              />
            </span>
            {uav.services.map((e, idx) => (
              <ServiceCard
                key={idx}
                service={e}
                showId={true}
                onDelete={() => deleteService(e)}
                onSelect={() => {
                  setServiceIdx(idx);
                }}
              />
            ))}
          </>
        )}
      </div>
      <SwarmFormation />
      {serviceIdx !== undefined && activeUavIdx !== null ? (
        <AddNewServiceDialog
          servicesList={services}
          title="Add new service"
          serviceToEdit={
            serviceIdx !== -1
              ? swarms[activeSwarmIdx].uavs[activeUavIdx].services[serviceIdx]
              : undefined
          }
          onExit={() => setServiceIdx(undefined)}
          onAccept={(service) => {
            if (!service) return setServiceIdx(undefined);

            if (serviceIdx === -1) {
              addService(service);
            } else {
              updateService(serviceIdx, service);
            }
            setServiceIdx(undefined);
          }}
        />
      ) : null}
      {baseServiceType !== undefined && activeUavIdx !== null ? (
        <AddNewServiceDialog
          servicesList={baseServiceType === "mixer" ? mixers : controllers}
          title={`Edit Base ${baseServiceType === "mixer" ? "Mixer" : "Controller"}`}
          serviceToEdit={
            baseServiceType === "mixer"
              ? swarms[activeSwarmIdx].uavs[activeUavIdx].mixer ?? defaultMixer
              : swarms[activeSwarmIdx].uavs[activeUavIdx].controller ?? defaultController
          }
          onExit={() => setBaseServiceType(undefined)}
          onAccept={(service) => {
            if (!service) return setBaseServiceType(undefined);

            if (baseServiceType === "mixer") {
              updateUav(activeUavIdx, { mixer: service });
            } else {
              updateUav(activeUavIdx, { controller: service });
            }
            setBaseServiceType(undefined);
          }}
        />
      ) : null}
    </div>
  );
}
