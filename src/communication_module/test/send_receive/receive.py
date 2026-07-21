import socket
import time
import random 
import sys
import json
from statistics import mean

def getData(json_bytes):
    decoded_message = json.loads(json_bytes.decode('utf-8'))
    message_time = decoded_message.get("time")
    current_time = time.time()
    time_difference = current_time - message_time
    sender = decoded_message.get("sender")
    return time_difference,sender

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


# Send data to a specific IP address and port
remote_address = (BROKER_IP, BROKER_PORT)

message = subscribe = b"""
{
  "topic": "$subscribe",
  "subscribe_to": "algo/#"
}"""


udp_socket.sendto(message, remote_address)

latencyDict = {}

waitingForFirstMessage = True
lastMessageReceived = 0
timeDiffLastMessage = 0

while timeDiffLastMessage < 5000 or waitingForFirstMessage:
    # Receive data from any IP address and port
    try:
        data, address = udp_socket.recvfrom(1024)
        time_difference,sender = getData(data)

        waitingForFirstMessage = False
        lastMessageReceived = time.time() * 1000
        if sender not in latencyDict:
            latencyDict[sender] = []

        latencyDict[sender].append(time_difference*1000)

        #print(f"Time to receive from {sender} = {time_difference} ms",flush=True)
        #print(f"Received message from {address}: {data.decode()}",flush=True)
    except:
        pass

    timeDiffLastMessage = (time.time()*1000) - lastMessageReceived
    
udp_socket.close()

averages = []
for sender, time_diffs in latencyDict.items():
    average = mean(time_diffs)
    averages.append(average)
    print(sender, ":", average, ":", len(time_diffs))

print("total: ", mean(averages), len(averages))

