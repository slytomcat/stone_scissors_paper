FROM scratch
WORKDIR /opt/game
COPY stone_scissors_paper .
CMD ["./stone_scissors_paper"]
