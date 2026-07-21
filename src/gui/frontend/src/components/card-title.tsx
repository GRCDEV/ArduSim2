import InfoTooltip from "./info-tooltip";

interface Props {
  icon: string;
  label: string;
  tooltip?: string;
}

export default function CardTitle({ icon, label, tooltip }: Props) {
  return (
    <span className="flex gap-2 items-center">
      <span className="material-symbols-rounded text-primary">{icon}</span>
      <h5 className="font-medium text-xl">{label}</h5>
      {tooltip && <InfoTooltip text={tooltip} />}
    </span>
  );
}
