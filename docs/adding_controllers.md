# Controllers

This directory contains the UAV Controllers available for the simulation.

## How to add a new UAV Controller

To add a new UAV Controller that is compatible with the GUI and the dynamic build system, follow these steps:

1. **Create a Folder**: Create a new folder inside `src/controllers/` with the name of your controller (e.g., `my_custom_controller`).

2. **Add a `schema.json`**: This file is required for the GUI to discover and configure your controller. It must contain at least `service_id` and `title`.
   ```json
   {
       "service_id": "my_custom_controller",
       "title": "My Custom Controller",
       "properties": {
           // Any configuration properties your controller requires
       }
   }
   ```

3. **Add a Base Dockerfile**: Your controller folder must include a `Dockerfile` (or `SITL` file) that defines the **base image** for your controller. 
   - *Important*: This Dockerfile should NOT copy the ArduCopter binary directly. It should only install dependencies, copy your controller's source code, and set up the execution environment (e.g., `run.sh`).
   - The simulation build system will automatically create a derived image that inherits from your base image and copies the user-selected ArduCopter binary into `/app/arducopter`.

4. **Add an `ardupilot/` folder with base parameters**: Create a subfolder named `ardupilot/` inside your controller folder and place a `copter.parm` file in it.

   ```
   src/controllers/my_custom_controller/
   ├── Dockerfile
   ├── schema.json
   └── ardupilot/
       └── copter.parm   ← ArduPilot base parameters for this controller
   ```

   The GUI uses this file as the **base parameter set** when generating the `.param` file for each UAV. It then appends simulation-specific overrides on top (battery capacity, speed, wind, logging bitmask, etc.). If this file is missing, the GUI will fall back to the default `copter.parm` from the built-in `uav_controller` controller.

   > [!TIP]
   > You can obtain a valid `copter.parm` by running the ArduCopter SITL once and copying the file from `ardupilot/Tools/autotest/default_params/copter.parm`, or by copying it directly from the existing `src/controllers/uav_controller/ardupilot/copter.parm` and adjusting parameters as needed.

5. **Execution Script**: Ensure your entrypoint (e.g., `run.sh`) expects the ArduCopter binary to be located at `/app/arducopter` and has execution permissions.

6. **Configuration**: If your controller requires a `config.json`, include it in your folder. The GUI will generate a specific config file for each UAV and mount it to `/app/config.json`.

By following these steps, your controller will be automatically discovered by the GUI, and users will be able to select it for their UAVs along with any ArduCopter binary.
