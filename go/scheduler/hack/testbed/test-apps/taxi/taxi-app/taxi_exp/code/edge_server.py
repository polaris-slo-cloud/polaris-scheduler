import os, sys
from utils.rabbitmq import RabbitMQ

def save(ch, method, properties, body):
    print(f'Received: {body!r}')
    myFile = open("data_edge", 'a+')
    myFile.write(str(body))
    myFile.write('\n')
    myFile.close()

if __name__ == '__main__':
    mq = RabbitMQ('edge')
    try:
        print('Waiting for messages. To exit press CTRL+C')
        mq.start_consuming(save)
    except KeyboardInterrupt:
        print('Interrupted')
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)
