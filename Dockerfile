FROM postgres:18.4-alpine3.23

ENV POSTGRES_USER=vando
ENV POSTGRES_DB=vanwalletdb


COPY ./database/backenvanwallet.sql /docker-entrypoint-initdb.d/

EXPOSE 5432