package grc.ardusim2.drone;

import grc.ardusim2.Config;
import io.dronefleet.mavlink.common.MavMode;
import io.dronefleet.mavlink.minimal.MavModeFlag;
import io.dronefleet.mavlink.util.EnumValue;

/** UAV flight modes available.
 * <p>Developed by: Jamie Wubben, from GRC research group in Universitat Polit&egrave;cnica de Val&egrave;ncia (Valencia, Spain).</p> */

public class FlightMode {

    private int customMode;
    private int baseMode;

    public FlightMode(int baseMode,long customMode){
        this.customMode = (int)customMode;
        this.baseMode = baseMode;
    }

    public FlightMode(int customMode){
        this.customMode = customMode;
        this.baseMode = -1;
    }

    public void setCustomMode(int customMode) {
        this.customMode = customMode;
    }

    public void setBaseMode(int baseMode) {
        this.baseMode = baseMode;
    }

    public String decipherBaseMode(int baseMode){
        boolean[] bitMask = int2bitArray(baseMode);
        StringBuilder MAVmode = new StringBuilder();

        if(bitMask[0]){
            MAVmode.append("Custom mode; ");
        }
        if(bitMask[1]){
            MAVmode.append("Test enabled; ");
        }
        if(bitMask[2]){
            MAVmode.append("Autonomous mode; ");
        }
        if(bitMask[3]){
            MAVmode.append("guided mode; ");
        }
        if(bitMask[4]){
            MAVmode.append("Stabilize; ");
        }
        if(bitMask[5]){
            MAVmode.append("Hardware in the loop; ");
        }
        if(bitMask[6]){
            MAVmode.append("Manual input; ");
        }
        if(bitMask[7]){
            MAVmode.append("Armed");
        }
        return MAVmode.toString();
    }

    private static boolean[] int2bitArray(int input) {
        boolean[] bits = new boolean[8];
        for (int i = 7; i >= 0; i--) {
            bits[i] = (input & (1 << i)) != 0;
        }
        return bits;
    }

    @Override
    public String toString() {
        FlightModes f = FlightModes.getMode(customMode);
        if(f != null){
            return f.toString() + " " + decipherBaseMode(baseMode);
        }
        return "N/A";
    }
}
