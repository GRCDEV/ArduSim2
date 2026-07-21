package grc.ardusim2.connection;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkConnection;

import java.io.IOException;
import java.net.Socket;

import org.slf4j.Logger;

public class TCPConnection extends Connection {

    private Socket socket = null;
    static Logger logger = Config.logger;

    protected TCPConnection() throws IOException {
        logger.debug("Creating SITLconnection at {}:{}", Config.DRONE_IP, Config.DRONE_PORT);
        socket = new Socket(Config.DRONE_IP, Config.DRONE_PORT);
        socket.setTcpNoDelay(true);
        super.connection = MavlinkConnection.create(socket.getInputStream(), socket.getOutputStream());
    }

    public void close(){
        try {
            logger.debug("Closing SITL connection.");
            socket.close();
        } catch (IOException e) {
            logger.error("Cannot close SITL connection {}",e.getMessage());
            throw new RuntimeException(e);
        }
    }
}
