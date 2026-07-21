package grc.ardusim2.command;

import grc.ardusim2.Config;
import org.json.JSONObject;

public class CommandFactory {

    public static Command getCommand(JSONObject message) {
        String commandName = message.getString(Config.ENDPOINT);
        return switch (commandName) {
            case "SetMessageInterval" -> new SetMessageInterval(message);
            case "RequestMessage" -> new RequestMessage(message);
            case "Arm" -> new Arm();
            case "Disarm" -> new Disarm();
            case "SetFlightmode" -> new SetFlightmode(message);
            case "Takeoff" -> new Takeoff(message);
            case "Land" -> new Land();
            case "MoveToPosition" -> new MoveToPosition(message);
            case "MoveByVector" -> new MoveByVector(message);
            case "Rotate" -> new Rotate(message);
            case "RecoverControl" -> new RecoverControl(message);
            default -> null;
        };
    }
}
