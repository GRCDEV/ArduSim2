package grc.ardusim2.message;

import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.GlobalPositionInt;

class GlobalPositionIntMessage extends Message{

    private final GlobalPositionInt msg;
    public GlobalPositionIntMessage(MavlinkMessage<?> inMsg){
        msg = (GlobalPositionInt) inMsg.getPayload();
    }
    @Override
    public void process() {
        int lat = msg.lat();
        int lon = msg.lon();
        if(lat == 0 || lon == 0){
            return;
        }

        Drone.getInstance().updatePosition(
                msg.timeBootMs(),
                msg.lat()* 0.0000001,
                msg.lon()* 0.0000001,
                (double) Math.round((msg.alt() * 0.001) * 100) /100,
                (double) Math.round((msg.relativeAlt() * 0.001) * 100) /100,
                msg.vx() * 0.01,
                msg.vy() * 0.01,
                msg.vz() * 0.01,
                msg.hdg() *0.01
        );
    }
}
