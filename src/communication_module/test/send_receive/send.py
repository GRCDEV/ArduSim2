import socket
import time
import random 
import sys
import json

def createTimeMessage():
    current_time_nano = time.time()
    message = {"topic": topic, "sender":socket.gethostname(), "time": current_time_nano}
    json_bytes = json.dumps(message).encode('utf-8')
    return json_bytes

HOST_IP = "0.0.0.0"
HOST_PORT = 3000

topic = "algo/" + socket.gethostname() + "/timing"

try:
    BROKER_IP = sys.argv[1]
    BROKER_PORT = int(sys.argv[2])
except ValueError:
    print("Error: Second argument must be an integer.",flush=True)
    sys.exit(1)

random.seed()
# Create a UDP socket
udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

# Bind the socket to a specific address and port to receive data
local_address = (HOST_IP, HOST_PORT)

# Send data to a specific IP address and port
remote_address = (BROKER_IP, BROKER_PORT)


print("start sending messages in 60 seconds",flush=True)

time.sleep(5)
start = time.time() * 1000

counter = 0
while (time.time()*1000) - start < 10000:
    message = createTimeMessage()
    udp_socket.sendto(message, remote_address)

    time.sleep(1)
    counter +=1

# Close the socket when finished
udp_socket.close()

