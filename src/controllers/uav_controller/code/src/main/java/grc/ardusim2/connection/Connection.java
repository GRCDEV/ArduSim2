package grc.ardusim2.connection;

import grc.ardusim2.Config;
import grc.ardusim2.command.Command;
import grc.ardusim2.message.Message;
import grc.ardusim2.message.MessageFactory;
import io.dronefleet.mavlink.MavlinkConnection;
import io.dronefleet.mavlink.MavlinkMessage;

import java.io.EOFException;
import java.io.IOException;

public class Connection {

    public enum ConnectionType {TCP, SERIAL}

    protected MavlinkConnection connection;

    private static Connection INSTANCE = null;

    public static Connection getInstance() throws IOException {
        if (INSTANCE == null) {
            switch(Config.CONNECTION_TYPE) {
                case TCP: {
                    INSTANCE = new TCPConnection();
                    break;
                }
                case SERIAL: {
                    INSTANCE = new SerialConnection();
                    break;
                }
            }
        }
        return INSTANCE;
    }

    public Message getNext(){
        Message message = null;
        try {
            MavlinkMessage<?> inMsg = this.connection.next();
            if (inMsg != null) {
                message = MessageFactory.identifyMessage(inMsg);
            }
        }catch(EOFException e) {
            Config.logger.error("EOFException while obtaining MAVlink message: {}", e.getMessage());
            close();
        } catch (IOException e) {
            Config.logger.error("IO exception while obtaining MAVlink message: {}", e.getMessage());
        }
        return message;
    }
    public void send(int systemId, int componentId, Object payload){
        try {
            this.connection.send1(systemId,componentId,payload);
        } catch (IOException e) {
            Config.logger.error("Error in sending MAVlink message {} \n. {} ", payload.toString(), e.getMessage());
        }
    }
    public void send(Command command){
        send(Config.GCS_ID,0,command.payload);
    }

    public void close(){}
}
