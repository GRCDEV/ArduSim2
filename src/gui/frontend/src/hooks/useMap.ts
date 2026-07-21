/* eslint-disable @typescript-eslint/no-explicit-any */
import { AmbientLight, LightingEffect, _SunLight as SunLight } from "@deck.gl/core";
import { PathLayer } from "@deck.gl/layers";
import { MapboxOverlay } from "@deck.gl/mapbox";
import { SimpleMeshLayer } from "@deck.gl/mesh-layers";
import { OBJLoader } from "@loaders.gl/obj";
import maplibregl from "maplibre-gl";
import { create } from "zustand";
import { getRgbUavColor } from "../constants/uav-colors";
import type { TelemetryData } from "./useTelemetry";

// Internal data out of React state
const trailsData: Record<string, { id: string; path: number[][]; color: number[] }> = {};
const markersData: Record<string, { id: string; position: number[]; color: number[] }> = {};
let overlayRef: MapboxOverlay | null = null;
let showTrailsRef = true;

const ambientLight = new AmbientLight({
  color: [255, 255, 255],
  intensity: 1.5,
});
const sunLight = new SunLight({
  color: [255, 255, 255],
  intensity: 2,
  timestamp: 0,
});
const lightingEffect = new LightingEffect({ ambientLight, sunLight });

function refreshLayers() {
  if (!overlayRef) return;
  overlayRef.setProps({
    layers: [
      showTrailsRef &&
        new PathLayer({
          id: "uav-path-layer",
          data: Object.values(trailsData),
          getPath: (d: any) => d.path,
          getColor: (d: any) => d.color,
          widthMinPixels: 2,
          widthUnits: "meters",
          getWidth: 0.25,
          jointRounded: true,
          capRounded: true,
          billboard: true,
          opacity: 0.5,
        }),
      new SimpleMeshLayer({
        id: "uav-mesh-layer",
        data: Object.values(markersData),
        mesh: "/uav1.obj",
        loaders: [OBJLoader],
        getPosition: (d: any) => d.position.slice(0, 3),
        getColor: (d: any) => d.color,
        getScale: [10, 10, 10],
        getOrientation: (d: any) => [0, -d.position[3] || 0, 90],
        _lighting: "phong",
        autoHighlight: true,
      }),
    ].filter(Boolean),
  });
}

interface State {
  showTerrain: boolean;
  showBuildings: boolean;

  toggleTerrain: () => void;
  toggleBuildings: () => void;

  showTrails: boolean;
  followTarget?: string;
  map: maplibregl.Map | null;
  mode2D: boolean;
  setMode2D(value?: boolean): void;

  uavMarkers: Record<string, { id: string; position: number[]; color: number[] }>;

  toggleShowTrails(): void;
  /** One-shot: flies to the given UAV and immediately clears the follow state. */
  toggleFollowTarget(value?: string): void;
  /** Flies to the first UAV with a valid position. Used on "All Ready". */
  flyToFirstUav(): void;
  updateTrails(uavId: string, lat: number, lon: number, alt: number): void;
  updateMarkers(uavs: TelemetryData[]): void;

  init(container: HTMLDivElement): any;
  reset(): void;
}

