import { useShallow } from "zustand/shallow";
import Button from "../../../components/button";
import Card from "../../../components/card";
import { useMap } from "../../../hooks/useMap";
import { cn } from "../../../utils/cn";

interface Props {
  className?: string;
}

export default function MapControls({ className }: Props) {
  const [
    toggleShowTrails,
    showTrails,
    toggleFollowTarget,
    followTarget,
    uavMarkers,
    mode2D,
    setMode2D,
    showBuildings,
    showTerrain,
  ] = useMap(
    useShallow((s) => [
      s.toggleShowTrails,
      s.showTrails,
      s.toggleFollowTarget,
      s.followTarget,
      s.uavMarkers,
      s.mode2D,
      s.setMode2D,
      s.showBuildings,
      s.showTerrain,
    ]),
  );

  return (
    <Card className={cn("p-1 gap-1 flex", className)}>
      <Button
        onClick={() => {
          const id = Object.values(uavMarkers)[0].id;
          toggleFollowTarget(id);
        }}
        type={followTarget !== undefined ? "filled" : undefined}
        icon="center_focus_strong"
        className="size-10 text-base p-2"
      />
      <Button
        type={showTrails ? "filled" : undefined}
        onClick={toggleShowTrails}
        icon="route"
        className="size-10 text-base p-2"
      />
      <Button
        type={mode2D ? "filled" : undefined}
        onClick={() => setMode2D(true)}
        label="2D"
        className="size-10 text-base p-2 "
      />
      <Button
        type={showBuildings ? "filled" : undefined}
        onClick={useMap.getState().toggleBuildings}
        icon="account_balance"
        className="size-10 text-base p-2 "
      />
      <Button
        type={showTerrain ? "filled" : undefined}
        onClick={useMap.getState().toggleTerrain}
        icon="landscape"
        className="size-10 text-base p-2 "
      />
    </Card>
  );
}
