/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useMemo, useState } from "react";
import { useShallow } from "zustand/shallow";
import { domain } from "../../wailsjs/go/models";
import Button from "../components/button";
import { useLogs } from "../hooks/useLogs";
import { cn } from "../utils/cn";

export default function LogsPage() {
  const [loadAll, logPaths, selectLog, selectedIdx] = useLogs(
    useShallow((s) => [s.loadAll, s.logPaths, s.selectLog, s.selectedIdx]),
  );

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  const groupedLogs = useMemo(() => {
    const groups: Record<string, { idx: number; name: string }[]> = {};
    logPaths.forEach(([name, path], idx) => {
      const parts = path.replace(/\\/g, "/").split("/");
      let simName = "Unknown";
      if (parts.length >= 3) {
        simName = parts[parts.length - 3];
      }
      if (!groups[simName]) groups[simName] = [];
      groups[simName].push({ idx, name });
    });
    return groups;
  }, [logPaths]);

  return (
    <div className="h-[calc(100vh-60px)] flex overflow-clip">
      <aside className="bg-cwhite border-r border-border h-full flex flex-col overflow-hidden w-70 select-none shrink-0">
        <header className="bg-gray border-b border-border px-3 flex items-center justify-between py-2">
          <h4 className="text-cblack font-medium">Logs</h4>
          <Button icon="refresh" onClick={loadAll} />
        </header>
        <main className="p-3 flex-1 flex flex-col gap-4 overflow-y-auto">
          {Object.entries(groupedLogs).map(([simName, logs]) => (
            <div key={simName} className="flex flex-col gap-1.5">
              <h5 className="text-xs font-bold text-dark-gray uppercase tracking-wider px-1">
                {simName}
              </h5>
              <div className="flex flex-col gap-1">
                {logs.map((log) => (
                  <LogPathCard
                    key={log.idx}
                    name={log.name.replace(".zip", "")}
                    selected={selectedIdx === log.idx}
                    onClick={() =>
                      selectLog(selectedIdx === log.idx ? undefined : log.idx)
                    }
                  />
                ))}
              </div>
            </div>
          ))}
          {logPaths.length === 0 && (
            <p className="text-sm text-dark-gray italic px-2">No logs found.</p>
          )}
        </main>
      </aside>
      <main className="flex-1 flex flex-col overflow-hidden bg-cwhite text-cblack relative">
        <FilterBar />
        <LogConsole />
      </main>
    </div>
  );
}

function LogPathCard({
  name,
  selected,
  onClick,
}: {
  name: string;
  selected: boolean;
  onClick(): void;
}) {
  return (
    <div
      onClick={onClick}
      className={cn(
        `hoverable-gray text-cblack flex items-center gap-2 px-3 py-1.5 cursor-pointer rounded-md`,
        {
          "bg-primary text-white hover:bg-primary": selected,
        },
      )}
    >
      <span className="material-symbols-rounded">folder_zip</span>
      <span className="truncate">{name}</span>
    </div>
  );
}

