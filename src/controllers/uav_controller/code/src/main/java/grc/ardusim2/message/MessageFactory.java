package grc.ardusim2.message;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkMessage;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.ardupilotmega.EkfStatusReport;
import io.dronefleet.mavlink.common.*;
import io.dronefleet.mavlink.minimal.Heartbeat;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Set;

public class MessageFactory {

    private static final Set<Integer> mavlinkIdIgnore = new HashSet<>(Arrays.asList(
            SystemTime.class.getAnnotation(MavlinkMessageInfo.class).id(),
            GpsRawInt.class.getAnnotation(MavlinkMessageInfo.class).id(),
            RawImu.class.getAnnotation(MavlinkMessageInfo.class).id(),
            ScaledPressure.class.getAnnotation(MavlinkMessageInfo.class).id(),
            Attitude.class.getAnnotation(MavlinkMessageInfo.class).id(),
            RcChannelsRaw.class.getAnnotation(MavlinkMessageInfo.class).id(),
            ServoOutputRaw.class.getAnnotation(MavlinkMessageInfo.class).id(),
            MissionCurrent.class.getAnnotation(MavlinkMessageInfo.class).id(),
            NavControllerOutput.class.getAnnotation(MavlinkMessageInfo.class).id(),
            RcChannels.class.getAnnotation(MavlinkMessageInfo.class).id(),
            VfrHud.class.getAnnotation(MavlinkMessageInfo.class).id(),
            ScaledImu.class.getAnnotation(MavlinkMessageInfo.class).id(),
            PowerStatus.class.getAnnotation(MavlinkMessageInfo.class).id(),
            TerrainReport.class.getAnnotation(MavlinkMessageInfo.class).id(),
            152, //Meminfo message, but I cannot find the id in the mavlink classes
            FenceStatus.class.getAnnotation(MavlinkMessageInfo.class).id(),
            163, //Ahrs message (some sensor I think), but I cannot find the id in the mavlink classes
            178, //Ahrs2 message (some sensor I think), but I cannot find the id in the mavlink classes
            EkfStatusReport.class.getAnnotation(MavlinkMessageInfo.class).id(),
            Vibration.class.getAnnotation(MavlinkMessageInfo.class).id(),
            Timesync.class.getAnnotation(MavlinkMessageInfo.class).id(),
            LocalPositionNed.class.getAnnotation(MavlinkMessageInfo.class).id(),
            ScaledImu2.class.getAnnotation(MavlinkMessageInfo.class).id()
        )
    );

    private static Set<Integer> mavlinkNotRecognized = new HashSet<>();

    public static Message identifyMessage(MavlinkMessage<?> inMsg) {
        Object payload = inMsg.getPayload();
        if (payload instanceof Heartbeat) {
            return new HeartbeatMessage(inMsg);
        } else if (payload instanceof Statustext) {
            return new StatusTextMessage(inMsg);
        } else if (payload instanceof SysStatus) {
            return new SysStatusMessage(inMsg);
        } else if (payload instanceof GlobalPositionInt) {
            return new GlobalPositionIntMessage(inMsg);
        } else if (payload instanceof CommandAck) {
            return new CommandAckMessage(inMsg);
        } else if (payload instanceof GpsStatus) {
            return new GPSStatus(inMsg);
        } else if (payload instanceof AutopilotVersion) {
            return new AutopilotVersionMessage(inMsg);
        } else if (payload instanceof ParamValue){
            return new ParamValueMessage(inMsg);
        } else if (payload instanceof GpsGlobalOrigin){
            return new GPSGlobalOriginMessage(inMsg);
        } else if (payload instanceof HomePosition){
            return new HomePositionMessage(inMsg);
        } else{
            int messageID = getMessageId(inMsg.toString());
            if (!mavlinkIdIgnore.contains(messageID) && !mavlinkNotRecognized.contains(messageID)) {
                Config.logger.warn("Received unknown message {}", inMsg);
                mavlinkNotRecognized.add(messageID);
            }
            return null;
        }
    }

    public static int getMessageId(String payloadString){
        int startIndex = payloadString.indexOf("messageId=") + "messageId=".length();
        int endIndex = payloadString.indexOf(",", startIndex);
        String messageIdStr = payloadString.substring(startIndex, endIndex).trim();
        return Integer.parseInt(messageIdStr);
    }
}
