FROM golang:1.23.4-bookworm

RUN apt update && apt -y upgrade 

RUN apt -y install chromium

WORKDIR /app

COPY . .


RUN go mod download
RUN go build -o main .

EXPOSE 1234

CMD [ "./main" ]


