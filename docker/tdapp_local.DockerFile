FROM ubuntu:24.04
WORKDIR /root/
RUN apt update && apt install ffmpeg -y
COPY ./bin/tdapp .
CMD ["./tdapp"]