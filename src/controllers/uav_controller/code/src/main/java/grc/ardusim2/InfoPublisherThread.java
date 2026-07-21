package grc.ardusim2;

import grc.ardusim2.drone.Drone;
import org.slf4j.Logger;

import java.io.IOException;
import java.net.*;

public class InfoPublisherThread extends Thread {

    private boolean running;
    private final DatagramSocket publisher;
    private final InetAddress address;
    private final int port;
    private final Drone drone = Drone.getInstance();

    static Logger logger = Config.logger;

    public InfoPublisherThread() {
        logger.debug("creating info publisher at {} {}",Config.PUBLISHING_IP, Config.PUBLISHING_PORT);
        this.running = true;
        try {
            this.publisher = new DatagramSocket();
            this.address = InetAddress.getByName(Config.PUBLISHING_IP);
            this.port = Config.PUBLISHING_PORT;
        } catch (SocketException | UnknownHostException e) {
            logger.error("Could not create the info publisher thread due to {}",e.getMessage());
            throw new RuntimeException(e);
        }

    }

    @Override
    public void run(){
        long prevTime = System.currentTimeMillis();
        logger.debug("Starting info publisher.");
        while(running){
            long posTime = System.currentTimeMillis();
            long remainingMs = (long)(Config.PUBLISHING_PERIOD - (posTime - prevTime));
            if (remainingMs <= 0) {
                byte[] sendData = drone.toJSON().toString().getBytes();
                DatagramPacket packet = new DatagramPacket(sendData, sendData.length, address, port);
                try {
                    this.publisher.send(packet);
                    logger.trace(address.toString() + " " + port + " " + drone.toJSON().toString());
                } catch (IOException e) {
                    logger.warn("Could not send drone data, {}",e.getMessage());
                }
                prevTime = posTime;
            } else {
                try { Thread.sleep(remainingMs); } catch (InterruptedException e) {}
            }
        }
    }

    public void end(){
        this.running = false;
        this.publisher.close();
        logger.debug("Closing info publisher.");
    }
}
