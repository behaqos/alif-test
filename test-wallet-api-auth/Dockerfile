FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app/cmd 
RUN go build -o main . 
EXPOSE 3030
CMD ["/app/cmd/main"]