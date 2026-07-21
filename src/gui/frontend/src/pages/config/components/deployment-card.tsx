import { SelectKubeConfig } from "../../../../wailsjs/go/main/App";
import Card from "../../../components/card";
import CardTitle from "../../../components/card-title";
import FilePickerField from "../../../components/file-picker-field";
import FormField from "../../../components/form-field";
import TabSelector from "../../../components/tab-selector";
import { useConfig } from "../../../hooks/useConfig";

export default function DeploymentCard() {
  const activeMode = useConfig((s) => s.activeMode);
  const setActiveMode = useConfig((s) => s.setActieMode);

  return (
    <Card className="flex flex-col gap-2">
      <CardTitle
        label="Deployment Mode"
        icon="dns"
        tooltip="In Kubernetes mode, nodes must have the ports to be used available and visible, as well as port 6443 so Kubernetes can deploy."
      />
      <TabSelector
        initValue={activeMode}
        onChange={(e) => {
          setActiveMode(e.value);
        }}
        options={[
          { value: "LOCAL", label: "Local", component: <div></div> },
          {
            value: "KUBERNETES",
            label: "Kubernetes",
            component: <KubernetesForm />,
          },
        ]}
      />
    </Card>
  );
}

function KubernetesForm() {
  const { dockerHubRepository, kubeConfigPath } = useConfig((s) => s.config);
  const update = useConfig((s) => s.update);

  return (
    <div className="space-y-2 mt-3">
      <FormField
        label="Docker Hub Repository"
        hint="<user>/<repository>"
        initValue={dockerHubRepository ?? ""}
        onChange={(e) => update((s) => ({ ...s, dockerHubRepository: e }))}
      />
      <FilePickerField
        label="KubeConfig Path (Optional)"
        hint="kubeconfig file"
        initValue={kubeConfigPath ?? ""}
        onChange={(val) => update((s) => ({ ...s, kubeConfigPath: val }))}
        onBrowse={SelectKubeConfig}
      />
    </div>
  );
}
