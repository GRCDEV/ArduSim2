package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.common.*;
import org.json.JSONObject;

import java.util.function.Function;

public class MoveToPosition extends Command {

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();

    private static final String LAT = "latitude";
    private static final String LON = "longitude";
    private static final String ALT = "altitude";

    private float lat,lon,alt;

    public MoveToPosition(float lat, float lon, float alt) {
        super();
        this.lat = lat;
        this.lon = lon;
        this.alt = alt;
        this.commandID = SetPositionTargetGlobalInt.class.getAnnotation(MavlinkMessageInfo.class).id();
        super.expectsACKMessage = false;
        super.payload = SetPositionTargetGlobalInt.builder()
                .timeBootMs(System.currentTimeMillis())
                .targetSystem(mavId)
                .targetComponent(0)
                .coordinateFrame(MavFrame.MAV_FRAME_GLOBAL_RELATIVE_ALT_INT)
                .typeMask(
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_VX_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_VY_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_VZ_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AX_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AY_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AZ_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_FORCE_SET,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_YAW_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_YAW_RATE_IGNORE
                )
                .latInt(Math.round(10000000L * lat))
                .lonInt(Math.round(10000000L * lon))
                .alt(alt)
                .build();
        Config.logger.info("Moving to position {} {} {} [lat, lon, alt]",lat,lon,alt);
    }

    public MoveToPosition(JSONObject APIrequest) {
        this(APIrequest.getFloat(LAT), APIrequest.getFloat(LON),APIrequest.getFloat(ALT));
    }

    @Override
    public void processACK() {
        Config.logger.warn("MoveToPosition should not reach processACK, since no response from the flight controller is expected.");
        api.respond(this, "NACK");
        drone.setStatus(Drone.Status.OK);
    }

    @Override
    public void processError() {
        Config.logger.warn("MoveToPosition should not reach processError, since no response from the flight controller is expected.");
        api.respond(this, "NACK");
        drone.setStatus(Drone.Status.OK);
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> request.has(LAT) && request.has(LON) && request.has(ALT);
    }
}
