import socket
import time
import random 
import sys


HOST_IP = "0.0.0.0"
HOST_PORT = 3000

try:
    BROKER_IP = sys.argv[1]
    BROKER_PORT = int(sys.argv[2])
except ValueError:
    print("Error: Second argument must be an integer.",flush=True)
    sys.exit(1)


# Create a UDP socket
udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

# Bind the socket to a specific address and port to receive data
local_address = (HOST_IP, HOST_PORT)
udp_socket.bind(local_address)
udp_socket.settimeout(2.0)


data = [
    b"""
    {
      "topic": "receive",
      "payload": "hello world"
    }""",
    b"""
    {
      "topic": "other_topic",
      "payload": "test test"
    }"""
]

# Send data to a specific IP address and port
remote_address = (BROKER_IP, BROKER_PORT)

message = subscribe = b"""
{
  "topic": "$subscribe",
  "subscribe_to": "receive"
}"""

udp_socket.sendto(message, remote_address)

while True:
    message = random.choice(data)
    udp_socket.sendto(message, remote_address)

    # Receive data from any IP address and port
    try:
        data, address = udp_socket.recvfrom(1024)
        print(f"Received message from {address}: {data.decode()}",flush=True)
    except:
        pass
    time.sleep(2)

# Close the socket when finished
udp_socket.close()

