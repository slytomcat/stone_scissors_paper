FROM scratch
WORKDIR /opt/game
COPY stone_scissors_paper .
EXPOSE 8080
ENV SSP_HOSTPORT=localhost:8080
ENV SSP_SERVERSALT
ENV SSP_REDISADDRS
ENV SSP_REDISPASSWORD
CMD ["./stone_scissors_paper"]
