interface Props {
  text: string;
}

export default function InfoTooltip({ text }: Props) {
  return (
    <div className="group relative flex items-center justify-center ml-1">
      <span className="material-symbols-rounded text-dark-gray hover:text-primary cursor-help text-xl transition-colors">
        info
      </span>
      {/* Tooltip Content */}
      <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 hidden group-hover:block w-72 z-[999] bg-cwhite text-cblack border border-border text-sm font-normal rounded-md p-3 shadow-lg pointer-events-none text-left leading-snug">
        {text}
        {/* Triangle arrow outer (border) */}
        <div 
          className="absolute top-full left-1/2 -translate-x-1/2 border-[8px] border-transparent"
          style={{ borderTopColor: "var(--color-border)" }}
        ></div>
        {/* Triangle arrow inner (background) */}
        <div 
          className="absolute top-[calc(100%-1px)] left-1/2 -translate-x-1/2 border-[8px] border-transparent"
          style={{ borderTopColor: "var(--color-cwhite)" }}
        ></div>
      </div>
    </div>
  );
}
