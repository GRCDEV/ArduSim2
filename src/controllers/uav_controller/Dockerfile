FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y openjdk-17-jdk dos2unix

WORKDIR /app
# COPY ardupilot/arducopter4_5_3 /app/arducopter
COPY ardupilot/copter.parm /app
COPY ./uavController.jar /app
COPY ./config.json /app
COPY run.sh /app
# RUN chmod +x /app/arducopter4_5_3 
RUN dos2unix /app/run.sh && chmod +x /app/run.sh

EXPOSE 9876
EXPOSE 9877

CMD ["./run.sh"]