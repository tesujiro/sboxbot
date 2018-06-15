FROM alpine
MAINTAINER tesujiro <tesujiro@gmail.com>
RUN echo "now building..."
RUN apk add --no-cache ca-certificates docker
RUN mkdir /volume
ADD ./sbox /
CMD ["/sbox"]
