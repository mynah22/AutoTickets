FROM gcr.io/distroless/static-debian12
COPY dist/AutoTickets-Linux /autotickets
ENTRYPOINT ["/autotickets"]