package grc.ardusim2.drone;

import grc.ardusim2.Config;

public enum FlightModes {
    //These flightmodes are used coming within a json message from the API
    //DO NOT DELETE THEM ALTHOUGH THEY MAY SEEM UNUSED THEY ARE NOT!

    STABILIZE(0),   //Self-levels the roll and pitch axis
    ACRO(1),        //Holds attitude, no self-level
    ALTHOLD(2),     //Holds altitude and self-levels the roll & pitch
    AUTO(3),        //Executes pre-defined mission
    GUIDED(4),      //Navigates to single points commanded by GCS
    LOITER(5),      //Holds altitude and position, uses GPS for movements
    RTL(6),         //Returns above takeoff location, may also include landing
    CIRCLE(7),      //Automatically circles a point in front of the vehicle
    LAND(9),        //Reduces altitude to ground level, attempts to go straight down
    DRIFT(11),      //Like stabilize, but coordinates yaw with roll like a plane
    SPORT(13),      //Alt-hold, but holds pitch & roll when sticks centered
    FLIP(14),       //Rises and completes an automated flip
    AUTOTUNE(15),   //Automated pitch and bank procedure to improve control loops
    POSHOLD(16),    //Like loiter, but manual roll and pitch when sticks not centered
    BRAKE(17),      //Brings copter to an immediate stop
    THROW(18),      //Holds position after a throwing takeoff
    AVOID_ADSB(19), //ADS-B based avoidance of manned aircraft
    GUIDED_NOGPS(20), //Guided mode does not require a GPS but it only accepts attitude targets
    SMART_RTL(21), //Retrace a safe path home
    FLOWHOLD(22), //Uses an optical flow sensor to hold position
    FOLLOW(23), //The vehicle will attempt to follow another vehicle at a specified offset
    ZIGZAG(24), //Zig-zag motion useful for crop spraying
    SYTEMID(25), //For advanced users and provides a means to generate mathematical models of the vehicles flight behavior for model generation
    AUTO_RTL(27); //Returns above takeoff location, may also include landing

    public final int customMode;

    FlightModes(int customMode){
        this.customMode = customMode;
    }

    public static FlightModes getMode(int custom){
        for(FlightModes p: FlightModes.values()){
            if(p.customMode == custom){
                return p;
            }
        }
        Config.logger.warn("Trying to identify flightmode with custom_mode {}, but was not yet included into the flightmodes enum so returning null",custom);
        return null;
    }
}
