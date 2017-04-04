FROM scratch
EXPOSE 8000
COPY dataflowkit /
ENTRYPOINT ["/dataflowkit"]