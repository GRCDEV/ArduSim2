import type { domain } from "../../wailsjs/go/models";
import Button from "./button";

export default function ServiceCard({
  service,
  onSelect,
  onDelete,
  showId,
  isModified,
}: {
  service: domain.DeployedService;
  onSelect?(): void;
  onDelete?(): void;
  showId?: boolean;
  isModified?: boolean;
}) {
  return (
    <div
      className="border border-border rounded-md hoverable-gray shadow-sm
      flex justify-between items-center bg-cwhite px-3 py-1.5 cursor-pointer"
    >
      <span className="text-lg">
        {service.serviceTitle}
        {isModified && <span className="text-primary font-bold ml-1">*</span>}
        <p className="text-gray-500 text-sm">{showId && service.serviceId ? ` (${service.serviceId})` : ""}</p>
      </span>
      <span className="flex gap-2">
        <Button icon="edit" onClick={onSelect} />
        {onDelete && <Button icon="close" onClick={onDelete} />}
      </span>
    </div>
  );
}
