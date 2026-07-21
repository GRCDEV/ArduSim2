package grc.ardusim2.command;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.common.MavFrame;
import io.dronefleet.mavlink.common.PositionTargetTypemask;
import io.dronefleet.mavlink.common.SetPositionTargetGlobalInt;
import org.json.JSONObject;

import java.util.function.Function;

public class MoveByVector extends Command {

    private static final String VX = "vx";
    private static final String VY = "vy";
    private static final String VZ = "vz";

    /*
    * vx: X velocity in m/s (positive is North)
    * vy: Y velocity in m/s (positive is East)
    * vz: Z velocity in m/s (positive is down)
    * this command should be re-sent every second (the vehicle will stop after 3 seconds if no command is received)
     */
    public MoveByVector(float vx, float vy, float vz){
        super();
        this.commandID = SetPositionTargetGlobalInt.class.getAnnotation(MavlinkMessageInfo.class).id();
        super.expectsACKMessage = false;
        super.payload = SetPositionTargetGlobalInt.builder()
                .timeBootMs(System.currentTimeMillis())
                .targetSystem(mavId)
                .targetComponent(0)
                .coordinateFrame(MavFrame.MAV_FRAME_GLOBAL_RELATIVE_ALT_INT)
                .typeMask(
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_X_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_Y_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_Z_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AX_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AY_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_AZ_IGNORE,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_FORCE_SET,
                        PositionTargetTypemask.POSITION_TARGET_TYPEMASK_YAW_RATE_IGNORE
                )
                .vx(vx)
                .vy(vy)
                .vz(vz)
                .build();
        Config.logger.debug("Moving by vector vx {}, vy{}, vz{} ", vx,vy,vz);
    }

    public MoveByVector(JSONObject APIrequest) {
        this(APIrequest.getFloat(VX), APIrequest.getFloat(VY),APIrequest.getFloat(VZ));
    }

    @Override
    public void processACK() {
        Config.logger.warn("MoveToPosition should not reach processACK, since no response from the flight controller is expected.");
    }

    @Override
    public void processError() {
        Config.logger.warn("MoveToPosition should not reach processError, since no response from the flight controller is expected.");
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> request.has(VX) && request.has(VY) && request.has(VZ);
    }
}
