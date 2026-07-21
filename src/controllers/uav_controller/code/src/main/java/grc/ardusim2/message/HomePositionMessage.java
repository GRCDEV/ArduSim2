package grc.ardusim2.message;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.HomePosition;

public class HomePositionMessage extends Message{

    HomePosition msg;
    public HomePositionMessage(MavlinkMessage<?> inMsg){
        msg = (HomePosition) inMsg.getPayload();
    }
    @Override
    public void process() {
        Config.logger.info("Home position set at {} {} {} [lat,lon,alt]",
                (float) msg.latitude()/10000000,
                (float) msg.longitude()/10000000,
                (float) msg.altitude()/1000);
    }
}
