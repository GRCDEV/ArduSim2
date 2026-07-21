package grc.ardusim2.command;

import grc.ardusim2.APIThread;
import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;
import io.dronefleet.mavlink.annotations.MavlinkEntryInfo;
import io.dronefleet.mavlink.common.CommandLong;
import io.dronefleet.mavlink.common.MavCmd;
import org.json.JSONObject;

import java.util.function.Function;

public class SetMessageInterval extends Command{

    private final Drone drone = Drone.getInstance();
    private final APIThread api = APIThread.getInstance();
    private final int messageID;

    private final static String MESSAGE_ID = "messageID";

    public SetMessageInterval(int messageID){
        super();
        try {
            super.commandID = MavCmd.class.getField(MavCmd.MAV_CMD_SET_MESSAGE_INTERVAL.name()).getAnnotation(MavlinkEntryInfo.class).value();
        } catch (NoSuchFieldException e) {
            super.commandID = MavCmd.MAV_CMD_SET_MESSAGE_INTERVAL.ordinal();
        }
        this.messageID = messageID;
        super.payload = CommandLong.builder()
                .targetSystem(Drone.getInstance().getMavID())
                .targetComponent(0) // MavComponent.MAV_COMP_ID_ALL
                .command(MavCmd.MAV_CMD_SET_MESSAGE_INTERVAL)
                .confirmation(0)
                .param1(this.messageID) 	//messageId
                .param2(250000) // interval 4 time per second
                .param7(0) 		// respond target 0 = default
                .build();
    }

    public SetMessageInterval(JSONObject APIrequest){
        this(APIrequest.getInt(MESSAGE_ID));
    }

    @Override
    public void processACK() {
        drone.setStatus(Drone.Status.OK);
        Config.logger.info("Successfully requested stream of message {} ", messageID);
        api.respond(this,"ACK");
    }

    @Override
    public void processError() {
        switch (messageID){
            case 33 -> {
                Config.logger.error("Drone was not able to request GPS information.");
                drone.setStatus(Drone.Status.FATAL_ERROR);
                api.respond(this,"NACK");
            }
            case 25 -> {
                Config.logger.error("Unable to request GPS status");
                drone.setStatus(Drone.Status.OK);
                api.respond(this,"NACK");
            }
            default -> {
                Config.logger.error("An error occured requesting messageID {}, and no behaviour was implemented",messageID);
                api.respond(this,"NACK");
                drone.setStatus(Drone.Status.OK);
            }
        }
    }

    public static Function<JSONObject, Boolean> isValidMessage() {
        return request -> {
            if (!request.has(MESSAGE_ID)) {
                Config.logger.warn("Requesting message stream but message ID is not specified");
                return false;
            }
            int messageId = request.getInt(MESSAGE_ID);
            return messageId >= 0;
        };
    }
}
