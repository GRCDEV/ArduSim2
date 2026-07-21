package grc.ardusim2.command;


import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.common.RcChannelsOverride;
import org.json.JSONObject;

import java.util.function.Function;

public class RecoverControl extends Command{


    public RecoverControl(){
        super();
        super.expectsACKMessage = false;
        this.commandID = RcChannelsOverride.class.getAnnotation(MavlinkMessageInfo.class).id();
        this.payload = new RcChannelsOverride.Builder()
                .chan1Raw(0)
                .chan2Raw(0)
                .chan3Raw(0)
                .chan4Raw(0)
                .chan5Raw(0)
                .chan6Raw(0)
                .chan7Raw(0)
                .chan8Raw(0)
                .targetSystem(mavId)
                .targetComponent(0)
                .build();
        Config.logger.info("Returning control to the remote control.");
    }

    public RecoverControl(JSONObject APIrequest) {
        this();
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
        return request -> true;
    }
}
