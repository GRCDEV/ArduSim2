package grc.ardusim2;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.networknt.schema.JsonSchema;
import com.networknt.schema.JsonSchemaFactory;
import com.networknt.schema.SpecVersion;
import com.networknt.schema.ValidationMessage;
import grc.ardusim2.connection.Connection.ConnectionType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.*;
import java.util.Set;


public class Config {

    public static int GCS_ID;
    public static int API_PORT;
    public static ConnectionType CONNECTION_TYPE;

    public static float PUBLISHING_PERIOD;//ms
    public static String PUBLISHING_IP;
    public static int PUBLISHING_PORT = 9877;

    // For SITL connection
    public static int DRONE_PORT;
    public static String DRONE_IP;
    // For Serial connection
    public static String SERIAL_PORT;
    public static int BAUDRATE;

    //common words
    public static final String ENDPOINT = "endpoint";

    public static Logger logger;

    public static void init(File configFile) throws Exception{
        ObjectMapper objectMapper = new ObjectMapper();
        JsonSchemaFactory schemaFactory = JsonSchemaFactory.getInstance(SpecVersion.VersionFlag.V4);

        InputStream jsonStream = new FileInputStream(configFile);
        InputStream schemaStream = inputStreamFromClasspath("config-schema.json");
        JsonNode json = objectMapper.readTree(jsonStream);
        JsonSchema schema = schemaFactory.getSchema(schemaStream);
        Set<ValidationMessage> validationResult = schema.validate(json);

        if(validationResult.isEmpty()){
            setValues(json);
            logger.debug("Config values are set.");
        }else{
            logger = LoggerFactory.getLogger("error");
            logger.error("Config file invalid.");
            validationResult.forEach(vm -> logger.error(vm.getMessage()));
            throw new RuntimeException("validation error");
        }
    }

    private static void setValues(JsonNode json) {
        logger = LoggerFactory.getLogger(json.get("LOGGING").textValue());
        GCS_ID = json.get("GCS_ID").asInt();
        API_PORT = json.get("API_PORT").asInt();
        CONNECTION_TYPE = ConnectionType.valueOf(json.get("CONNECTION_TYPE").textValue());
        PUBLISHING_PERIOD = json.get("PUBLISHING_PERIOD").asInt();
        PUBLISHING_IP = json.get("PUBLISHING_IP").textValue();
        PUBLISHING_PORT = json.get("PUBLISHING_PORT").asInt();

        if(CONNECTION_TYPE == ConnectionType.TCP){
            DRONE_PORT = json.get("DRONE_PORT").asInt();
            DRONE_IP = json.get("DRONE_IP").textValue();
        } else if (CONNECTION_TYPE == ConnectionType.SERIAL){
            SERIAL_PORT = json.get("SERIAL_PORT").textValue();
            BAUDRATE = json.get("BAUDRATE").asInt();
        }
        logger.info("Configuration file loaded: {} ", json.toPrettyString());
    }

    private static InputStream inputStreamFromClasspath(String path) {
        return Thread.currentThread().getContextClassLoader().getResourceAsStream(path);
    }
}