export const useMap = create<State>((set, get) => ({
  showTrails: true,
  showBuildings: true,
  showTerrain: false,

  mode2D: false,
  map: null,
  uavMarkers: {},

  toggleShowTrails: () => {
    const next = !get().showTrails;
    showTrailsRef = next;
    refreshLayers();
    set({ showTrails: next });
  },

  toggleTerrain: () => {
    const { map, showTerrain } = get();
    const next = !showTerrain;
    if (map) {
      map.setTerrain(
        next ? { source: "terrain-source", exaggeration: 1.5 } : null,
      );
    }
    set({ showTerrain: next });
  },

  toggleBuildings: () => {
    const { map, showBuildings } = get();
    const next = !showBuildings;
    if (map && map.getLayer("3d-buildings")) {
      map.setLayoutProperty(
        "3d-buildings",
        "visibility",
        next ? "visible" : "none",
      );
    }
    set({ showBuildings: next });
  },

  setMode2D(value?: boolean) {
    const next = value ?? !get().mode2D;
    if (next) {
      get().map?.easeTo({ pitch: 0 });
    }
    set({ mode2D: next });
  },

  /** One-shot center: flies to the UAV once, does not lock the map. */
  toggleFollowTarget(value?: string) {
    const { map } = get();
    if (value && map && markersData[value]) {
      const pos = markersData[value].position;
      map.flyTo({ center: [pos[0], pos[1]], zoom: 18 });
    }
    // Mark as active for the button highlight, then clear immediately after
    // the animation ends so the map is never locked.
    set({ followTarget: value });
    map?.once('moveend', () => set({ followTarget: undefined }));
  },

  flyToFirstUav() {
    const { map } = get();
    if (!map) return;
    const first = Object.values(markersData)[0];
    if (!first || first.position.length < 2) return;
    map.flyTo({ center: [first.position[0], first.position[1]], zoom: 18 });
  },

  updateMarkers(uavs: TelemetryData[]) {
    let changed = false;
    for (const uav of uavs) {
      const pos = uav.position;
      // Skip UAVs that haven't acquired a valid GPS fix yet, or have corrupt data
      if (
        pos.lat === 0 && pos.lon === 0 && pos.alt === 0 ||
        isNaN(pos.lat) || isNaN(pos.lon) || isNaN(pos.alt)
      ) continue;

      let marker = markersData[uav.uav_id!];
      if (!marker) {
        marker = {
          color: getRgbUavColor(Number(uav.uav_id!)),
          id: uav.uav_id!,
          position: [],
        };
        markersData[uav.uav_id!] = marker;
        changed = true;
      }

      marker.position = [pos.lon, pos.lat, pos.alt, pos.heading];
    }

    if (Object.keys(markersData).length > 0) {
      refreshLayers();
      if (changed) {
        set({ uavMarkers: { ...markersData } });
      }
    }
  },

  updateTrails(uavId: string, lat: number, lon: number, alt: number) {
    if (lat === 0 && lon === 0 && alt === 0) return;
    if (isNaN(lat) || isNaN(lon) || isNaN(alt)) return;

    if (!trailsData[uavId]) {
      trailsData[uavId] = {
        id: uavId,
        color: getRgbUavColor(Number(uavId)),
        path: [],
      };
    }
    const { path } = trailsData[uavId];
    const last = path[path.length - 1];
    if (!last || last[0] !== lon || last[1] !== lat || last[2] !== alt) {
      path.push([lon, lat, alt]);
      refreshLayers();
    }
  },

  init(container: HTMLDivElement) {
    const map = new maplibregl.Map({
      container: container,
      style: "https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json",
      center: [-0.349228, 39.481645],
      zoom: 13,
      maxPitch: 85,
      pitchWithRotate: true,
      aroundCenter: false,
    });

    map.on("load", () => {
      map.addSource("terrain-source", {
        type: "raster-dem",
        tiles: [
          "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png",
        ],
        encoding: "terrarium",
        tileSize: 256,
        maxzoom: 14,
      });

      if (map.getSource("carto")) {
        map.addLayer({
          id: "3d-buildings",
          source: "carto",
          "source-layer": "building",
          type: "fill-extrusion",
          minzoom: 15,
          paint: {
            "fill-extrusion-color": "#d1d5db",
            "fill-extrusion-height": [
              "coalesce",
              ["get", "render_height"],
              ["get", "height"],
              15,
            ],
            "fill-extrusion-base": [
              "coalesce",
              ["get", "render_min_height"],
              ["get", "min_height"],
              0,
            ],
            "fill-extrusion-opacity": 0.5,
            "fill-extrusion-vertical-gradient": true,
          },
        });
      }

      map.setLight({
        anchor: "viewport",
        color: "white",
        intensity: 0.3,
        position: [1.15, 90, 40],
      });

      overlayRef = new MapboxOverlay({
        interleaved: true,
        effects: [lightingEffect],
        layers: [],
      });
      map.addControl(overlayRef as any);

      refreshLayers();

      // Ensure terrain/buildings sync
      if (get().showTerrain) {
        map.setTerrain({ source: "terrain-source", exaggeration: 1.5 });
      }
      if (!get().showBuildings && map.getLayer("3d-buildings")) {
        map.setLayoutProperty("3d-buildings", "visibility", "none");
      }
    });

    map.on('pitch', () => {
      if (map.getPitch() > 0 && get().mode2D) {
        set({ mode2D: false });
      }
    });



    map.addControl(
      new maplibregl.NavigationControl({
        visualizePitch: true,
        visualizeRoll: true,
        showZoom: false,
        showCompass: true,
      }),
    );

    set({ map: map });

    return () => {
      map.remove();
      overlayRef = null;
      set({ map: null });
    };
  },

  reset() {
    for (const key in trailsData) delete trailsData[key];
    for (const key in markersData) delete markersData[key];
    set({ uavMarkers: {}, followTarget: undefined });
    refreshLayers();
  },
}));
