package grc.ardusim2;

import grc.ardusim2.command.Command;
import grc.ardusim2.connection.Connection;
import grc.ardusim2.drone.Drone;
import grc.ardusim2.message.Message;
import org.slf4j.Logger;
import java.io.File;

public class Arducopter453Controller {

    static Connection droneConnection;
    static HeartbeatThread heartbeatThread;
    static Logger logger;

    public static void main(String[] args) {
        checkIfConfigFileIsIncluded(args);
        File configFile = checkIfConfigFileExist(args);

        try {
            Config.init(configFile);
            logger = Config.logger;
            logger.debug("Starting ArduCopter 4.5.3 Controller");
            droneConnection = Connection.getInstance();
            heartbeatThread = new HeartbeatThread();
        } catch (Exception e) {
            logger.error("Error in config or drone connection. Closing ArduCopter 4.5.3 Controller");
            throw new RuntimeException(e);
        }

        final Drone drone = Drone.getInstance();
        final InfoPublisherThread infoPublisherThread = new InfoPublisherThread();

        heartbeatThread.start();
        infoPublisherThread.start();

        APIThread apiThread = APIThread.getInstance();
        apiThread.start();

        logger.info("ArduSim init done, now starting");
        while (drone.getStatus() != Drone.Status.FATAL_ERROR && drone.getStatus() != Drone.Status.FINISHED){
            Message inmsg = droneConnection.getNext();
            if(inmsg != null){
                inmsg.process();
            }

            if(drone.getStatus() == Drone.Status.OK){
                Command command = apiThread.getNext();
                if(command != null){
                    droneConnection.send(command);
                    drone.setLastCommandSend(command);
                    if(command.expectsACKMessage){
                        drone.setStatus(Drone.Status.PENDING_ACK);
                    }
                }
            }

            if(drone.getStatus() == Drone.Status.LANDING){
                if(drone.getRelAltitude() < 0.1){
                    drone.setStatus(Drone.Status.FINISHED);
                }
            }
        }

        apiThread.end();
        heartbeatThread.end();
        infoPublisherThread.end();
        droneConnection.close();
        logger.info("Stopping ArduCopter 4.5.3 Controller");
    }

    private static File checkIfConfigFileExist(String[] args) {
        String filename = args[0];
        File file = new File(filename);
        if (!file.exists()) {
            System.out.println("ERROR: config file does not exist: " + file.getAbsolutePath());
            System.exit(1);
        }
        return file;
    }

    private static void checkIfConfigFileIsIncluded(String[] args) {
        if (args.length != 1) {
            System.out.println("ERROR: the path towards the config file should be included in the program arguments.");
            System.out.println("Usage: java Arducopter453Controller <filename>");
            System.exit(1); // Exit the program with a status code 1 indicating an error
        }
    }
}