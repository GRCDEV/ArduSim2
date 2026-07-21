import { useNavigation, TABS } from "../hooks/useNavigation";
import { useSimulationConfig } from "../hooks/useSimulationConfig";
import { useSimulationSession } from "../hooks/useSimulationSession";
import { useSimulationPersistence } from "../hooks/useSimulationPersistence";
import { cn } from "../utils/cn";
import Button from "./button";

export default function Header() {
  const navigateTo = useNavigation((e) => e.navigateTo);
  const tab = useNavigation((e) => e.current);

  const activeMode = useSimulationConfig((s) => s.lastConfig.value.activeMode);

  const lastConfig = useSimulationConfig((s) => s.lastConfig);
  const undos = useSimulationConfig((s) => s.undos);
  const redos = useSimulationConfig((s) => s.redos);
  const lastHash = useSimulationConfig((s) => s.lastHash);
  const undo = useSimulationConfig((s) => s.undo);
  const redo = useSimulationConfig((s) => s.redo);
  const saveConfig = useSimulationPersistence((s) => s.saveConfig);
  const newConfig = useSimulationPersistence((s) => s.newConfig);

  const startSimulation = useSimulationSession((s) => s.startSimulation);
  const buildImages = useSimulationSession((s) => s.buildImages);
  const exportSimulation = useSimulationSession((s) => s.exportSimulation);

  return (
    <header className="border-b border-border flex gap-4 p-3 items-center bg-cwhite sticky top-0 z-50 h-15">
      <Button icon="note_add" onClick={newConfig} type="filled" />
      <h1 className="text-primary text-xl font-extrabold">ArduSim</h1>
      <span>
        {TABS.map((e) => (
          <a
            key={e.label}
            onClick={() => navigateTo(e)}
            className={cn(
              `decoration-primary underline-offset-3 decoration-2 text-lg 
              hoverable-gray px-2 py-1 rounded-md cursor-pointer`,
              tab.label === e.label ? "underline" : "hoverable-underline",
            )}
          >
            {e.label}
          </a>
        ))}
      </span>
      <div className="flex-1" />
      <span className="flex flex-row gap-1.5">
        <Button icon="undo" onClick={undo} enabled={undos.length !== 0} />
        <Button icon="redo" onClick={redo} enabled={redos.length !== 0} />
        <Button
          icon="save"
          type="outlined"
          onClick={saveConfig}
          enabled={lastConfig.hash !== lastHash}
        />
        <Button
          icon="play_arrow"
          type="filled"
          onClick={startSimulation}
        />
        <Button
          icon="file_download"
          type="filled"
          onClick={exportSimulation}
        />
        <Button
          label={activeMode === "KUBERNETES" ? "Build & Publish images" : "Build images"}
          type="filled"
          onClick={buildImages}
        />
      </span>
    </header>
  );
}
