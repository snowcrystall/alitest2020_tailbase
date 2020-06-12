FROM alpine:latest

WORKDIR /usr/local/tailtrace/

COPY ./agentd/agentd /usr/local/tailtrace/
COPY ./processd/processd /usr/local/tailtrace/ 

COPY ./start.sh  /usr/local/tailtrace/
RUN mkdir /usr/local/tailtrace/tracedata/
RUN chmod +x /usr/local/tailtrace/start.sh
RUN chmod +x /usr/local/tailtrace/agentd
RUN chmod +x /usr/local/tailtrace/processd
RUN apk add --no-cache \
        libc6-compat
ENTRYPOINT /usr/local/tailtrace/start.sh


