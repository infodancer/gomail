FROM scratch

ADD ca-certificates.crt /etc/ssl/certs/
ADD main /
CMD ["/main"]
LABEL maintainer="matthew@infodancer.org"
EXPOSE 25
EXPOSE 587
EXPOSE 465
