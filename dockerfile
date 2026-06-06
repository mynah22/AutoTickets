FROM gcr.io/distroless/static-debian12
COPY dist/AutoTickets-linux /autotickets
ENTRYPOINT ["/autotickets"]