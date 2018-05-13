FROM alpine
MAINTAINER tesujiro <tesujiro@gmail.com>
RUN echo "now building..."
RUN apk add --no-cache ca-certificates docker openrc
ENV HASHTAG "#sboxbot"
ENV TWITTER_CONSUMER_KEY "Lkxlbw6dACV8ROtt2ASjBFomw"
ENV TWITTER_CONSUMER_SECRET "TWAHpsxAva1yNLx0uuLBCGhEJCB5E00snbQUonwyAKJGkULaIc"
ENV TWITTER_ACCESS_TOKEN "955044037743886337-Ts3x8yR7bi4dMpK4qWrlHiHoVq3qehd"
ENV TWITTER_ACCESS_TOKEN_SECRET "Q4gvsfT655jwHMqKUSUnLw9JMWnA4IJywAq03G3UTf8qQ"
ADD ./sbox /
VOLUME ./volume
CMD ["/sbox"]
