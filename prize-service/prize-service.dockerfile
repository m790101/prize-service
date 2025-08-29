# base go image
FROM alpine:latest

RUN mkdir /app

COPY prizeApp /app

CMD [ "/app/prizeApp" ]



