FROM scratch
WORKDIR /opt/game
COPY stone_scissors_paper .
EXPOSE 8080/tcp
ENV SSP_HOSTPORT=localhost:8080
CMD ["./stone_scissors_paper"]
