import os
import pika

class RabbitMQ:

    __queue_name_base = 'rainbow-taxi-'

    def __init__(self, queue_name: str):
        self.__queue_name = self.__queue_name_base + queue_name
        host = self.__get_rabbit_mq_host()
        self.__connection = pika.BlockingConnection(pika.ConnectionParameters(host))
        self.__channel = self.__connection.channel()
        self.__channel.queue_declare(self.__queue_name)

    def __get_rabbit_mq_host(self):
        return os.getenv('MQ_HOST', 'localhost')

    def send_message(self, msg: str):
        self.__channel.basic_publish(exchange = '', routing_key = self.__queue_name, body = msg)

    def start_consuming(self, callback):
        self.__channel.basic_consume(self.__queue_name, on_message_callback = callback, auto_ack = True)
        self.__channel.start_consuming()

    def close(self):
        self.__connection.close()
