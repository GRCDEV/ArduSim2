# Mixers

This directory contains the Mixers (movement mixers) available for the simulation. The mixer is the central orchestrator of each UAV: it receives movement suggestions from various algorithms, evaluates them according to a specific strategy, and sends the final decision to the UAV controller.

## How to add a new Mixer

To add a new Mixer that is compatible with the GUI and ArduSim2's dynamic build system, follow these steps:

1. **Create a Folder**: Create a new folder inside `src/mixers/` with the name of your mixer (e.g., `my_custom_mixer`).

2. **Add a `schema.json`**: This file is strictly required for the Graphical User Interface (GUI) to discover and configure your mixer. It must contain at least the `service_id` and `title` keys.
   ```json
   {
       "service_id": "my_custom_mixer",
       "title": "My Custom Mixer",
       "properties": {
           // Any configuration properties your mixer needs go here.
           // For example:
           "mix_window_ms": {
               "type": "integer",
               "title": "Mix Window (ms)",
               "default": 200
           }
       }
   }
   ```

3. **Add a `Dockerfile`**: Your mixer folder must include a `Dockerfile` that defines how to build your service's image. The simulation system will detect this folder and build the image automatically when needed.

4. **Configuration Management**: You do not need to provide an initial `config.json` file. When the simulation starts, the GUI will dynamically generate a `config.json` file containing ONLY the values from the user-configured properties defined in your `schema.json`. This generated configuration file will be automatically mounted inside your mixer's container (typically at `/app/config.json` or as defined in your Dockerfile).

5. **Communication Logic**: 
   - **Connections**: Your mixer must be able to connect to the communication module (UDP broker) to subscribe to the algorithms' suggestions (`suggestions_topic`) and forward telemetry (`telemetry_topic`). It must also establish a direct link (UDP or otherwise defined) with the UAV controller (`uav_controller`).
   - **Endpoints and Configuration**: Ensure your mixer's code correctly accepts and parses the configuration variables injected by the GUI (like `broker_ip`, `uav_controller_ip`, `mix_window_ms`, `services_priority`, etc.) in order to operate within each UAV's isolated network.

By following these 5 steps, the GUI will automatically recognize your new mixer, making it available to be selected and deployed on drones in simulations.
