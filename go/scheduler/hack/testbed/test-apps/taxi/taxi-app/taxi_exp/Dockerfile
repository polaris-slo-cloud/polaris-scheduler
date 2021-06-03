FROM python:3.8 AS builder

# Compile the cpu-limit tool for a fully functional Fogify
WORKDIR /src
RUN git clone https://github.com/opsengine/cpulimit/
WORKDIR /src/cpulimit
RUN make

# Modify the directory structure to fit the one exepected by the scripts.
WORKDIR /src/taxi
COPY . .
RUN mv ./dataset ./code && \
    mv ./Pipfile ./code && \
    mv ./Pipfile.lock ./code && \
    mv ./README.md ./code


FROM python:3.8

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1
ENV MQ_HOST "rabbitmq"

# Copy the cpu-limit tool for Fogify
COPY --from=builder /src/cpulimit/src/cpulimit /usr/bin

WORKDIR /code
COPY --from=builder /src/taxi/code .

RUN pip install pipenv && \
    pipenv install --deploy --system && \
    apt-get update && \
    apt-get install -y stress && \
    apt-get -y install cron rsyslog && \
    apt-get clean && \
    pip uninstall pipenv -y && \
    chmod 0644 ./cronjobs && \
#try to keep logs
    echo "cron.*				/var/log/cron.log" >> /etc/rsyslog.conf && \
    service rsyslog start && \
    chmod +x ./entrypoint.sh

CMD [ "./entrypoint.sh" ]
