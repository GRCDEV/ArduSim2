//! This script runs a publish subscribe broker.
//! It accepts JSON messages on port 3400 and distribute the messages over the subscribers.

extern crate topic_validator;
use std::net::UdpSocket;
use std::collections::HashSet;
use std::collections::HashMap;
use anyhow::Result;
use topic_validator::*;
use std::fs::OpenOptions;
use std::io::Write;
use std::path::Path;

/// Main loop that will run forever, accepting messages and distributing them through the network. 
///
/// In order to subscribe one must send the following JSON message:
/// ```
///{
///  "topic": "$subscribe",
///  "subscribe_to": "topic_name"
///}
/// ```
///
/// In order to unsubscribe one must send the following JSON message:
/// ```
///{
///  "topic": "$unsubscribe",
///  "unsubscribe_from": "topic_name"     
///}
/// ```
/// Publishing a messages can be with any JSON message as long as it has the field "topic". e.g:
/// ```
///{
/// "topic": "test",
/// "payload": "hello world"    
///}
/// ```
fn main(){
    let socket = UdpSocket::bind("0.0.0.0:3400").expect("couldn't bind to address");

    let local_addr = socket.local_addr().expect("Failed to get local address");
    println!("Listening for UDP messages on {}", local_addr);

    // Initial log message to file if directory exists
    log_to_file("Communication module (broker) started.");

    let mut nodes: HashMap<std::net::SocketAddr, HashSet<String>> = HashMap::new();
    
    loop{
        let (msg, src_addr) = match receive_json_msg(&socket) {
            Ok((msg,src_addr)) => (msg,src_addr),
            Err(e) => {
                println!("Error: {:?}", e.context("Cannot receive json message"));
                continue;
            }
        };
        let topic = match get_string_from_json("topic",&msg){
            Some(topic) => topic,
            None => continue,
        };

        if topic.starts_with("$"){
            log_to_file(&format!("Command from {:?}: {}", src_addr, topic));
            process_command_msg(topic,msg,src_addr,&mut nodes);
        }
        else{
            log_to_file(&format!("Publishing on topic: {}", topic));
            publish_msg(topic,msg,&nodes,&socket);
        }
    }                       
}

/// Receives a JSON message from a UDP socket.
///
/// This function reads data from the provided `UdpSocket`, parses it as JSON, and returns
/// the parsed JSON message along with the source address from which the message was received.
///
/// If the received data cannot be parsed as JSON, or if there is an error receiving data
/// from the socket, an error is returned.
///
/// # Arguments
///
/// * `socket` - A reference to a `UdpSocket` from which the JSON message will be received.
///
/// # Returns
///
/// A Result containing a tuple with the parsed JSON message as a `serde_json::Value` and the
/// source address (`std::net::SocketAddr`) from which the message was received, if successful.
/// If an error occurs during receiving or parsing, an error is returned.
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn receive_json_msg(socket: &UdpSocket) -> Result<(serde_json::Value, std::net::SocketAddr)>{
    const MAX_BUFFER_OPERATING_SYS: usize = 65507;
    let mut buf = [0; MAX_BUFFER_OPERATING_SYS];

    let (number_of_bytes, src_addr) = socket.recv_from(&mut buf)?;
    let filled_buf = &mut buf[..number_of_bytes];

    let parsed = serde_json::from_slice(filled_buf)?;
    return Ok((parsed, src_addr));
}


/// Publishes a message to nodes subscribed to the specified topic.
///
/// This function iterates through the nodes and their subscribed topics in the given `nodes` hashmap.
/// For each node, it checks if any subscribed topic matches the provided `publish_topic`. If a match is found,
/// it sends the `msg` to that node via the provided `socket`.
///
/// # Arguments
///
/// * `publish_topic` - A String representing the topic to publish the message to.
/// * `msg` - A serde_json::Value containing the message to be published.
/// * `nodes` - A reference to a hashmap where each key is a SocketAddr representing a node's address,
///             and each value is a HashSet of Strings representing topics that node is subscribed to.
/// * `socket` - A reference to a UdpSocket used for sending messages to nodes.
fn publish_msg(publish_topic: String, msg: serde_json::Value,nodes: &HashMap<std::net::SocketAddr, HashSet<String>>,socket: &UdpSocket){
    if !is_publisher_topic_valid(publish_topic.as_str()){return;}

    for (node, subscribed_topics) in nodes{
        for subscribed_topic in subscribed_topics{
            if does_valid_topic_match(publish_topic.as_str(), subscribed_topic.as_str()){
                send_msg(&node, &msg,&socket);
                break;
            }
        }
    }
}

