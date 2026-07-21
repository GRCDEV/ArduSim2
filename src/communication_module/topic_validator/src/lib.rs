//! This library exist to (i) validate topic names, and (ii) check if two topic names match.
//! The validation mostly follows the MQTT standard, which can be found [here](https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901241)


/// Function to check if two topic names match.
///
/// # Arguments
///
/// * `publish` - A string slice that holds the topic name of the publisher.
/// * `subscribe`- A string slice that holds the topic name of the subscriber.
///
/// # Return
/// An Option &lt; bool &gt;
/// * `None` - In case either the publish or subscribe topic is invalid.
/// * `Some(true)` In case both publish and subscribe topics are valid and match.
/// * `Some(false)` In case both publish and subscribe topics are valid and do not match.
///
/// # Note
/// A publisher topic name is valid as defined in [is_publisher_topic_valid].
/// A subscriber topic name is valid as defined in [is_subscriber_topic_valid].
/// A topic name is divided into levels using `/`.
/// For a pair of topic names to match they need to be:
/// * are exactly the same.
/// * use `+` to subsitude one level
/// * use `#` to subsitude all levels
/// 
/// # Examples
///
/// ```
/// use topic_validator::does_topic_match;
/// let publish = "level1/level2/level3";
/// let subscribe = "level1/+/level3";
/// let is_match = does_topic_match(publish, subscribe);
/// if is_match.is_none(){
///     println!("topic name was invalid");
/// }else{
///     println!("topic names match: {}",is_match.unwrap());    
///}
/// ```
pub fn does_topic_match(publish: &str, subscribe: &str) -> Option<bool>{
    if is_publisher_topic_valid(publish) == false || is_subscriber_topic_valid(subscribe) == false{
        return None;
    }
    return Some(does_valid_topic_match(publish,subscribe));
    
}

/// Checks if a given valid publish topic matches a subscribe topic.
/// Only use if you are certain that a topic name is valid. If you are not certain use [does_topic_match].
///
/// # Arguments
///
/// * `publish` - A string slice representing the publish topic.
/// * `subscribe` - A string slice representing the subscribe topic.
///
/// # Returns
///
/// A boolean value indicating whether the publish topic matches the subscribe topic.
///
/// # Examples
///
/// ```
/// use topic_validator::does_valid_topic_match;
///
/// let publish = "sports/tennis/player1";
/// let subscribe = "sports/tennis/player1";
/// assert!(does_valid_topic_match(publish, subscribe));
///
/// let publish = "sports/tennis/player1";
/// let subscribe = "sports/tennis/player2";
/// assert!(!does_valid_topic_match(publish, subscribe));
/// ```
pub fn does_valid_topic_match(publish: &str, subscribe: &str) -> bool{
    let perfect_match = publish.eq(subscribe);
    if perfect_match{
        return true;
    }

    let contains_wildcard = subscribe.contains("#") || subscribe.contains("+");
    if !perfect_match && !contains_wildcard{
        return false;
    }

    let mut pub_parts = publish.split("/");

    for sub_part in subscribe.split("/"){
        if sub_part.eq("#"){
            return true;
        }

        let pub_part = pub_parts.next();
        if pub_part.is_none(){
            return false;
        }

        if !sub_part.eq("+") && !sub_part.eq(pub_part.unwrap()){
            return false;
        }
    }
    return true;
}

/// Checks if a given topic is a valid publisher topic.
///
/// # Arguments
///
/// * `topic` - A string slice representing the topic to be checked.
///
/// # Returns
///
/// A boolean value indicating whether the topic is valid for publishing.
///
/// # Note
/// A publisher topic may not:
/// * contain wildcars (`+` ir `#`)
/// * contain a null character (`\0`)
/// * be longer then 65535 characters.
///
/// # Examples
///
/// ```
/// use topic_validator::is_publisher_topic_valid;
///
/// let valid_topic = "sports/tennis/player1";
/// assert!(is_publisher_topic_valid(valid_topic));
///
/// let invalid_topic = "invalid\0topic";
/// assert!(!is_publisher_topic_valid(invalid_topic));
/// ```
pub fn is_publisher_topic_valid(topic:&str) -> bool{
    if topic.is_empty(){
        return false;
    }else if topic.len() > 65535{
        return false;
    }else if topic.contains("\0"){
        return false;
    }else if topic.contains("#"){
        return false;
    }else if topic.contains("+"){
        return false;
    }
    return true;
}


