package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkEntryInfo;
import io.dronefleet.mavlink.common.CommandLong;
import io.dronefleet.mavlink.common.MavCmd;
import org.json.JSONObject;

import java.util.function.Function;

public class RequestMessage extends Command {

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();
    private final int messageID;

    private final static String MESSAGE_ID = "messageID";

    public RequestMessage(int messageID) {
        super();
        try {
            super.commandID = MavCmd.class.getField(MavCmd.MAV_CMD_REQUEST_MESSAGE.name()).getAnnotation(MavlinkEntryInfo.class).value();
        } catch (NoSuchFieldException e) {
            super.commandID = MavCmd.MAV_CMD_REQUEST_MESSAGE.ordinal();
        }
        this.messageID = messageID;
        super.payload = CommandLong.builder()
                .targetSystem(Drone.getInstance().getMavID())
                .targetComponent(0) // MavComponent.MAV_COMP_ID_ALL
                .command(MavCmd.MAV_CMD_REQUEST_MESSAGE)
                .confirmation(0)
                .param1(this.messageID) 	//messageId
                .build();
    }

    public RequestMessage(JSONObject APIrequest){
        this(APIrequest.getInt(MESSAGE_ID));
    }

    @Override
    public void processACK() {
        drone.setStatus(Drone.Status.OK);
        Config.logger.info("Succesfully requested message {} ", messageID);
        api.respond(this,"ACK");
    }

    @Override
    public void processError() {
        Config.logger.warn("Could not request message {} ", messageID);
        drone.setStatus(Drone.Status.OK);
        api.respond(this,"NACK");
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> {
            if (!request.has(MESSAGE_ID)) {
                Config.logger.warn("Trying to request message but Message ID was not specified");
                return false;
            }
            int messageId = request.getInt(MESSAGE_ID);
            return messageId >= 0;
        };
    }
}
