package grc.ardusim2.message;

import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.AutopilotVersion;
import org.slf4j.Logger;

public class AutopilotVersionMessage extends Message{

    private final AutopilotVersion msg;
    private final Drone drone;

    static Logger logger = Config.logger;

    public AutopilotVersionMessage(MavlinkMessage<?> inMsg) {
        msg = (AutopilotVersion) inMsg.getPayload();
        drone = Drone.getInstance();

    }
    @Override
    public void process() {
        String version = decodeFlightSwVersion(msg.flightSwVersion());
        logger.info("Running ArduSim version {}",version);
        drone.setVersion(version);
    }

    public static String decodeFlightSwVersion(long flightSwVersion) {
        long majorVersion = (flightSwVersion >> 24) & 0xFF;
        long  minorVersion = (flightSwVersion >> 16) & 0xFF;
        long  patchVersion = (flightSwVersion >> 8) & 0xFF;

        return String.format("%d.%d.%d", majorVersion, minorVersion, patchVersion);
    }
}
