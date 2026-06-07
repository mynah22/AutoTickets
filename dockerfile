FROM busybox AS builder
COPY dist/AutoTickets-Linux /autotickets
RUN chmod +x /autotickets

FROM gcr.io/distroless/static-debian12
COPY --from=builder /autotickets /autotickets
ENTRYPOINT ["/autotickets"]