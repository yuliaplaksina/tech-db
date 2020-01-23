FROM golang:1.13-stretch AS builder

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -v -work

FROM ubuntu:19.04
ENV DEBIAN_FRONTEND=noninteractive
ENV PGVER 11
ENV PORT 5000
ENV POSTGRES_HOST localhost
ENV POSTGRES_PORT 5432
ENV POSTGRES_DB forum_db
ENV POSTGRES_USER forum_user
ENV POSTGRES_PASSWORD testpass
EXPOSE $PORT

EXPOSE 5432

RUN apt-get update && apt-get install -y postgresql-$PGVER

USER postgres

RUN service postgresql start &&\
    psql --command "CREATE USER forum_user WITH SUPERUSER PASSWORD 'testpass';" &&\
    createdb -O forum_user forum_db &&\
    service postgresql stop

COPY postgres.conf /etc/postgresql/12/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

COPY dum_hw_pdb.sql .
COPY --from=builder /usr/src/app/tech-db .
CMD service postgresql start && ./tech-db

