package grc.ardusim2.drone;

import grc.ardusim2.command.Command;
import org.json.JSONObject;

public class Drone {

    private static Drone INSTANCE;

    public enum Status{OK, PENDING_ACK, FATAL_ERROR, LANDING, FINISHED}
    private Status status;
    private Command lastCommandSend;

    // The filtered global position (e.g. fused GPS and accelerometers). The position is in GPS-frame (right-handed, Z-up).
    // Coming from the GLOBAL_POSTION_INT(33) message (https://mavlink.io/en/messages/common.html#GLOBAL_POSITION_INT)
    private long time_boot_ms; //Timestamp (time since system boot). [ms]
    private double lat; //Latitude, expressed [deg]
    private double lon; //Longitude, expressed [degE7]
    private double alt; //Altitude (MSL). Note that virtually all GPS modules provide both WGS84 and MSL. [m]
    private double relative_alt; //Altitude above ground [m]
    private double vx; //Ground X Speed (Latitude, positive north)[m/s]
    private double vy; //Ground Y Speed (Longitude, positive east)[m/s]
    private double vz; //Ground Z Speed (Altitude, positive down)[m/s]
    private double heading; //Vehicle heading (yaw angle), 0.0..359.99 degrees. If unknown, set to: UINT16_MAX [deg]

    private String type;
    private String version;
    private int battery = 100; //Battery energy remaining, -1: Battery remaining energy not sent by autopilot
    private int nrGpsOnline = 0; //Nr of gps signals received.
    private FlightMode flightMode;
    private int mavID = 1; // ID of the multicopter in the MAVLink protocol


    private Drone(){
        flightMode = new FlightMode(-1,-1);
        status = Status.OK;
    }

    public static Drone getInstance(){
        if(INSTANCE == null){
            INSTANCE = new Drone();
        }
        return INSTANCE;
    }

    public Status getStatus(){
        return status;
    }

    public void setStatus(Status s){
        if(this.status != Status.FATAL_ERROR){
            status = s;
        }
    }

    public void setType(String type) {
        this.type = type;
    }

    public void setBattery(int battery) {
        this.battery = battery;
    }

    public int getNrGpsOnline() {
        return nrGpsOnline;
    }

    public void incrementNrGpsOnline() {
        this.nrGpsOnline +=1;
    }

    public void setFlightMode(FlightMode flightMode) {
        this.flightMode = flightMode;
    }

    public void setFlightMode(int baseMode,long customMode){
        this.flightMode.setBaseMode(baseMode);
        this.flightMode.setCustomMode((int)customMode);

    }

    public int getMavID() {
        return mavID;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public double getRelAltitude(){return this.relative_alt;}

    public Command getLastCommandSend(){
        return lastCommandSend;
    }

    public void setLastCommandSend(Command command){
        lastCommandSend = command;
    }

    public void updatePosition(long time_boot_ms, double lat, double lon, double alt, double relative_alt, double vx, double vy, double vz, double heading){
        this.time_boot_ms = time_boot_ms;
        this.lat = lat;
        this.lon = lon;
        this.alt = alt;
        this.relative_alt = relative_alt;
        this.vx = vx;
        this.vy = vy;
        this.vz = vz;
        this.heading = heading;
    }

    @Override
    public String toString(){
        if(nrGpsOnline >=2){
            return "MavID: " + mavID + "\n"
                            + "Type: " + type + "\n"
                            + "Time [ms]: " + time_boot_ms + "\n"
                            + "Latitude [deg]: " + lat + "\n"
                            + "Longitude [deg]: " + lon + "\n"
                            + "Speed [m/s]: " + vx + " " + vy + " " + vz + "\n"
                            + "Altitue MSL [m]: " + alt + " " + relative_alt + "\n"
                            + "heading [deg]: " + heading + "\n"
                            + "Flight mode: " + flightMode.toString();
        }
        return "";
    }

    public JSONObject toJSON(){
        JSONObject data = new JSONObject();
        data.put("status", status);
        data.put("type", type);
        data.put("version", version);
        data.put("time_boot_ms", time_boot_ms);
        if(this.flightMode != null){
            data.put("flight_mode", flightMode.toString());
        }
        data.put("battery", battery);
        data.put("nr_gps_online", nrGpsOnline);

        JSONObject position = new JSONObject();
        position.put("lat", lat);
        position.put("lon", lon);
        position.put("alt", alt);
        position.put("relative_alt", relative_alt);
        position.put("heading", heading);
        data.put("position", position);

        JSONObject speed = new JSONObject();
        speed.put("vx", vx);
        speed.put("vy", vy);
        speed.put("vz", vz);
        data.put("speed", speed);
        return data;
    }
}
