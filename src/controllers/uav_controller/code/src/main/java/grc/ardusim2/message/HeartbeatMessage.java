package grc.ardusim2.message;

import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import grc.ardusim2.drone.FlightMode;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.minimal.Heartbeat;

class HeartbeatMessage extends Message{

    private final Heartbeat msg;
    private final Drone drone;
    public HeartbeatMessage(MavlinkMessage<?> inMsg){
        msg = (Heartbeat) inMsg.getPayload();
        drone = Drone.getInstance();
    }
    @Override
    public void process() {
        int baseMode = msg.baseMode().value();
        long customMode = msg.customMode();
        drone.setFlightMode(baseMode, customMode);
        drone.setType(msg.type().entry().toString());
    }
}
