import Button from "../button";

interface ButtonSpecs {
  type?: "outlined" | "filled";
  label: string;
  action?(): void;
}

export interface ConfirmDialogProps {
  title: string;
  text?: string;
  buttons: ButtonSpecs[];
  onExit?(): void;
}

export default function ConfirmDialog({
  title,
  text,
  buttons,
  onExit,
}: ConfirmDialogProps) {
  return (
    <div className="absolute inset-0 z-50">
      <div className="w-screen h-screen bg-black opacity-80" onClick={onExit} />
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="max-w-1/3 border-border border rounded-lg bg-cwhite overflow-clip p-5">
          <h3 className="text-xl font-medium">{title}</h3>
          {text && <p className="text-dark-gray mt-2">{text}</p>}
          <span className="flex items-center justify-end gap-2 mt-2">
            {buttons.map((e) => (
              <Button
                key={e.label}
                type={e.type}
                label={e.label}
                onClick={e.action}
              />
            ))}
          </span>
        </div>
      </div>
    </div>
  );
}
