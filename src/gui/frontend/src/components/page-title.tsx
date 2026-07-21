interface Props {
  title: string;
  subtitle?: string;
}

export default function PageTitle({ title, subtitle }: Props) {
  return (
    <div>
      <h2 className="text-4xl font-extrabold">{title}</h2>
      <h4 className="text-lg text-dark-gray leading-4">{subtitle}</h4>
    </div>
  );
}