function FilterBar() {
  const [filter, setFilter, search, isLoading] = useLogs(
    useShallow((s) => [s.filter, s.setFilter, s.search, s.isLoading]),
  );

  const [localFilter, setLocalFilter] = useState<domain.LogFilter>(filter);

  // Sync local if global changes
  useEffect(() => {
    setLocalFilter(filter);
  }, [filter]);

  const handleSearch = () => {
    setFilter(localFilter);
    search();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearch();
    }
  };

  const updateField = (field: keyof domain.LogFilter, val: string) => {
    setLocalFilter((prev) => {
      const next = new domain.LogFilter(prev);
      (next as any)[field] = val;
      return next;
    });
  };

  return (
    <div className="bg-gray border-b border-border p-2 flex items-center gap-3 text-sm shrink-0 flex-wrap">
      <div className="flex items-center gap-2">
        <span className="text-dark-gray font-medium">Level:</span>
        <select
          value={localFilter.Level}
          onChange={(e) => updateField("Level", e.target.value)}
          className="bg-cwhite text-cblack border border-border rounded px-2 py-1 outline-none focus:border-primary"
        >
          <option value="">ALL</option>
          <option value="INFO">INFO</option>
          <option value="WARN">WARN</option>
          <option value="ERROR">ERROR</option>
          <option value="DEBUG">DEBUG</option>
        </select>
      </div>

      <div className="flex items-center gap-2">
        <span className="text-dark-gray font-medium">Instance:</span>
        <input
          type="text"
          value={localFilter.InstanceID}
          onChange={(e) => updateField("InstanceID", e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="e.g. netsim_1"
          className="bg-cwhite text-cblack border border-border rounded px-2 py-1 w-24 outline-none focus:border-primary"
        />
      </div>

      <div className="flex items-center gap-2">
        <span className="text-dark-gray font-medium">Service:</span>
        <input
          type="text"
          value={localFilter.ServiceID}
          onChange={(e) => updateField("ServiceID", e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="e.g. application"
          className="bg-cwhite text-cblack border border-border rounded px-2 py-1 w-28 outline-none focus:border-primary"
        />
      </div>

      <div className="flex items-center gap-2">
        <span className="text-dark-gray font-medium">Event:</span>
        <input
          type="text"
          value={localFilter.EventID}
          onChange={(e) => updateField("EventID", e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="e.g. CMD_START"
          className="bg-cwhite text-cblack border border-border rounded px-2 py-1 w-28 outline-none focus:border-primary"
        />
      </div>

      <div className="flex items-center gap-2 flex-1">
        <span className="text-dark-gray font-medium">Search:</span>
        <input
          type="text"
          value={localFilter.SearchText}
          onChange={(e) => updateField("SearchText", e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Search in message..."
          className="bg-cwhite text-cblack border border-border rounded px-2 py-1 w-full outline-none focus:border-primary"
        />
      </div>

      <button
        onClick={handleSearch}
        disabled={isLoading}
        className="bg-primary hover:bg-blue-600 text-white px-4 py-1 rounded flex items-center gap-2 transition-colors disabled:opacity-50"
      >
        <span className="material-symbols-rounded text-sm">search</span>
        {isLoading ? "Searching..." : "Search"}
      </button>
    </div>
  );
}

function LogConsole() {
  const [messages, selectedIdx] = useLogs(
    useShallow((s) => [s.messages, s.selectedIdx]),
  );

  if (selectedIdx === undefined) {
    return (
      <div className="flex-1 flex items-center justify-center text-dark-gray font-mono">
        Select a log archive from the sidebar
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center text-dark-gray font-mono">
        No logs match the current filters
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 font-mono text-sm leading-relaxed bg-background text-cblack">
      {messages.map((msg, i) => (
        <LogLine key={i} msg={msg} />
      ))}
    </div>
  );
}
function getServiceColor(serviceId: string, isDark: boolean) {
  let hash = 0;
  for (let i = 0; i < serviceId.length; i++) {
    hash = serviceId.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  const lightness = isDark ? 30 : 45;
  const borderLightness = isDark ? 45 : 35;

  return {
    backgroundColor: `hsl(${hue}, 65%, ${lightness}%)`,
    color: "#ffffff",
    borderColor: `hsl(${hue}, 65%, ${borderLightness}%)`,
  };
}
function LogLine({ msg }: { msg: domain.LogMessage }) {
  // Parse date safely
  let dateStr = msg.Timestamp;
  try {
    const d = new Date(msg.Timestamp);
    // Format: HH:MM:SS.mmm
    dateStr = d.toISOString().split("T")[1].replace("Z", "");
  } catch {
    // Ignore
  }

  const isError = msg.Level.toUpperCase() === "ERROR";
  const isWarn = msg.Level.toUpperCase() === "WARN";
  const isInfo = msg.Level.toUpperCase() === "INFO";

  const isDark = document.documentElement.classList.contains("dark");
  const serviceStyle = msg.ServiceID
    ? getServiceColor(msg.ServiceID, isDark)
    : undefined;

  return (
    <div className="flex gap-3 hoverable-gray py-1 px-2 rounded group border-l-4 border-transparent hover:border-border transition-colors">
      <span className="text-dark-gray shrink-0 select-none">[{dateStr}]</span>

      <span
        className={cn("w-12 shrink-0 font-bold", {
          "text-red-500 dark:text-red-400": isError,
          "text-yellow-600 dark:text-yellow-400": isWarn,
          "text-blue-500 dark:text-blue-400": isInfo,
          "text-dark-gray": !isError && !isWarn && !isInfo,
        })}
      >
        {msg.Level.toUpperCase().padEnd(5, " ")}
      </span>

      <div className="flex flex-col gap-0.5 w-full">
        <div className="flex items-center gap-2 text-xs opacity-80 group-hover:opacity-100 transition-opacity select-none">
          {msg.InstanceID && (
            <span className="bg-gray text-cblack border border-border px-1.5 py-0.5 rounded shadow-sm">
              {msg.InstanceID}
            </span>
          )}
          {msg.ServiceID && (
            <span
              className="px-1.5 py-0.5 rounded shadow-sm border"
              style={serviceStyle}
            >
              {msg.ServiceID}
            </span>
          )}
          {msg.EventID && (
            <span className="bg-primary/10 text-primary px-1.5 py-0.5 rounded border border-primary/30 shadow-sm">
              {msg.EventID}
            </span>
          )}
        </div>
        <span
          className={cn("whitespace-pre-wrap break-words mt-0.5", {
            "text-red-700 dark:text-red-300 font-medium": isError,
            "text-cblack": !isError,
          })}
        >
          {msg.Message}
        </span>
      </div>
    </div>
  );
}
