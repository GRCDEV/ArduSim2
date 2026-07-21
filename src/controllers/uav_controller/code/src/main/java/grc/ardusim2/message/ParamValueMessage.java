package grc.ardusim2.message;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.ParamValue;

public class ParamValueMessage extends Message{

    private final ParamValue msg;
    public ParamValueMessage(MavlinkMessage<?> inMsg){
        msg = (ParamValue) inMsg.getPayload();
    }

    @Override
    public void process() {
        Config.logger.trace("onboard parameter {} = {}", msg.paramId(), msg.paramValue());
    }
}
