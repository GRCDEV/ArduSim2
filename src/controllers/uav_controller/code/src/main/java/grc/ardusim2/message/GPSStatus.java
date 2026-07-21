package grc.ardusim2.message;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.GpsStatus;

public class GPSStatus extends Message {

    private final GpsStatus msg;

    public GPSStatus(MavlinkMessage<?> inMsg) {
        msg = (GpsStatus) inMsg.getPayload();
    }
    @Override
    public void process() {
        Config.logger.info("Number of satellites visible: {}",msg.satellitesVisible());
    }
}
