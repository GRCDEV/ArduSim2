package grc.ardusim2;

import grc.ardusim2.command.*;
import io.dronefleet.mavlink.annotations.MavlinkMessageInfo;
import io.dronefleet.mavlink.common.AutopilotVersion;
import io.dronefleet.mavlink.common.GlobalPositionInt;
import io.dronefleet.mavlink.common.SysStatus;
import org.json.JSONException;
import org.json.JSONObject;
import org.slf4j.Logger;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.Queue;
import java.util.function.Function;

public class APIThread extends Thread{

    private static APIThread INSTANCE;
    private boolean running;

    private DatagramSocket socket = null;
    private byte[] receiveData;
    private final HashMap<String,Function<JSONObject,Boolean>> endPoints = new HashMap<>();

    private final Queue<Command> commands = new LinkedList<>();
    private final HashMap<Command,DatagramPacket> pendingResponse = new HashMap<>();

    private final int sysStatusId = SysStatus.class.getAnnotation(MavlinkMessageInfo.class).id();
    private final int globalPositionIntId = GlobalPositionInt.class.getAnnotation(MavlinkMessageInfo.class).id();
    private final int autopilotVerisionId = AutopilotVersion.class.getAnnotation(MavlinkMessageInfo.class).id();

    static Logger logger = Config.logger;

    private APIThread(){
        logger.debug("Starting API thread at port {}", Config.API_PORT);
        running = true;
        try {
            socket = new DatagramSocket(Config.API_PORT);
            socket.setSoTimeout(10);
            receiveData = new byte[1024];
        }catch(Exception ignored){}

        initializeEndpoints();
    }

    public static APIThread getInstance(){
        if(INSTANCE == null){
            INSTANCE = new APIThread();
        }
        return INSTANCE;
    }

    @Override
    public void run(){
        commands.add(new SetMessageInterval(sysStatusId));
        commands.add(new SetMessageInterval(globalPositionIntId));
        commands.add(new RequestMessage(autopilotVerisionId));

        while(running){
            DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
            try {
                socket.receive(receivePacket);
                JSONObject jsonMessage = new JSONObject(new String(receivePacket.getData(), 0, receivePacket.getLength()));
                if(jsonMessage.has(Config.ENDPOINT) && isValidRequest(jsonMessage)){
                    Command c = CommandFactory.getCommand(jsonMessage);
                    commands.add(c);
                    if(c.expectsACKMessage){
                        pendingResponse.put(c, receivePacket);
                    }else{
                        sendData(receivePacket,"N/A");
                    }
                }else{
                    logger.warn("Got a new command request, but request was invalid: {}",jsonMessage);
                    sendData(receivePacket,"NACK");
                }
            } catch (IOException ignored) {
                //Timed out
            } catch (JSONException e){
                Config.logger.warn("JSON message was not valid: {}", Arrays.toString(receivePacket.getData()));
                sendData(receivePacket,"N/A");
            }
        }
    }

    public Command getNext(){
        return commands.poll();
    }

    public void respond(Command c, String sendData){
        DatagramPacket receivePacket = pendingResponse.get(c);
        logger.trace("responding on command {} with {}",c.toString(),sendData);
        if(receivePacket != null){
            sendData(receivePacket,sendData);
            pendingResponse.remove(c);
        }
    }

    public void end(){
        this.running = false;
    }

    private void initializeEndpoints(){
        endPoints.put("SetMessageInterval",SetMessageInterval.isValidMessage());
        endPoints.put("RequestMessage",RequestMessage.isValidMessage());
        endPoints.put("Arm", Arm.isValidMessage());
        endPoints.put("Disarm", Disarm.isValidMessage());
        endPoints.put("SetFlightmode", SetFlightmode.isValidMessage());
        endPoints.put("Takeoff", Takeoff.isValidMessage());
        endPoints.put("Land", Land.isValidMessage());
        endPoints.put("MoveToPosition", MoveToPosition.isValidMessage());
        endPoints.put("MoveByVector", MoveByVector.isValidMessage());
        endPoints.put("Rotate", Rotate.isValidMessage());
        endPoints.put("RecoverControl", RecoverControl.isValidMessage());
    }

    private void sendData(DatagramPacket receivePacket,String sendData) {
        InetAddress clientAddress = receivePacket.getAddress();
        int clientPort = receivePacket.getPort();
        DatagramPacket sendPacket = new DatagramPacket(sendData.getBytes(StandardCharsets.UTF_8), sendData.length(), clientAddress, clientPort);
        try {
            socket.send(sendPacket);
        }catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private boolean isValidRequest(JSONObject message){
        String command = message.getString(Config.ENDPOINT);
        if(endPoints.containsKey(command)){
            return endPoints.get(command).apply(message);
        }
        return false;
    }

}
