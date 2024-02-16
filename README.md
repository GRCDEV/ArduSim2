# ArduSim2

## ArduSim2 - The new modular drone control software:

ArduSim2 is software made to control drones. It is still under development. Hence, it is not to be used yet. However, it is already published and git will be used, so that in the future it will be easy to understand what design decisions are taken when and why. The code is made to be used for **real UAVs** that run on the [Ardupilot](https://ardupilot.org/) flightcontroller firmware. In order to safely test features, and learn about drones without buying the necessary hardware, ArduSim2 also includes a **simulation environment**.
If you cannot wait to use ArduSim2, then I kindly refer you to its predecessor [ArduSim](https://github.com/GRCDEV/ArduSim).


## Install:

**ArduSim2 is not operational yet, although you can install parts it won't be very useful**

I try to make it very easy to install and use ArduSim2. Hence, scripts are available to install everything.
Follow the steps for your operating system below.

### Ubuntu:
However, you should do the following steps on your own:

1. Install Git:
    - sudo apt-get install git
2. Fork or clone ArduSim2:
    - git clone https://github.com/GRCDEV/ArduSim2.git
3. Enter the project:
    - cd ArduSim2
4. Make the installation script executable:
    - chmod +x ./install.sh
5. Run the installation script:
    - ./install.sh
6. Start ArduSim2:
	- docker compose up

## Repository structure:

In this section, I explain how this repository is structured so you can easily navigate through its numerous folders. 

The main folder (the one you are in now), contains the following files:

1. **License:** the MIT License which explains to you what you can do with this code;
2. **README.md**: the readme file explaining ArduSim2;
3. **install.sh**: the installation script for ArduSim2.
4. **docker compose file**: the docker compose file in order to run ArduSim2.

Furthermore, many other folders exist. Each folder corresponds to a specific component of ArduSim2.
ArduSim2 is a modular project that heavily depends on containers (Docker). Hence, its folder represents a container.
Each folder contains (at least) the following items:

1. **README.md**: more information about the specific container;
2. **src**: a folder containing the actual source code;
3. **Dockerfile**: a dockerfile necessary to create the docker image;

### Use of branches:

This GitHub repository uses branches for organizational purposes
It includes two branches:

1. **main**: This branch represents versions of ArduSim2 that are fully working and ready for 'production'.
2. **develop**: This branch is the branch that is used for active development. The code here, is the newest code but might still have bugs that need to be fixed. Hence, it is not ready to be put on the main branch. Only after thorough implementation testing, and tests on real UAVs, the code can to the main branch.  
3. **feature**: Since multiple people might want to develop new code at the same time (although at the moment I am alone), a new sub-branch has to be created for each new feature. This sub-branch must have a descriptive name, and will be used to track small changes on the feature under development. Once a feature has been implemented, the code should go to the develop branch in order for it to be tested.  
