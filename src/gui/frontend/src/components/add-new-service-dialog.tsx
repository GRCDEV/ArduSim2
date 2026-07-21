/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";
import { domain } from "../../wailsjs/go/models";
import Button from "./button";
import CardTitle from "./card-title";
import DynamicForm from "./dynamic-form";
import FormField from "./form-field";
import { useServices } from "../hooks/useServices";
import { cn } from "../utils/cn";

interface Props {
  title?: string;
  servicesList: domain.ServiceType[];
  onExit?(): void;
  serviceToEdit?: domain.DeployedService;
  onAccept?(service?: domain.DeployedService): void;
}

export default function AddNewServiceDialog({
  title = "Add new service",
  servicesList,
  onExit,
  serviceToEdit,
  onAccept,
}: Props) {
  const [serviceIdx, setServiceIdx] = useState(() => {
    const idx = servicesList.findIndex((e) => e.id === serviceToEdit?.serviceId);
    if (idx !== -1) return idx;
    if (servicesList.length > 0) return 0;
    return -1;
  });
  const [currentValues, setCurrentValues] = useState<Record<string, any>>(
    serviceToEdit?.config ?? {},
  );
  const [nodeLabel, setNodeLabel] = useState(serviceToEdit?.nodeLabel ?? "");

  useEffect(() => {
    useServices.getState().loadServices();
  }, []);

  return (
    <div className="absolute inset-0 flex h-screen overflow-clip z-60">
      <div
        className="bg-black opacity-60 flex-1/4 cursor-pointer"
        onClick={onExit}
      />
      <div className="bg-background flex-3/4 shadow-lg flex flex-col h-screen">
        <span className="bg-gray border-b border-border p-3 flex justify-between items-center sticky top-0 h-15">
          <CardTitle label={title} icon="add_circle" />
          <Button icon="close" type="outlined" onClick={onExit} />
        </span>
        <div className="flex flex-row flex-1 overflow-hidden">
          <aside className="flex-1/3 h-full border-r border-border max-w-80 flex flex-col">
            <h4 className="text-lg font-bold p-3 bg-gray border-b border-border">
              Available Services
            </h4>
            <div className="flex flex-col gap-2 px-2 pt-2 overflow-y-auto">
              {servicesList.map((e, idx) => (
                <ServiceCard
                  service={e}
                  selected={idx === serviceIdx}
                  onSelect={() => {
                    setServiceIdx(idx);
                    setCurrentValues({});
                    setNodeLabel("");
                  }}
                />
              ))}
            </div>
          </aside>
          <div className="flex-2/3 flex flex-col overflow-hidden">
            <div className="bg-background flex-1 p-5 overflow-y-auto">
              {serviceIdx !== -1 && servicesList[serviceIdx] ? (
                <DynamicForm
                  key={serviceIdx}
                  schemaRaw={servicesList[serviceIdx].schemaRaw}
                  values={currentValues}
                  onChange={(key, value) => {
                    setCurrentValues((s) => ({
                      ...s,
                      [key]: value,
                    }));
                  }}
                />
              ) : null}
              {serviceIdx !== -1 && servicesList[serviceIdx] ? (
                <div className="mt-4 pt-4 border-t border-border">
                  <h4 className="font-semibold text-dark-gray mb-2">Kubernetes Execution</h4>
                  <FormField
                    label="Node Label (Optional)"
                    hint="e.g. node-1"
                    tooltip="If specified, adds a nodeSelector to this service for kubernetes deployments"
                    initValue={nodeLabel}
                    onChange={setNodeLabel}
                  />
                </div>
              ) : null}
            </div>
            <div className="flex justify-end gap-3 p-3 border-t border-border bg-gray h-15 shrink-0">
              <Button
                label="Cancel"
                type="outlined"
                onClick={onExit}
                className="px-5 py-2"
              />
              <Button
                label="Accept"
                type="filled"
                className="px-5 py-2"
                onClick={() => {
                  console.log("hola");
                  onAccept?.(
                    domain.DeployedService.createFrom({
                      instanceId: serviceToEdit?.instanceId ?? crypto.randomUUID(),
                      serviceId: servicesList[serviceIdx].id,
                      folderName: servicesList[serviceIdx].folderName,
                      serviceTitle: servicesList[serviceIdx].title,
                      config: currentValues,
                      nodeLabel: nodeLabel !== "" ? nodeLabel : undefined,
                    }),
                  );
                }}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ServiceCard({
  service,
  onSelect,
  selected,
}: {
  service: domain.ServiceType;
  onSelect?(): void;
  selected?: boolean;
}) {
  return (
    <div
      onClick={onSelect}
      className={cn(
        `border border-border bg-cwhite shadow-sm rounded-md px-3 py-1.5 
        cursor-pointer hoverable-gray`,
        {
          "bg-primary text-onPrimary hoverable-primary": selected,
        },
      )}
    >
      <p>{service.title}</p>
    </div>
  );
}
