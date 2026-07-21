package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkEntryInfo;
import io.dronefleet.mavlink.common.CommandLong;
import io.dronefleet.mavlink.common.MavCmd;
import org.json.JSONObject;

import java.util.function.Function;

public class Takeoff extends Command {

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();

    private final float altitude;
    private final static String ALTITUDE = "altitude";

    public Takeoff(float altitude) {
        super();
        try {
            super.commandID = MavCmd.class.getField(MavCmd.MAV_CMD_NAV_TAKEOFF.name()).getAnnotation(MavlinkEntryInfo.class).value();
        } catch (NoSuchFieldException e) {
            super.commandID = MavCmd.MAV_CMD_NAV_TAKEOFF.ordinal();
        }
        this.altitude = altitude;
        super.payload = CommandLong.builder()
                .targetSystem(Drone.getInstance().getMavID())
                .targetComponent(0) // MavComponent.MAV_COMP_ID_ALL
                .command(MavCmd.MAV_CMD_NAV_TAKEOFF)
                .confirmation(0)
                .param7(this.altitude)
                .build();
    }

    public Takeoff(JSONObject APIrequest) {
        this(APIrequest.getInt(ALTITUDE));
    }

    @Override
    public void processACK() {
        drone.setStatus(Drone.Status.OK);
        Config.logger.info("Taking off to {} m",altitude);
        api.respond(this,"ACK");
    }

    @Override
    public void processError() {
        api.respond(this,"NACK");
        Config.logger.warn("Unable to take off");
        drone.setStatus(Drone.Status.OK);
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> {
            if (!request.has(ALTITUDE)) {
                Config.logger.warn("Requesting takeoff but altitude was not specified");
                return false;
            }
            int altitude = request.getInt(ALTITUDE);

            boolean insideLegalBounds = altitude >= 0 && altitude <= 120;
            if(!insideLegalBounds){
                Config.logger.warn("Requesting to takeoff to an altitude outside of the legal bounds");
                return false;
            }
            return true;
        };
    }
}
