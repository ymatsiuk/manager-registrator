FROM golang:latest as builder
RUN go get -d -v github.com/ymatsiuk/manager-registrator
WORKDIR /go/src/github.com/ymatsiuk/manager-registrator
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o manager-registrator .

FROM gruebel/upx:latest as upx
COPY --from=builder /go/src/github.com/ymatsiuk/manager-registrator/manager-registrator /manager-registrator.orig
RUN upx --best --lzma -o /manager-registrator /manager-registrator.orig

FROM scratch
COPY --from=upx /manager-registrator /manager-registrator
CMD ["/manager-registrator"]
