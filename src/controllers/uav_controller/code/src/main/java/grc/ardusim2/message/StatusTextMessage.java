package grc.ardusim2.message;

import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.Statustext;


public class StatusTextMessage extends Message{

    private final Statustext msg;
    private final Drone drone;
    public StatusTextMessage(MavlinkMessage<?> inMsg){
        msg = (Statustext ) inMsg.getPayload();
        drone = Drone.getInstance();
    }

    @Override
    public void process() {
        String text = msg.text().toUpperCase();
        if (text.startsWith("APM:COPTER") || text.startsWith("ARDUCOPTER")) {
            Config.logger.info(text);
        }
        if (text.contains("EKF") && text.contains("IMU") && text.contains("USING GPS")) {
            drone.incrementNrGpsOnline();
            Config.logger.info("GPS signal received, now {} online.", drone.getNrGpsOnline());
        }
    }
}