/// Checks if a given topic is a valid subscriber topic.
///
/// # Arguments
///
/// * `topic` - A string slice representing the topic to be checked.
///
/// # Returns
///
/// A boolean value indicating whether the topic is valid for subscribing.
///
/// # note
/// A subscriber topic may not:
/// * contain a null character (`\0`)
/// * be longer then 65535 characters.
///
/// may contain a wildcard but:
/// * the single level wildcard `+` (slv) can only be between two `/` so `/+/`.
/// * the multi level wildcard `#` (mlv) can only exists once and be the last character. 
/// # Examples
///
/// ```
/// use topic_validator::is_subscriber_topic_valid;
///
/// let valid_topic = "sports/tennis/player1";
/// assert!(is_subscriber_topic_valid(valid_topic));
///
/// let invalid_topic = "invalid\0topic";
/// assert!(!is_subscriber_topic_valid(invalid_topic));
/// ```

pub fn is_subscriber_topic_valid(topic:&str) -> bool{
    if topic.is_empty(){
        return false;
    }else if topic.len() > 65535{
        return false;
    }else if topic.contains('\0'){
        return false;
    }


    let str_vec: Vec<char> = topic.chars().collect();
    let length = str_vec.len();

    for i in 0..length{
        let mlv_is_not_last_character = str_vec[i] == '#' && i != length-1;
        if mlv_is_not_last_character{
            return false;
        }

        if str_vec[i] == '+'{
            let no_slash_before_slv = i == 0 || str_vec[i-1] != '/';
            if no_slash_before_slv{
                return false;
            }

            let slv_is_last_character = i+1 >= length-1;
            if slv_is_last_character{
                return false;
            } 

            let no_slash_after_slv = str_vec[i+1] != '/';
            if no_slash_after_slv{
                return false;
            }
        }
    }
    return true;
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn publisher_topic_valid() {
        // Test empty topic
        assert_eq!(is_publisher_topic_valid(""), false);
        // Test topic length greater than 65535
        let long_topic = "a".repeat(65536);
        assert_eq!(is_publisher_topic_valid(&long_topic), false);
        // Test topic containing '#'
        assert_eq!(is_publisher_topic_valid("topic#"), false);
        // Test topic containing '+'
        assert_eq!(is_publisher_topic_valid("topic+"), false);
        // Test include null character
        assert_eq!(is_publisher_topic_valid("topic\0"),false);
        // 
        // Test valid topic
        assert_eq!(is_publisher_topic_valid("valid_topic"), true);
    }

    #[test]
    fn subscriber_topic_valid(){
        // Test empty topic
        assert_eq!(is_subscriber_topic_valid(""), false);
        // Test topic length greater than 65535
        let long_topic = "a".repeat(65536);
        assert_eq!(is_subscriber_topic_valid(&long_topic), false);
        // Test include null character
        assert_eq!(is_subscriber_topic_valid("topic\0"),false);
        // Test topic containing '#' on last character
        assert_eq!(is_subscriber_topic_valid("topic#"), true);
        // Test topic containg '#' on character other then last one
        assert_eq!(is_subscriber_topic_valid("to#pic"), false);
        // Test topic containing '+' between two /
        assert_eq!(is_subscriber_topic_valid("level1/+/leveln"),true);
        // Test topic containing '+' but not between two /
        assert_eq!(is_subscriber_topic_valid("level1+leveln"),false);
        assert_eq!(is_subscriber_topic_valid("level1/+leveln"),false);
        assert_eq!(is_subscriber_topic_valid("level1+/leveln"),false);
        assert_eq!(is_subscriber_topic_valid("level1/leveln+"),false);
        assert_eq!(is_subscriber_topic_valid("level1/leveln/+"),false);
        assert_eq!(is_subscriber_topic_valid("+/leveln"),false);
        // Test topics that are correct
        assert_eq!(is_subscriber_topic_valid("level1"),true);
        assert_eq!(is_subscriber_topic_valid("level1/level2"),true);
        assert_eq!(is_subscriber_topic_valid("level1/+/level3"),true);
        assert_eq!(is_subscriber_topic_valid("level1/+/+/level4"),true);
        assert_eq!(is_subscriber_topic_valid("level1/#"),true);
        assert_eq!(is_subscriber_topic_valid("level1/+/level3/#"),true);
    }

    #[test]
    fn does_match(){
        assert_eq!(does_topic_match("level1/+/level2", "level1/level2"), None);
        assert_eq!(does_topic_match("level1/level2", "level1/level2/+"), None);
        assert_eq!(does_topic_match("level1/level2", "something/different"), Some(false));
        assert_eq!(does_topic_match("level1/level2", "level1/level2"), Some(true));
        assert_eq!(does_topic_match("level1/level2/level3", "level1/+/level3"),Some(true));
        assert_eq!(does_topic_match("level1/level2/level3", "level1/level2/#"),Some(true));
        assert_eq!(does_topic_match("level1/level2/level3", "level1/#"),Some(true));
        assert_eq!(does_topic_match("level1/level2/level3/level4", "level1/+/level3/#"),Some(true));
        assert_eq!(does_topic_match("level1/level2", "level1/+/level3"),Some(false));
        assert_eq!(does_topic_match("level1/level2/level3","level1/+/something"),Some(false));
    }
}

