import { useState } from "react";
import Button from "../../../components/button";
import FormField from "../../../components/form-field";
import { useSwarms, type Swarm, type UAV } from "../../../hooks/useSwarms";
import { cn } from "../../../utils/cn";

export default function UavsList() {
  const swarms = useSwarms((s) => s.swarms);
  const addSwarms = useSwarms((s) => s.addSwarms);
  const cloneSwarm = useSwarms((s) => s.cloneSwarm);
  const activeSwarmIdx = useSwarms((s) => s.activeSwarmIdx);
  const activeUavIdx = useSwarms((s) => s.activeUavIdx);
  const setSelectedSwarm = useSwarms((s) => s.setSelectedSwarm);
  const setSelectedUav = useSwarms((s) => s.setSelectedUav);
  const updateSwarm = useSwarms((s) => s.updateSwarm);
  const deleteSwarm = useSwarms((s) => s.deleteSwarm);
  
  const [swarmsToAdd, setSwarmsToAdd] = useState(1);

  return (
    <aside className="border-r border-border min-w-70 max-w-90 flex-1/4 bg-cwhite overflow-y-auto">
      <span className="border-b border-border flex flex-row items-center justify-between px-3 py-3 bg-gray sticky top-0 z-10">
        <h3 className="font-bold text-xl">Swarms</h3>
        <div className="w-40 flex items-end gap-2">
          <FormField
            type="number"
            label="Swarms to Add"
            min={0}
            initValue={swarmsToAdd.toString()}
            onChange={(e) => setSwarmsToAdd(Number(e))}
            hint="1"
            suffix={
              <Button
                icon="add"
                type="filled"
                className="rounded-full"
                onClick={() => addSwarms(swarmsToAdd)}
              />
            }
          />
        </div>
      </span>
      <div className="p-2 flex flex-col gap-2">
        {swarms.map((swarm, idx) => (
          <SwarmAccordion
            key={swarm.id || idx}
            swarm={swarm}
            swarmIdx={idx}
            isActiveSwarm={idx === activeSwarmIdx}
            activeUavIdx={idx === activeSwarmIdx ? activeUavIdx : null}
            onSelectSwarm={() => setSelectedSwarm(idx)}
            onSelectUav={(uIdx) => {
              setSelectedSwarm(idx);
              setSelectedUav(uIdx);
            }}
            onChangeName={(name) => updateSwarm(idx, { id: name })}
            onClone={() => cloneSwarm(idx)}
            onDelete={() => {
              setSelectedSwarm(idx);
              deleteSwarm();
            }}
          />
        ))}
      </div>
    </aside>
  );
}

interface SwarmAccordionProps {
  swarm: Swarm;
  swarmIdx: number;
  isActiveSwarm: boolean;
  activeUavIdx: number | null;
  onSelectSwarm: () => void;
  onSelectUav: (idx: number) => void;
  onChangeName: (name: string) => void;
  onClone: () => void;
  onDelete: () => void;
}

function SwarmAccordion({
  swarm,
  isActiveSwarm,
  activeUavIdx,
  onSelectSwarm,
  onSelectUav,
  onChangeName,
  onClone,
  onDelete,
}: SwarmAccordionProps) {
  const [isExpanded, setIsExpanded] = useState(isActiveSwarm);
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState(swarm.id);

  const addUavs = useSwarms((s) => s.addUavs);

  return (
    <div className="flex flex-col border border-border rounded-md shadow-sm overflow-hidden">
      <div
        className={cn(
          "flex justify-between items-center px-3 py-2 cursor-pointer transition-colors duration-200",
          {
            "bg-primary text-onPrimary": isActiveSwarm && activeUavIdx === null,
            "bg-gray hover:bg-gray-100": !(isActiveSwarm && activeUavIdx === null),
          }
        )}
        onClick={() => {
          onSelectSwarm();
        }}
      >
        <div className="flex items-center gap-2 flex-1">
          <span className="material-symbols-rounded" style={{ fontSize: "18px" }}>
            {isExpanded ? "folder_open" : "folder"}
          </span>
          {isEditing ? (
            <input
              autoFocus
              className="bg-white text-black px-1 rounded text-sm w-24 outline-none"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              onBlur={() => {
                setIsEditing(false);
                onChangeName(editName);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  setIsEditing(false);
                  onChangeName(editName);
                }
              }}
              onClick={(e) => e.stopPropagation()}
            />
          ) : (
            <p className="font-semibold select-none" onDoubleClick={(e) => {
              e.stopPropagation();
              setIsEditing(true);
            }}>
              Swarm {swarm.id}
            </p>
          )}
        </div>
        <div className="flex items-center gap-1">
          <span
            className="material-symbols-rounded hover:opacity-70 p-1"
            title="Duplicate Swarm"
            onClick={(e) => {
              e.stopPropagation();
              onClone();
            }}
          >
            content_copy
          </span>
          <span 
            className="material-symbols-rounded hover:opacity-70 p-1"
            title="Delete Swarm"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
          >
            delete
          </span>
          <span
            className="material-symbols-rounded hover:opacity-70 p-1"
            title="Add UAV"
            onClick={(e) => {
              e.stopPropagation();
              if (!isActiveSwarm) onSelectSwarm();
              setIsExpanded(true);
              addUavs(1);
            }}
          >
            add
          </span>
          <span 
            className="material-symbols-rounded hover:opacity-70 p-1"
            onClick={(e) => {
              e.stopPropagation();
              setIsExpanded(!isExpanded);
            }}
          >
            {isExpanded ? "expand_less" : "expand_more"}
          </span>
        </div>
      </div>
      
      {isExpanded && (
        <div className="flex flex-col bg-cwhite">
          {swarm.uavs.map((e, idx) => (
            <UavCard
              key={e.id || idx}
              uav={e}
              selected={isActiveSwarm && idx === activeUavIdx}
              onClick={() => onSelectUav(idx)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface UavCardProps {
  uav: UAV;
  onClick?(): void;
  selected?: boolean;
}

function UavCard({ uav, onClick, selected }: UavCardProps) {
  return (
    <div
      onClick={() => onClick?.()}
      className={cn(
        `hoverable-gray w-full transition-colors duration-200 flex 
        justify-between items-center px-4 py-2 cursor-pointer border-t border-border`,
        {
          "bg-primary hoverable-primary": selected,
        },
      )}
    >
      <div className="flex-1 ml-4 flex flex-row items-center gap-2">
        <span 
          className={cn("material-symbols-rounded text-dark-gray opacity-50", {
            "text-onPrimary opacity-100": selected,
          })} 
          style={{ fontSize: "18px" }}
        >
          subdirectory_arrow_right
        </span>
        <div className="flex flex-col justify-center">
          <p
            className={cn("text-sm font-medium text-dark-gray", {
              "text-onPrimary": selected,
            })}
          >
            Uav {uav.id}
          </p>
          {uav.services.length > 0 && 
          <span className="flex gap-1 items-center mt-1">
            {uav.services.map((service, idx) => (
              <div
                key={service.serviceId || idx}
                className={cn("rounded-full aspect-square h-1.5 bg-dark-gray", {
                  "bg-onPrimary": selected,
                })}
              />
            ))}
          </span>
          }
        </div>
      </div>
      {selected === true ? (
        <span className="material-symbols-rounded text-onPrimary" style={{ fontSize: "16px" }}>
          chevron_right
        </span>
      ) : null}
    </div>
  );
}
