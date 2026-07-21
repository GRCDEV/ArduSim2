import Button from "../../components/button";
import Page from "../../components/page";
import PageTitle from "../../components/page-title";
import { useSimulationPersistence } from "../../hooks/useSimulationPersistence";
import BaseServicesCard from "./components/base-services-card";
import BatteryCard from "./components/battery-card";
import LoggingCard from "./components/logging-card";
import DefaultArduPilotCard from "./components/default-ardupilot-card";
import DefaultSpeedCard from "./components/default-speed-card";
import DeploymentCard from "./components/deployment-card";
import NetsimCard from "./components/netsim-card";
import SimNameCard from "./components/sim-name-card";
import WindCard from "./components/wind-card";

export default function ConfigPage() {
  const loadConfig = useSimulationPersistence((s) => s.loadConfig);

  return (
    <Page>
      <span className="flex flex-row items-end justify-between">
        <PageTitle
          title="General Config"
          subtitle="Define environment parameters and deployment infrastructure."
        />
        <Button
          label="Load Simulation Config"
          icon="folder_open"
          type="outlined"
          onClick={loadConfig}
        />
      </span>
      <div className="flex flex-row gap-3 mt-5 pb-5">
        <div className="flex-2/3 flex flex-col gap-3">
          <SimNameCard />
          <DefaultArduPilotCard />
          <BaseServicesCard />
          <DefaultSpeedCard />
          <BatteryCard />
        </div>
        <div className="flex-1/3 flex flex-col gap-3">
          <DeploymentCard />
          <NetsimCard />
          <WindCard />
          <LoggingCard />
        </div>
      </div>
    </Page>
  );
}
