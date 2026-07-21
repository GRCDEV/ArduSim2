# Communication module

## Purpose:
As the name suggest, this model will be used for the communication.
The idea is that each container can communicate with all the other containers through the use of JSON messages and a publish/subscribe system.
The module is in charge of receiving messages from the containers and sending them to all the containers that are subscribed to a specific topic.
Many ideas are inspired by MQTT, and this module will function as a MQTT broker. So a valid question to ask is why not use [MQTT](https://mqtt.org/) directly, or some other technologies like [zeroMQ](https://zeromq.org/)? I thought about it, and decided not to use those technologies in order to allow a more detailed message filtering system at the broker. For instance, when this communication module is used to simulate communication between two UAVs, then we should be able to drop certain messages in order to simulate Wi-Fi (e.g. drop messages that are out of communication range). Hence, I implemented my own version. 

## Structure:

This communication module exists of two parts:
1. Topic-validator: a library with functions that validate a specific topic.
2. Broker: a broker that excepts JSON messages (with a valid topic) and sends them to all subscribed nodes.

For both modules there is [documentation](./target/doc/broker/index.html)
Also a [unit test report](./target/llvm-cov/html/index.html) is available.
Finally, a integration test is with the use of docker and some additional code. This can be seen [here](./test), and can be run with `docker-compose up`

## To-do list:

- :white_check_mark: Library to valid topics
- :white_check_mark: Broker code
- :white_check_mark: Dockerfile
- :white_check_mark: Test nodes
- :white_check_mark: Implementation test  