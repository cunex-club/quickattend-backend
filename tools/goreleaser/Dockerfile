FROM gcr.io/distroless/static-debian12:nonroot

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/server /usr/bin/

USER nonroot:nonroot
EXPOSE 8000

ENTRYPOINT ["/usr/bin/server"]
