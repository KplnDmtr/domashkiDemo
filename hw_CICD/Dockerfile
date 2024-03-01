FROM golang:1.21 AS Build

COPY . /note_app

WORKDIR /note_app

RUN go build -o noteapp ./note_app


FROM gcr.io/distroless/base-debian12

WORKDIR /kapalinNoteApp

COPY --from=Build /note_app/noteapp bin/

EXPOSE 8080

CMD ["./bin/noteapp"]

