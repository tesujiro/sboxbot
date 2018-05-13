FROM alpine
MAINTAINER tesujiro <tesujiro@gmail.com>
RUN echo "now building..."
RUN apk add --no-cache ca-certificates docker openrc
ADD ./sbox /
#VOLUME ./volume
CMD ["/sbox"]
