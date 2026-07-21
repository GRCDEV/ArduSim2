package grc.ardusim2.command;

import grc.ardusim2.Config;
import grc.ardusim2.drone.Drone;

public abstract class Command {

    public int commandID;
    public Object payload;
    public int mavId;
    public int gcsId;
    public boolean expectsACKMessage = true;

    public Command(){
        this.mavId = Drone.getInstance().getMavID();
        this.gcsId = Config.GCS_ID;
    }

    public abstract void processACK();
    public abstract void processError();

    @Override
    public String toString(){
        return "Command: " + commandID + " payload" + payload.toString();
    }
}
