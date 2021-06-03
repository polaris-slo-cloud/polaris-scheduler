from utils.edge_fuctionality import propagate
from utils.rabbitmq import RabbitMQ
import os

if os.path.exists("./data_edge"):
    myFile = open("./data_edge", 'r')

    mq = RabbitMQ('cloud')
    propagate(mq, data=myFile.readlines())
    print('Propagated data to cloud')

    myFile.close()
    os.remove("./data_edge")
else:
    print("There is no data")
