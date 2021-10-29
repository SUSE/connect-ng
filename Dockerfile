FROM registry.suse.com/bci/golang:1.16 AS builder
WORKDIR /go/src/connect-ng
COPY ./ ./
RUN echo 0.0.0~0 > ./internal/connect/version.txt
RUN go build -o out/ github.com/SUSE/connect-ng/suseconnect

FROM registry.suse.com/suse/sle15:latest
RUN zypper --non-interactive rm container-suseconnect && \
    zypper ref && \
    zypper in -y dmidecode less && \
    zypper clean --all
COPY --from=builder /go/src/connect-ng/out/suseconnect /usr/local/bin/SUSEConnect
