FROM scratch

ADD scavenge .
ADD config.yaml .

CMD ["./scavenge", "serve"]

EXPOSE 4343

