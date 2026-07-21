import socket
import json
import time
import threading
import datetime
import statistics
import re
import sys

GATEWAY_IP = "gateway"
TELEMETRY_PORT = 3000
MESSAGES_PORT = 3001
LOGGER_PORT = 5000

logs_lock = threading.Lock()
forwarded_logs = []  # list of datetime (from gateway logs)
delivered_logs = []  # list of (datetime, target_uav_id) (from gateway logs)

# Track send timestamps to calculate message send interval
send_timestamps_lock = threading.Lock()
send_timestamps = []  # list of datetime when broadcast from UAV 1 was sent

def logger_server():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(("0.0.0.0", LOGGER_PORT))
    while True:
        data, _ = sock.recvfrom(65535)
        try:
            log_entry = json.loads(data.decode('utf-8'))

            # Only process logs from the gateway
            if log_entry.get("InstanceID") != "netsim_gateway":
                continue

            msg = log_entry.get("Message", "")
            
            if "Forwarded broadcast from 1 to" in msg:
                # Extract timestamp
                parts = msg.split("timestamp ")
                if len(parts) == 2:
                    ts_str = parts[1].strip()
                    ts = datetime.datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
                    with logs_lock:
                        forwarded_logs.append(ts)
            
            elif "Delivered message from 1 to" in msg:
                # Extract target UAV and timestamp
                match = re.search(r'to (\w+) timestamp (.*)', msg)
                if match:
                    target_uav = match.group(1)
                    ts_str = match.group(2).strip()
                    ts = datetime.datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
                    with logs_lock:
                        delivered_logs.append((ts, target_uav))
        except Exception:
            pass


def send_telemetry(sock, uav_count):
    # Send telemetry for N UAVs
    for i in range(1, uav_count + 1):
        payload = {
            "uav_id": str(i),
            "payload": {
                "lat": 39.48,
                "lon": -0.34,
                "alt": 512.3,
                "relative_alt": 15.0,
                "heading": 270.0
            }
        }
        sock.sendto(json.dumps(payload).encode('utf-8'), (GATEWAY_IP, TELEMETRY_PORT))

def send_broadcast(sock, uav_id):
    payload = {
        "uav_id": str(uav_id),
        "payload": {
            "nr_gps_online": 2,
            "position": {
                "heading": 1,
                "alt": 0.06,
                "relative_alt": -0.04,
                "lon": -0.3462649,
                "lat": 39.4825939
            },
            "type": "MAV_TYPE_QUADROTOR",
            "battery": 100,
            "version": "4.5.3",
            "time_boot_ms": 127069,
            "speed": {
                "vx": -0.01,
                "vy": 0.01,
                "vz": 0
            },
            "status": "OK",
            "flight_mode": "STABILIZE Custom mode; Stabilize; Manual input; "
        }
    }
    if uav_id == 1:
        with send_timestamps_lock:
            send_timestamps.append(datetime.datetime.now(datetime.timezone.utc))
    sock.sendto(json.dumps(payload).encode('utf-8'), (GATEWAY_IP, MESSAGES_PORT))

def calculate_send_interval():
    """Calculate average interval between consecutive sends from UAV 1."""
    with send_timestamps_lock:
        ts_copy = list(send_timestamps)
        send_timestamps.clear()
    if len(ts_copy) < 2:
        return None
    intervals = [(ts_copy[i+1] - ts_copy[i]).total_seconds() * 1000
                 for i in range(len(ts_copy) - 1)]
    return statistics.mean(intervals)

def calculate_delays():
    # Group deliveries to the closest forwarded log
    global forwarded_logs, delivered_logs
    
    with logs_lock:
        fwd_logs = sorted(forwarded_logs)
        deliv_logs = sorted(delivered_logs, key=lambda x: x[0])
        
        forwarded_logs.clear()
        delivered_logs.clear()
        
    if not fwd_logs:
        return []
        
    delays = []
    
    for fwd_ts in fwd_logs:
        # Find all deliveries within a 500ms window
        window_end = fwd_ts + datetime.timedelta(milliseconds=500)
        matched_deliveries = [d for d in deliv_logs if fwd_ts <= d[0] <= window_end]
        
        if matched_deliveries:
            max_delay = max([(d[0] - fwd_ts).total_seconds() * 1000 for d in matched_deliveries])
            delays.append(max_delay)
            
    return delays

