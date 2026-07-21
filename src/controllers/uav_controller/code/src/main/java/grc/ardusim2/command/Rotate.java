package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkEntryInfo;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.common.*;
import org.json.JSONObject;

import java.util.function.Function;

public class Rotate extends Command{

    private static final String YAW = "yaw";
    private static final String SPEED = "speed";
    private static final String DIRECTION = "direction";
    private static final String RELATIVE = "relative";

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();
    float yaw,speed,direction,relative;

    public Rotate(float yaw, float speed, float direction, float relative){
        super();
        try {
            super.commandID = MavCmd.class.getField(MavCmd.MAV_CMD_CONDITION_YAW.name()).getAnnotation(MavlinkEntryInfo.class).value();
        } catch (NoSuchFieldException e) {
            super.commandID = MavCmd.MAV_CMD_CONDITION_YAW.ordinal();
        }

        this.yaw = yaw;
        this.speed = speed;
        this.direction = direction;
        this.relative = relative;
        this.payload = CommandLong.builder()
                .targetSystem(Drone.getInstance().getMavID())
                .targetComponent(0) // MavComponent.MAV_COMP_ID_ALL
                .command(MavCmd.MAV_CMD_CONDITION_YAW)
                .param1(yaw)
                .param2(speed)
                .param3(direction)
                .param4(relative)
                .build();


    }

    public Rotate(JSONObject APIrequest) {
        this(APIrequest.getFloat(YAW),10,0,0);
    }

    @Override
    public void processACK() {
        drone.setStatus(Drone.Status.OK);
        Config.logger.info("Rotating with direction {} to {} degrees, relative {}, with speed {}",direction,yaw,relative,speed);
        api.respond(this,"ACK");
    }

    @Override
    public void processError() {
        api.respond(this,"NACK");
        Config.logger.warn("Unable to rotate");
        drone.setStatus(Drone.Status.OK);
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> {
            if (!request.has(YAW) || !request.has(SPEED) || !request.has(DIRECTION) || !request.has(RELATIVE)) {
                Config.logger.warn("Trying to rotate but the parameter yaw, speed, direction or relative is missing {}", request);
                return false;
            }

            float yaw = request.getFloat(YAW);
            if(yaw < 0 || yaw > 360){
                Config.logger.warn("Trying to rotate but YAW {} was out of bounds [0,360]", yaw);
                return false;
            }

            float speed = request.getFloat(SPEED);
            if(speed < 0 || speed > 360){
                Config.logger.warn("Trying to rotate but SPEED {} was out of bounds [0,360]", speed);
                return false;
            }

            float direction = request.getFloat(DIRECTION);
            if(direction != -1 && direction != 0 && direction != 1){
                Config.logger.warn("trying to rotate but DIRECTION {} was out of bounds {-1,0,1}", direction);
                return false;
            }

            float relative = request.getFloat(RELATIVE);
            if(relative != 0 && relative !=1){
                Config.logger.warn("Trying to rotate but RELATIVE {} was out of bounds {0,1}",relative);
                return false;
            }
            return true;
        };
    }
}
