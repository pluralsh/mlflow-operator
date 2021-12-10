FROM python:3-slim
ARG MLFLOW_VERSION=1.19.0

WORKDIR /mlruns/
RUN apt update && \ 
         apt install -y gcc libpq-dev && \
         pip install --no-cache-dir mlflow==$MLFLOW_VERSION psycopg2==2.9.2
EXPOSE 5000

ENV BACKEND_URI sqlite:////mlflow/mlflow.db
ENV ARTIFACT_ROOT /mlflow/artifacts

CMD mlflow server --backend-store-uri ${BACKEND_STORE_URI} --default-artifact-root ${ARTIFACT_ROOT} --host 0.0.0.0 --port 5000