def main():
    print("Starting logger server...")
    logger_thread = threading.Thread(target=logger_server, daemon=True)
    logger_thread.start()
    
    # Wait for gateway and netsim to fully start
    print("Waiting 10s for components to initialize...")
    time.sleep(10)
    
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    
    uav_counts = [2, 4, 8, 16, 32, 64, 128, 256]
    
    results_phase1 = {}
    results_phase2 = {}
    send_intervals = {}  # uav_count -> avg send interval in ms
    
    for uav_count in uav_counts:
        print(f"\n--- Testing with {uav_count} UAVs ---")
        
        stop_event = threading.Event()
        def telemetry_loop():
            while not stop_event.is_set():
                send_telemetry(sock, uav_count)
                time.sleep(0.5)  # 2Hz
                
        telemetry_thread = threading.Thread(target=telemetry_loop, daemon=True)
        telemetry_thread.start()
        
        time.sleep(2)
        
        # Phase 1
        print("Running Phase 1 (Isolated Sending)...")
        with logs_lock:
            forwarded_logs.clear()
            delivered_logs.clear()
        with send_timestamps_lock:
            send_timestamps.clear()
            
        phase1_send_times = []
        for _ in range(3):
            t_send = datetime.datetime.now(datetime.timezone.utc)
            phase1_send_times.append(t_send)
            send_broadcast(sock, 1)
            time.sleep(1)
            
        time.sleep(1) 
        
        delays_1 = calculate_delays()
        interval_1 = None
        if len(phase1_send_times) >= 2:
            intervals = [(phase1_send_times[i+1] - phase1_send_times[i]).total_seconds() * 1000
                         for i in range(len(phase1_send_times) - 1)]
            interval_1 = statistics.mean(intervals)

        if delays_1:
            avg_1 = statistics.mean(delays_1)
            print(f"Phase 1 Average Delay: {avg_1:.2f} ms  |  Avg send interval: {interval_1:.0f} ms" if interval_1 else f"Phase 1 Average Delay: {avg_1:.2f} ms")
            results_phase1[uav_count] = (avg_1, interval_1)
        else:
            print("Phase 1: No deliveries recorded!")
            results_phase1[uav_count] = (-1, interval_1)
            
        # Phase 2
        print("Running Phase 2 (Saturated Sending)...")
        
        def background_traffic():
            while not stop_event.is_set():
                for i in range(2, uav_count + 1):
                    send_broadcast(sock, i)
                time.sleep(1)
                
        bg_traffic_thread = threading.Thread(target=background_traffic, daemon=True)
        bg_traffic_thread.start()
        
        time.sleep(2) 
        
        with logs_lock:
            forwarded_logs.clear()
            delivered_logs.clear()
        with send_timestamps_lock:
            send_timestamps.clear()
            
        phase2_send_times = []
        for _ in range(3):
            t_send = datetime.datetime.now(datetime.timezone.utc)
            phase2_send_times.append(t_send)
            send_broadcast(sock, 1)
            time.sleep(1) 
            
        time.sleep(1)
        
        delays_2 = calculate_delays()
        interval_2 = None
        if len(phase2_send_times) >= 2:
            intervals = [(phase2_send_times[i+1] - phase2_send_times[i]).total_seconds() * 1000
                         for i in range(len(phase2_send_times) - 1)]
            interval_2 = statistics.mean(intervals)

        if delays_2:
            avg_2 = statistics.mean(delays_2)
            print(f"Phase 2 Average Delay: {avg_2:.2f} ms  |  Avg send interval: {interval_2:.0f} ms" if interval_2 else f"Phase 2 Average Delay: {avg_2:.2f} ms")
            results_phase2[uav_count] = (avg_2, interval_2)
        else:
            print("Phase 2: No deliveries recorded!")
            results_phase2[uav_count] = (-1, interval_2)
            
        stop_event.set()
        telemetry_thread.join()
        bg_traffic_thread.join()
        
        print(f"Waiting 2 seconds before next configuration...")
        time.sleep(2)

    print("\n\n=== FINAL RESULTS (flush_interval_ms = 10, solo logs del gateway) ===")
    header = (
        f"{'UAVs':<6} | "
        f"{'Phase 1 avg (ms)':<18} | "
        f"{'P1 Send Interval':<18} || "
        f"{'Phase 2 avg (ms)':<18} | "
        f"{'P2 Send Interval':<18}"
    )
    print(header)
    print("-" * len(header))

    for c in uav_counts:
        p1 = results_phase1.get(c, (-1, None))
        p2 = results_phase2.get(c, (-1, None))
        
        p1_ms = f"{p1[0]:.2f}" if p1[0] >= 0 else "N/A"
        p1_interval = f"{p1[1]:.0f} ms" if p1[1] is not None else "N/A"
        
        p2_ms = f"{p2[0]:.2f}" if p2[0] >= 0 else "N/A"
        p2_interval = f"{p2[1]:.0f} ms" if p2[1] is not None else "N/A"
        
        print(
            f"{c:<6} | "
            f"{p1_ms:<18} | "
            f"{p1_interval:<18} || "
            f"{p2_ms:<18} | "
            f"{p2_interval:<18}"
        )

    # Summary by phase
    print("\n--- Phase 1 Summary (Isolated Sending) ---")
    for c in uav_counts:
        p1 = results_phase1.get(c, (-1, None))
        p1_ms = f"{p1[0]:.2f} ms" if p1[0] >= 0 else "N/A"
        p1_interval = f"{p1[1]:.0f} ms" if p1[1] is not None else "N/A"
        print(f"  {c:>3} UAVs -> avg delay: {p1_ms:<12}  send interval: {p1_interval}")

    print("\n--- Phase 2 Summary (Saturated Sending) ---")
    for c in uav_counts:
        p2 = results_phase2.get(c, (-1, None))
        p2_ms = f"{p2[0]:.2f} ms" if p2[0] >= 0 else "N/A"
        p2_interval = f"{p2[1]:.0f} ms" if p2[1] is not None else "N/A"
        print(f"  {c:>3} UAVs -> avg delay: {p2_ms:<12}  send interval: {p2_interval}")

    # Keep container alive briefly so we can read logs if we want
    time.sleep(5)
    print("Test finished successfully.")

if __name__ == "__main__":
    main()
