FROM ubuntu:24.04
WORKDIR /root/
COPY ./bin/tdapp .
CMD ["./tdapp"]