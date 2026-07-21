from flask import Flask, request, jsonify, render_template
import socket

app = Flask(__name__)

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/send_udp', methods=['POST'])
def send_udp():
    data = request.json
    message = data['message']
    ip = data['ip']
    port = data['port']

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.sendto(message.encode('utf-8'), (ip, port))

    # Optionally, listen for a response (this example assumes an immediate response is received)
    try:
        sock.settimeout(2)
        response, addr = sock.recvfrom(1024)  # buffer size is 1024 bytes
        response_message = response.decode('utf-8')
    except socket.timeout:
        response_message = "No response received"

    return jsonify({"response": response_message})

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