/// Sends a message to a specified node using a UDP socket.
///
/// # Arguments
///
/// * `node` - A reference to the target node's `SocketAddr`.
/// * `msg` - A reference to the message to be sent, represented as a `serde_json::Value`.
/// * `socket` - A reference to the `UdpSocket` through which the message will be sent.
fn send_msg(node: &std::net::SocketAddr, msg: &serde_json::Value,socket: &UdpSocket){
    socket.send_to(msg.to_string().as_bytes(),node).expect("could not send message");
}

/// Processes a command message received from a node.
///
/// This function takes a command string, a message as a `serde_json::Value`, the source address
/// of the node, and a mutable reference to a hashmap representing the nodes and their subscribed topics.
///
/// If the command is "$subscribe", it extracts the subscription topic from the message and adds it
/// to the subscription list for the source node in the provided `nodes` hashmap.
///
/// If the command is "$unsubscribe", it removes the topic specified in the message from the subscription
/// list for the source node in the `nodes` hashmap.
///
/// If the command is neither "$subscribe" nor "$unsubscribe", an error message is printed to stderr.
///
/// # Arguments
///
/// * `command` - A String representing the command received from the node.
/// * `msg` - A `serde_json::Value` containing additional data related to the command.
/// * `src_addr` - The source address of the node that sent the command message.
/// * `nodes` - A mutable reference to a hashmap where each key is a `SocketAddr` representing a node's address,
///             and each value is a `HashSet` of Strings representing topics that node is subscribed to.
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn process_command_msg(command: String, msg: serde_json::Value, src_addr: std::net::SocketAddr, nodes: &mut HashMap<std::net::SocketAddr, HashSet<String>>){
    if command == "$subscribe"{
        match get_subscription_topic(msg) {
            Some(topic) => add_topic_to_subscription_list(topic, src_addr, nodes),
            None => eprintln!("Error: Failed to get subscription topic"),
        }
    }else if command == "$unsubscribe"{
        remove_topic_from_subscription_list(msg,src_addr,nodes);
    }else{
        eprintln!("Error: got an invallid command message");
    }
}


/// Extracts the subscription topic from a message if present and valid.
///
/// This function takes a `serde_json::Value` message and attempts to extract the subscription topic
/// from the field named "subscribe_to". If the field is present and contains a valid topic, it returns
/// Some(topic); otherwise, it prints an error message to stderr and returns None.
///
/// # Arguments
///
/// * `msg` - A `serde_json::Value` containing the message from which the subscription topic will be extracted.
///
/// # Returns
///
/// An Option containing the subscription topic if it is valid and present in the message,
/// or None if the topic is not valid or the message does not contain the "subscribe_to" field.
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn get_subscription_topic(msg: serde_json::Value) -> Option<String>{
    let topic = match get_string_from_json("subscribe_to",&msg){
        Some(value) => value.to_string(),
        None => {
            eprintln!("error: no topic subscribe_to in message");
            return None;
        }
    };

    if is_subscriber_topic_valid(topic.as_str()){
        return Some(topic);
    }else{
        eprintln!("error: not a valid subscribe topic");
        return None;
    }
}

/// Adds a topic to the subscription list for a specific node in the nodes hashmap.
///
/// This function takes a topic string, the source address of the node, and a mutable reference to a hashmap
/// representing the nodes and their subscribed topics. It checks if the node already exists in the hashmap.
/// If it does, the topic is added to the existing set of topics for that node. If not, a new entry is created
/// for the node in the hashmap, with the topic as its only subscribed topic.
///
/// # Arguments
///
/// * `topic` - A String representing the topic to be added to the subscription list.
/// * `src_addr` - The source address of the node to which the topic will be added.
/// * `nodes` - A mutable reference to a hashmap where each key is a `SocketAddr` representing a node's address,
///             and each value is a `HashSet` of Strings representing topics that node is subscribed to.
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn add_topic_to_subscription_list(topic: String, src_addr: std::net::SocketAddr, nodes: &mut HashMap<std::net::SocketAddr, HashSet<String>>){
    match nodes.get_mut(&src_addr) {
        Some(existing_topics) => {
            existing_topics.insert(topic);
        }
        None => {
            let mut new_topics = HashSet::new();
            new_topics.insert(topic);
            nodes.insert(src_addr, new_topics);
        }
    }
}

