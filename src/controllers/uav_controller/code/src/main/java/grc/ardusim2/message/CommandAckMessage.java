package grc.ardusim2.message;

import grc.ardusim2.Config;
import grc.ardusim2.command.Command;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.common.CommandAck;
import io.dronefleet.mavlink.common.MavResult;
import org.slf4j.Logger;

public class CommandAckMessage extends Message{

    private final CommandAck msg;
    private final Drone drone = Drone.getInstance();
    public CommandAckMessage(MavlinkMessage<?> inMsg){
        msg = (CommandAck ) inMsg.getPayload();
    }

    static Logger logger = Config.logger;
    @Override
    public void process() {
        logger.trace("Command ACK message: {}",msg.toString());
        Command lastCommandSend = drone.getLastCommandSend();
        if(lastCommandSend.commandID == msg.command().value()){
            if(msg.result().entry() == MavResult.MAV_RESULT_ACCEPTED){
                lastCommandSend.processACK();
            }else{
                lastCommandSend.processError();
            }
        }
    }
}
