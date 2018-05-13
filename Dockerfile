FROM alpine
MAINTAINER tesujiro <tesujiro@gmail.com>
RUN echo "now building..."
RUN apk add --no-cache ca-certificates docker openrc
ENV HASHTAG "#sboxbot"
ENV TWITTER_CONSUMER_KEY
ENV TWITTER_CONSUMER_SECRET
ENV TWITTER_ACCESS_TOKEN
ENV TWITTER_ACCESS_TOKEN_SECRET
ADD ./sbox /
VOLUME ./volume
CMD ["/sbox"]
