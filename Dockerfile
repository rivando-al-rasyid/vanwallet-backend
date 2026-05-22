# check=skip=SecretsUsedInArgOrEnv
FROM postgres:18.4-alpine3.23

ENV POSTGRES_USER=vando
ENV POSTGRES_DB=vanwalletdb
ENV POSTGRES_PASSWORD=vanwalletdb


EXPOSE 5432