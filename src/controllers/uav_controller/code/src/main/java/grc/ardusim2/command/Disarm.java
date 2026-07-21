package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkEntryInfo;
import io.dronefleet.mavlink.common.CommandLong;
import io.dronefleet.mavlink.common.MavCmd;
import org.json.JSONObject;

import java.util.function.Function;

public class Disarm extends Command {

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();

    public Disarm() {
        super();
        try {
            super.commandID = MavCmd.class.getField(MavCmd.MAV_CMD_COMPONENT_ARM_DISARM.name()).getAnnotation(MavlinkEntryInfo.class).value();
        } catch (NoSuchFieldException e) {
            super.commandID = MavCmd.MAV_CMD_COMPONENT_ARM_DISARM.ordinal();
        }
        super.payload = CommandLong.builder()
                .targetSystem(drone.getMavID())
                .targetComponent(0)
                .command(MavCmd.MAV_CMD_COMPONENT_ARM_DISARM)
                .param1(0)
                .param2(0)
                .build();
    }

    @Override
    public void processACK() {
        drone.setStatus(Drone.Status.OK);
        Config.logger.info("Drone disarmed correctly.");
        api.respond(this, "ACK");
    }

    @Override
    public void processError() {
        Config.logger.warn("Unable to disarm, verify arming checks and try again.");
        api.respond(this, "NACK");
        drone.setStatus(Drone.Status.OK);
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> true;
    }
}
