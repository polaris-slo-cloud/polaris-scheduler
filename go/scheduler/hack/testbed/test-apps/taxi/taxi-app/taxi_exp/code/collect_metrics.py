from threading import Thread
from time import sleep

from utils.edge_fuctionality import get_random_metrics, propagate_to_edge
from utils.rabbitmq import RabbitMQ

mqEdge = RabbitMQ('edge')
mqCloud = RabbitMQ('cloud')

def helping_function():
    data = get_random_metrics()
    for t in data:
        propagate_to_edge(mqEdge = mqEdge, mqCloud = mqCloud, data = t)

while True:
    sleep(1)
    thread = Thread(target=helping_function)
    thread.start()