/// Removes a topic from the subscription list for a specific node in the nodes hashmap.
///
/// This function takes a message, the source address of the node, and a mutable reference to a hashmap
/// representing the nodes and their subscribed topics. It attempts to extract the topic to unsubscribe from
/// the message and removes it from the set of topics for the specified node in the hashmap.
///
/// If the node does not exist in the hashmap, or if the topic is not found in the node's set of topics, it prints an error message and returns false.
///
/// # Arguments
///
/// * `msg` - A `serde_json::Value` containing the message from which the topic to unsubscribe will be extracted.
/// * `src_addr` - The source address of the node from which the topic will be removed.
/// * `nodes` - A mutable reference to a hashmap where each key is a `SocketAddr` representing a node's address,
///             and each value is a `HashSet` of Strings representing topics that node is subscribed to.
///
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn remove_topic_from_subscription_list(msg: serde_json::Value, src_addr: std::net::SocketAddr, nodes: &mut HashMap<std::net::SocketAddr, HashSet<String>>){
    let topic = match get_string_from_json("unsubscribe_from",&msg){
        Some(value) => value.to_string(),
        None => {
            println!("error: no topic subscribe_to in message");
            return;
        }
    };

    match nodes.get_mut(&src_addr) {
        Some(existing_topics) => existing_topics.remove(&topic),
        None => {
            println!("error: intenting to remove topic from node that does not exists");
            return;
        }
    };
}


/// Extracts a string value from a JSON object based on the specified key.
///
/// This function takes a reference to a key and a reference to a `serde_json::Value` object,
/// attempts to retrieve the corresponding value from the JSON object using the provided key,
/// and returns the value as an Option(String).
///
/// If the key is found in the JSON object and the corresponding value is a string, it returns
/// Some(String) containing the value. If the key does not exist or the corresponding value
/// is not a string, it prints an error message to stderr and returns None.
///
/// # Arguments
///
/// * `key` - A reference to a string representing the key to search for in the JSON object.
/// * `msg` - A reference to a `serde_json::Value` containing the JSON object from which the value will be extracted.
///
/// # Returns
///
/// An Option containing the extracted string value if the key is found and the corresponding value is a string,
/// or None if the key does not exist or the corresponding value is not a string.
/// # Examples
/// Check the included unit test for good examples on how to use this function.
fn get_string_from_json(key: &str, msg: &serde_json::Value) -> Option<String>{
    let topic_result = msg.get(key);
    if topic_result.is_none(){
        eprintln!("no such key in json message");
        return None;
    }

    let topic:String = topic_result.unwrap().to_string();
    let topic_clean = topic[1..topic.len() -1].to_string();
    return Some(topic_clean);
}



#[cfg(test)]
mod test{
    use super::*;

