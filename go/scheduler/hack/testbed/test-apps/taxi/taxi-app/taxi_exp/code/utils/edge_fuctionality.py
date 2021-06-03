from time import sleep
from .rabbitmq import RabbitMQ

import os
def get_random_metrics(data_size=10, filesize=100000, file="./dataset/tripdata.csv"):
    import random
    data = []
    f = open(file, "r")

    offset = random.randrange(filesize)
    f.seek(offset)  # go to random position


    for i in range(0,data_size):
        f.readline()  # discard - bound to be partial line
        random_line = f.readline()
        data+=[random_line]

    f.close()
    return data

def propagate_to_edge(mqEdge: RabbitMQ, mqCloud: RabbitMQ, data):
    try:
        mqEdge.send_message(str(data))
    except:
        try:
            mqCloud.send_message(str(data))
        except Exception as e:
            print(e)
            print("data is lost")
            # sleep(5)
            # propagate(data=data)


def propagate(mq: RabbitMQ, data):
    try:
        mq.send_message(str(data))
    except Exception as e:
        print(e)
        print("data is lost")
        sleep(5)
        propagate(mq = mq, data=data)
