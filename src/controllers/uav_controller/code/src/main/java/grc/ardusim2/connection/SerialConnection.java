package grc.ardusim2.connection;

import com.fazecast.jSerialComm.SerialPort;
import com.fazecast.jSerialComm.SerialPortInvalidPortException;
import grc.ardusim2.Config;
import io.dronefleet.mavlink.MavlinkConnection;
import org.slf4j.Logger;


public class SerialConnection extends Connection {

    private SerialPort serialPort = null;
    static Logger logger = Config.logger;

    protected SerialConnection() throws SerialPortInvalidPortException{
        logger.debug("Creating SERIAL connection at {} with baudrate {}", Config.SERIAL_PORT, Config.BAUDRATE);
        serialPort = SerialPort.getCommPort(Config.SERIAL_PORT);
        serialPort.setComPortParameters(Config.BAUDRATE, 8, 1, SerialPort.NO_PARITY);
        serialPort.setComPortTimeouts(SerialPort.TIMEOUT_READ_SEMI_BLOCKING,0,0);
        if(serialPort.openPort()) {
            super.connection = MavlinkConnection.create(serialPort.getInputStream(),serialPort.getOutputStream());
        }else{
            logger.error("Failed to connect using SERIAL");
            serialPort.closePort();
            serialPort = null;
        }
    }

    @Override
    public void close() {
        logger.debug("Closing SERIAL connection");
        serialPort.closePort();
    }
}
