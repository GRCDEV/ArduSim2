package grc.ardusim2.message;

import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.SysStatus;

public class SysStatusMessage extends Message{

    private final SysStatus msg;
    private final Drone drone;

    public SysStatusMessage(MavlinkMessage<?> inMsg){
        msg = (SysStatus) inMsg.getPayload();
        drone = Drone.getInstance();
    }

    @Override
    public void process() {
        drone.setBattery(msg.batteryRemaining());
    }
}
