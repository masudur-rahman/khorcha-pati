FROM ghcr.io/masudur-rahman/golang:1.24

ARG TARGETOS
ARG TARGETARCH=amd64

RUN apt update && apt upgrade -y

RUN apt-get install -y fonts-lohit-beng-bengali fonts-dejavu fontconfig

RUN wget https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.bookworm_${TARGETARCH}.deb
RUN dpkg -i wkhtmltox_0.12.6.1-3.bookworm_${TARGETARCH}.deb || true
RUN apt-get update
RUN apt-get install -f -y
RUN ldconfig
RUN rm wkhtmltox_0.12.6.1-3.bookworm_${TARGETARCH}.deb

WORKDIR /expense-tracker

ADD . .
#RUN go mod tidy && go mod vendor
RUN go build -o expense-tracker

#USER nobody:nobody
USER 65535:65535

EXPOSE 8080

ENTRYPOINT ["./expense-tracker"]
CMD ["serve"]
