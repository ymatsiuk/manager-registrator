FROM golang:latest as builder
RUN go get -v -d github.com/ymatsiuk/manager-registrator
WORKDIR /go/src/github.com/ymatsiuk/manager-registrator
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o manager-registrator .

FROM gruebel/upx:latest as upx
COPY --from=builder /go/src/github.com/ymatsiuk/manager-registrator /manager-registrator.upx
RUN upx --best --lzma -o /manager-registrator /manager-registrator.upx

FROM scratch 
COPY --from=upx /manager-registrator  /
CMD ["/manager-registrator"] 