    mod process_command_msg{
        use super::*;
        #[test]
        fn subscribe_valid_topic() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));

            process_command_msg("$subscribe".to_string(), serde_json::json!({ "subscribe_to": "valid_topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 1);
            assert!(nodes.get(&src_addr).unwrap().contains("valid_topic"));
        }

        #[test]
        fn subscribe_no_topic() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));

            process_command_msg("$subscribe".to_string(), serde_json::json!({}), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 0);
        }

        #[test]
        fn unsubscribe() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));
            let mut existing_topics = HashSet::new();
            existing_topics.insert("existing_topic".to_string());
            nodes.insert(src_addr, existing_topics.clone());

            process_command_msg("$unsubscribe".to_string(), serde_json::json!({ "unsubscribe_from": "existing_topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 0);
        }

        #[test]
        fn invalid_command() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));

            process_command_msg("$invalid".to_string(), serde_json::json!({}), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 0);
        }
    }

    mod get_subscription_topic {
        use super::*;
        #[test]
        fn valid_topic() {
            let msg = serde_json::json!({ "subscribe_to": "valid_topic" });
            assert_eq!(get_subscription_topic(msg), Some("valid_topic".to_string()));
        }

        #[test]
        fn no_subscribe_to_field() {
            let msg = serde_json::json!({ "other_field": "value" });
            assert_eq!(get_subscription_topic(msg), None);
        }

        #[test]
        fn invalid_topic() {
            let msg = serde_json::json!({ "subscribe_to": "+invalid_topic" });
            assert_eq!(get_subscription_topic(msg), None);
        }
    }


    mod add_topic_to_subscription_list {
        use super::*;
        #[test]
        fn create_subscription_list_and_add_topic() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));

            add_topic_to_subscription_list("new_topic".to_string(), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 1);
            assert!(nodes.get(&src_addr).unwrap().contains("new_topic"));
        }

        #[test]
        fn add_topic_to_existing_subscription_list() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));
            let mut existing_topics = HashSet::new();
            existing_topics.insert("existing_topic".to_string());
            nodes.insert(src_addr, existing_topics);

            add_topic_to_subscription_list("new_topic".to_string(), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 2);
            assert!(nodes.get(&src_addr).unwrap().contains("new_topic"));
            assert!(nodes.get(&src_addr).unwrap().contains("existing_topic"));
        }
    }

    mod remove_topic_from_subscription_list {
        use super::*;
        #[test]
        fn existing_topic() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));
            let mut existing_topics = HashSet::new();
            existing_topics.insert("existing_topic".to_string());
            nodes.insert(src_addr, existing_topics.clone());

            remove_topic_from_subscription_list(serde_json::json!({ "unsubscribe_from": "existing_topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).is_some(),true);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 0);
        }

        #[test]
        fn non_existing_node() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));

            remove_topic_from_subscription_list(serde_json::json!({ "unsubscribe_from": "topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 0);
        }

        #[test]
        fn non_existing_topic() {
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));
            let mut existing_topics = HashSet::new();
            existing_topics.insert("existing_topic".to_string());
            nodes.insert(src_addr, existing_topics.clone());

            remove_topic_from_subscription_list(serde_json::json!({ "unsubscribe_from": "non_existing_topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 1);
            assert!(nodes.get(&src_addr).unwrap().contains("existing_topic"));
        }

        #[test]
        fn invallid_json(){
            let mut nodes = HashMap::new();
            let src_addr = std::net::SocketAddr::from(([127, 0, 0, 1], 8080));
            let mut existing_topics = HashSet::new();
            existing_topics.insert("existing_topic".to_string());
            nodes.insert(src_addr, existing_topics.clone());

            remove_topic_from_subscription_list(serde_json::json!({ "invallid": "non_existing_topic" }), src_addr, &mut nodes);

            assert_eq!(nodes.len(), 1);
            assert_eq!(nodes.get(&src_addr).unwrap().len(), 1);
            assert!(nodes.get(&src_addr).unwrap().contains("existing_topic"));
        }
    }

    mod get_string_from_json {
        use super::*;
        #[test]
        fn valid_input(){
            let json = serde_json::json!({
                "topic": "test"
            });
            let topic = get_string_from_json("topic", &json);
            assert_eq!(topic.is_some(), true);
            assert_eq!(topic.unwrap(),"test");
        }

        #[test]
        fn return_none_when_no_topic_exist(){
            let json = serde_json::json!({
                "something_else": "test"
            });
            let topic = get_string_from_json("topic", &json);
            assert_eq!(topic.is_none(), true);
        }
    }

    mod receive_json_msg {
        use super::*;
        use serial_test::serial;
        use std::time::Duration;

        #[test]
        #[serial]
        fn when_msg_is_valid(){
            let listener = UdpSocket::bind("127.0.0.1:3400").expect("couldn't bind to address");

            let sender = UdpSocket::bind("127.0.0.1:3401").expect("couldn't bind to address");
            let json = serde_json::json!({
                "something_else": "test"
            });
            sender.send_to(json.to_string().as_bytes(), "127.0.0.1:3400").expect("couldn't send data");

            let msg = receive_json_msg(&listener);
            assert_eq!(msg.is_ok(),true);
        }

        #[test]
        #[serial]
        fn when_msg_is_invalid(){
            let listener = UdpSocket::bind("127.0.0.1:3400").expect("couldn't bind to address");

            let sender = UdpSocket::bind("127.0.0.1:3401").expect("couldn't bind to address");
            let data = r#"
            {
                "topic": "123"
            "#.as_bytes();
            sender.send_to(data, "127.0.0.1:3400").expect("couldn't send data");

            let msg = receive_json_msg(&listener);
            assert_eq!(msg.is_err(),true);
        }

        #[test]
        #[serial]
        fn when_socket_fails(){
            let listener = UdpSocket::bind("127.0.0.1:3400").expect("couldn't bind to address");
            let second = Duration::new(0, 500);
            listener.set_read_timeout(Some(second)).expect("set_read_timeout call failed");

            let msg = receive_json_msg(&listener);
            assert_eq!(msg.is_err(),true);
        }
    }
}

fn log_to_file(message: &str) {
    if Path::new("/app/logs").is_dir() {
        if let Ok(mut file) = OpenOptions::new()
            .create(true)
            .append(true)
            .open("/app/logs/communication_module.log")
        {
            let now = std::time::SystemTime::now();
            let _ = writeln!(file, "[{:?}] {}", now, message);
        }
    }
}
