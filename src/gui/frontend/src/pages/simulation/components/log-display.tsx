/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useRef, useState } from "react";
import {
  EventsOff,
  EventsOn,
  EventsOnce,
} from "../../../../wailsjs/runtime/runtime";
import Button from "../../../components/button";
import { useSimulationLog } from "../../../hooks/useSimulationLog";
import { cn } from "../../../utils/cn";

declare global {
  interface Window {
    wails?: any;
    runtime?: any;
  }
}

const EVENT_TAGS = [
  {
    tag: "simulation:log",
    level: "[MSG]",
    levelClass: "text-primary",
    getMsg: (msg: any) => msg as string,
  },
  {
    tag: "netsim:message",
    level: "[MSG]",
    levelClass: "text-amber-600",
    getMsg: (msg: any) => msg.label as string,
  },
  {
    tag: "simulation:ready",
    level: "[SYS]",
    levelClass: "text-green-600",
    getMsg: (_: any) => "All Ready — All UAVs have GPS lock",
  },
];

interface Props {
  className?: string;
  onFinishReceived(): void;
}

export default function LogDisplay({ className, onFinishReceived }: Props) {
  const [isOpen, setIsOpen] = useState(true);
  const logs = useSimulationLog((s) => s.logs);
  const appendLog = useSimulationLog((s) => s.appendLog);
  const logEndRef = useRef<HTMLDivElement | null>(null);

  // Register event listeners once — the log state lives in the Zustand store
  // and survives tab navigation. We only clear it when the simulation exits.
  useEffect(() => {
    if (!window.wails && !window.runtime) {
      console.warn("Wails runtime no detectado. Simulando entorno de navegador.");
      return;
    }

    for (const eventType of EVENT_TAGS) {
      EventsOn(eventType.tag, (msg: any) => {
        appendLog({
          level: eventType.level,
          levelClass: eventType.levelClass,
          msg: eventType.getMsg(msg),
        });
      });
    }

    EventsOnce("simulation:finished", onFinishReceived);

    return () => {
      for (const eventType of EVENT_TAGS) {
        EventsOff(eventType.tag);
      }
    };
  }, [onFinishReceived, appendLog]);

  useEffect(() => {
    if (!isOpen) return;
    logEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs, isOpen]);

  return (
    <div
      className={cn(
        `h-1/3 bg-cwhite
          z-50 rounded-md border-border border overflow-y-auto`,
        className,
        {
          "h-min": !isOpen,
        },
      )}
    >
      <span
        onClick={() => setIsOpen((s) => !s)}
        className={cn(
          `bg-gray border-b border-border flex items-center 
          justify-between p-1 pl-3 cursor-pointer sticky top-0`,
          { "border-transparent": !isOpen },
        )}
      >
        <p className="text-base font-medium">Log</p>
        <Button icon={isOpen ? "keyboard_arrow_down" : "keyboard_arrow_up"} />
      </span>
      {isOpen && (
        <div className="h-full overlow-y-auto">
          {logs.map((log, i) => {
            return (
              <div key={i} className="flex flex-row gap-2 px-2 items-center text-gray text-lg">
                <span className="text-cblack font-bold">{log.time}</span>
                <span className={cn("font-bold", log.levelClass)}>
                  {log.level}
                </span>
                <p className="text-dark-gray">{log.msg}</p>
              </div>
            );
          })}
          <div ref={logEndRef} />
        </div>
      )}
    </div>
  );
}
