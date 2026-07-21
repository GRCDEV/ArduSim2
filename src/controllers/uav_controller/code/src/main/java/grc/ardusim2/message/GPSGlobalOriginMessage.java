package grc.ardusim2.message;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.GpsGlobalOrigin;

public class GPSGlobalOriginMessage extends Message{

    GpsGlobalOrigin msg;
    public GPSGlobalOriginMessage(MavlinkMessage<?> inMsg){
        msg = (GpsGlobalOrigin) inMsg.getPayload();
    }
    @Override
    public void process() {
        Config.logger.info("GPS global origin set at {} {} {} [lat,lon,alt]",
                (float) msg.latitude()/10000000,
                (float) msg.longitude()/10000000,
                (float) msg.altitude()/1000);
    }
}
