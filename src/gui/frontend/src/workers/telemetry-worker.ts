import type { InterpolationNode, TelemetryData } from "../hooks/useTelemetry";
import { lerpAngle } from "../utils/lerp-angle";

const nodes: Record<string, InterpolationNode> = {};
let lastPacketTime: number | undefined;
let allReady = false;

const requestFrame =
  typeof self.requestAnimationFrame === "function"
    ? self.requestAnimationFrame.bind(self)
    : (callback: FrameRequestCallback) => setTimeout(callback, 16);

self.onmessage = (e) => {
  if (e.data.type === "ALL_READY") {
    allReady = true;
    return;
  }

  if (e.data.type === "NEW_SNAPSHOT") {
    const { uavs } = e.data;
    const now = performance.now();

    // Compute duration ONCE per snapshot, before iterating over UAVs.
    // Moving this inside the loop caused duration=0 for the 2nd+ UAV
    // because lastPacketTime was already updated by the first UAV.
    const newDuration = lastPacketTime ? now - lastPacketTime : 1000;
    lastPacketTime = now;

    for (const [uavId, data] of Object.entries(uavs) as [
      string,
      TelemetryData,
    ][]) {
      data.uav_id = uavId;

      const savedData = nodes[uavId];

      nodes[uavId] = {
        start: allReady ? (savedData?.end ?? data) : data,
        end: data,
        startTime: now,
        duration: newDuration,
        trailEmited: !savedData,
      };
    }
  }
};

const update = () => {
  const now = performance.now();
  const interpolated: Record<string, TelemetryData> = {};

  for (const id in nodes) {
    const node = nodes[id];
    let t = (now - node.startTime) / node.duration;
    if (t > 1) t = 1; //TODO: continue?

    const iPos = node.start.position;
    const fPos = node.end.position;

    if (!node.trailEmited) {
      node.trailEmited = true;
      if (allReady) {
        self.postMessage({
          type: "POINT_REACHED",
          id,
          lat: iPos.lat,
          lon: iPos.lon,
          alt: iPos.alt,
        });
      }
    }
    interpolated[id] = {
      ...node.end,
      position: {
        lat: (iPos.lat ?? 0) + ((fPos.lat ?? 0) - (iPos.lat ?? 0)) * t,
        lon: (iPos.lon ?? 0) + ((fPos.lon ?? 0) - (iPos.lon ?? 0)) * t,
        alt: (iPos.alt ?? 0) + ((fPos.alt ?? 0) - (iPos.alt ?? 0)) * t,
        relative_alt:
          (iPos.relative_alt ?? 0) + ((fPos.relative_alt ?? 0) - (iPos.relative_alt ?? 0)) * t,
        heading: lerpAngle(iPos.heading ?? 0, fPos.heading ?? 0, t),
      },
    };
  }

  self.postMessage({ type: "TICK", interpolated });
  requestFrame(update);
};

update();
