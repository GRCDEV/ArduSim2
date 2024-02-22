# Communication module

## Purpose:
As the name suggest, this model will be used for the communication.
The idea is that each container can communicate with all the other containers through the use of JSON messages and a publish/subscribe system.
The module will be in charge of receiving messages from the containers and sending them to all the containers that are subscribed to a specific topic.
Many ideas are inspired by MQTT, and this module will function as a MQTT broker. So a valid question to ask is why not use [MQTT](https://mqtt.org/) directly, or some other technologies like [zeroMQ](https://zeromq.org/)? I thought about it, and decided not to use those technologies in order to allow a more detailed message filtering system at the broker. For instance, when this communication module is used to simulate communication between two UAVs, then we should be able to drop certain messages in order to simulate Wi-Fi (e.g. drop messages that are out of communication range). Hence, I implemented my own version. 
*Notices that the above text is written in the future tense because I still need test if my solution is efficient enough to be used for something so critical.*

## Structure:

This communication module exists of two parts:
1. Topic-validator: a library with functions that validate a specific topic.
2. Broker: a broker that excepts JSON messages (with a valid topic) and sends them to all subscribed nodes.

## To-do list:

- [ ] Library to valid topics
- [ ] Broker code
- [ ] Dockerfile
- [ ] Test nodes
- [ ] Implementation test  