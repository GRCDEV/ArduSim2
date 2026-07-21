import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FormField from "../../../components/form-field";
import { useConfig } from "../../../hooks/useConfig";

export default function SimNameCard() {
  const { simulationName } = useConfig((s) => s.config);
  const update = useConfig((s) => s.update);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle label="Simulation Name" icon="edit_document" />
      <FormField
        hint="Simulation name"
        initValue={simulationName}
        onChange={(e) => {
          update((s) => ({
            ...s,
            simulationName: e,
          }));
        }}
      />
    </Card>
  );
}
