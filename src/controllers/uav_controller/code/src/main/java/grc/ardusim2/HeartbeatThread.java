package grc.ardusim2;

import grc.ardusim2.connection.Connection;
import io.dronefleet.mavlink.minimal.Heartbeat;
import io.dronefleet.mavlink.minimal.MavAutopilot;
import io.dronefleet.mavlink.minimal.MavState;
import io.dronefleet.mavlink.minimal.MavType;
import org.slf4j.Logger;

import java.io.IOException;

public class HeartbeatThread extends Thread{

    private final Connection connection;
    private final Object payload;
    private static final long HEARTBEAT_PERIOD =  950000000L;
    private boolean running;
    static Logger logger = Config.logger;

    public HeartbeatThread() throws IOException {
        logger.debug("Creating HeartbeatThread.");
        connection = Connection.getInstance();
        payload = Heartbeat.builder()
                .type(MavType.MAV_TYPE_GCS)
                .autopilot(MavAutopilot.MAV_AUTOPILOT_INVALID)
                .systemStatus(MavState.MAV_STATE_UNINIT)
                .mavlinkVersion(3)
                .build();
        running = true;
    }

    @Override
    public void run(){
        logger.debug("Starting HeartbeatThread.");
        long prevTime = System.nanoTime();
        while(running){
            long posTime = System.nanoTime();
            long remainingNs = HEARTBEAT_PERIOD - (posTime - prevTime);
            if (remainingNs <= 0) {
                connection.send(Config.GCS_ID,0,payload);
                prevTime = posTime;
            } else {
                try { Thread.sleep(remainingNs / 1000000); } catch (InterruptedException e) {}
            }
        }
    }

    public void end(){
        running = false;
    }
}
