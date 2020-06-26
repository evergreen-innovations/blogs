
# Setting up Elasticsearch, Logstash, Kibana (ELK) and Thingsboard environment using Docker

## Overview
Modern software comprises modular, scalable applications that are often designed and built around a microservice architecture. Using a microservice design pattern ensures high service uptime, enables software debugging of self-contained units, eases server troubleshooting, and makes the CI/CD processes more streamlined. 

Docker provides a great way to create self-contained (or isolated) microservice applications with different software and package needs. It maximizes server utilization, helps in rapid prototyping, and reduces setup time when migrating servers. One can quickly spin up an application on any server where Docker is running, massively reducing the time required to set up the deployment/development environment.  

In many of our projects, we deal with complex IoT systems. When the IoT infrastructure and the codebase increase, there is a need to have efficient logging and server monitoring to reduce developer debugging time, improve troubleshooting, and increase productivity. This blog introduces a solution to logging and server management using the ELK open-source software stack: Elasticsearch, Logstash, Kibana. The ELK stack has often helped us troubleshooting large systems. 

The blog also demonstrates how Docker can be used to deploy multiple microservices on the same machine. 

## Elasticsearch, Logstash, Kibana (ELK)
Elasticsearch, Logstash, and Kibana are three open-source projects that are well-suited for applications such as server/log search and analytics. Using the ELK stack enables searching gigabytes of unstructured log data to locate e.g. a sensor connection issue. 

For example, we can find the time and reason the server went down, retrieve the services that were running at the time, get the logs from the services running at that time, and locate the service which was causing the issue. All this is possible with one or two simple queries.

The first component of the stack is Elasticsearch, a document-oriented database capable of storing the large time-series data as JSON.

The second component is Logstash, which forms a data processing pipeline. Logstash can ingest data from different sources, perform transformations as specified, and send it to the frontend. Sources could be log data that can be sent to Logstash using Filebeat, server logs collected via collectd, stasd, etc.

Finally, Kibana is used to visualize the real-time data in Elasticsearch, query the database, perform actions such as filtering the data, etc. using KQL (Kibana Query Language). 

 ![Image of ELK](/docker-tb-elk/images/elk.png)

An example of a Kibana dashboard with some log data is shown below. The top right corner (highlighted in red) can be used to filter real-time log data by a date range. The search bar (highlighted in black) is used to write queries. The selected fields are shown in a structured way on the dashboard.

![Kibana GUI](/docker-tb-elk/images/kibana.png)

The log information is captured in JSON format and is ingested by Logstash

```yaml
{
   "caller":"meter.go:263",
   "device":"deviceName",
   "level":"info",
   "msg":"Attempting to make TCP connection at address 127.0.01:1503",
   "ts":"2020-05-22T22:50:41.2078511Z"
}
```

KQL can be used to filter the further filter the logs. For example, if I want to view only the errors in the given date range. The JSON  data has a `key` called `level`, where I specify whether my log record is an error, info, debug, warning. So I write a query to search for logs where `level:'error'` as shown below.

![Kibana GUI](/docker-tb-elk/images/kibana-query1.png)

Further, regular expressions can be used for more complex querying. In the example below I use regular expressions to retrieve error logs where the log message contains the word register.

![Kibana GUI](/docker-tb-elk/images/kibana-query2.png)

Because of such powerful querying tools, Kibana makes it easy to analyze the logs and troubleshoot problems quickly.

### Thingsboard
Most of the Internet of Things (IoT) devices such as sensors, GPS, meters generate highly granular data (at least per second). An example could be a fleet application which comprises of 2-3 temperature sensors, IMU sensors, GPS on each vehicle. 

Typically, all the data needs to presented in real-time and in a meaningful way to the engineers. In our above example, an engineer or operator may need to be alerted when the temperature of the vehicle is too high or in an event of an accident. Thingsboard is an open-source IoT platform that can be used to create such a real-time dashboard for device management. It can integrate the data from all the devices, process them, and take action on a set event.

Thingsboard supports lightweight IoT protocols - MQTT, CoAP, and network protocols like HTTP. 

A demo real-time dashboard of Thingsboard is shown below.

![Thingsboard GUI](/docker-tb-elk/images/thingsboard-example.png)

The values can be filtered based on date-time. Thingsboard can be also used to perform transformations on the raw data. In the above example, the boolean value is converted to the string "ON" the Device status. An unexpected value or a value lower than the specified threshold occurs an alarm can be triggered which can be used to alert (using the Thingsboard rule engine) an operator on-site.

Third-party services such as AWS IoT, Kinesis, Azure Event Hub can be integrated to send device stream data to the dashboard. Apart from basic analytics supported by Thingsboard, advanced analytics can also be performed by integrating with Kafka streams. 

Additionally, the dashboards can be published publicly. Internally, users with different levels of visibility and control can be created as needed.

## Starting up a Docker container

Download and install docker on your computer 
```shell
https://docs.docker.com/docker-for-mac/install/
```

Use git clone or download the code.
```bash
git clone https://github.com/evergreen-innovations/blogs
```

Navigate to the folder.

```bash
cd blogs/docker-tb-elk/ 
```

Everything is contained within the `docker-compose.yml` file

```bash
docker-compose up -d
```
    
This will start up all the services in the background.

The command ```docker ps``` will give us a list of all active containers.

``` bash
c72515ad8a3d       docker-tb-elk_thingsboard     "..."   ...      ...        0.0.0.0:1883->1883/tcp, 0.0.0.0:5683->5683/tcp, 0.0.0.0:9090->9090/tcp, 5683/udp   docker-tb-elk_thingsboard_1
a313903192e1       docker-tb-elk_logstash        "..."   ...      ...        0.0.0.0:5000->5000/tcp, 0.0.0.0:9600->9600/tcp, 0.0.0.0:5000->5000/udp, 5044/tcp   docker-tb-elk_logstash_1
ad769dd5bc1f       docker-tb-elk_kibana          "..."   ...      ...        0.0.0.0:5601->5601/tcp                                                             docker-tb-elk_kibana_1
4291571d28d0       docker-tb-elk_elasticsearch   "..."   ...      ...        0.0.0.0:9200->9200/tcp, 0.0.0.0:9300->9300/tcp                                     docker-tb-elk_elasticsearch_1    
```

If the docker container does not startup, you can look up the Docker container logs to troubleshoot using ```docker follow logs --follow```

Kibana starts up at http://localhost:5601. 

Thingsboard starts up at http://localhost:9090. 

Thingsboard comes with three default users with default credentials as given below.

```
Systen Administrator: sysadmin@thingsboard.org / sysadmin 
Tenant Administrator: tenant@thingsboard.org / tenant 
Customer User: customer@thingsboard.org / customer 
```
## Docker Integration

We can see how docker to create four completely isolated servers, each with different dependencies. Even though Elastisearch and Kibana are on separate servers they can communicate using the network rules created by us in the ```docker-compose.yml``` and linking the service to network.

```shell
networks:
  elk:
    driver: bridge
```
```shell
networks:
      - elk
```
Kibana and Logstash need Elasticsearch to be configured and running first. While initialization both the services check whether they can communicate with Elastisearch. In case the communication is not established the start up fails.
The YAML key ```depends_on``` is used to specify the order of start up of the services. In this case, Kibana does not try to start until Elastisearch docker container is running.

Because our applications (Elasticsearch and Thingsboard) consist of data storage. We don't want to lose data everytime the docker containers are started and stopped. Docker volumes can be used to persist data in the Docker containers. The volumes can be used to share data between multiple containers. 

```yaml
- type: volume
        source: elasticsearch
        target: /usr/share/elasticsearch/data
```
```yaml
 volumes:
      - mytb-data:/data
      - mytb-logs:/var/log/thingsboard
```

We generally periodically take backups of the docker volumes and move it the cloud. This way even if the server comes down because of an issue, the docker container can be restarted with the latest volume backup attached. The applications starts up with the last backed up state. A ```docker volume ls``` will display the volumes created on the machine.

```yaml
local               docker-tb-elk_elasticsearch
local               docker-tb-elk_mytb-data
local               docker-tb-elk_mytb-logs
```

## Conclusion 
In this blog we have deployed a four services with a single docker-compose file on your local machine. 

This blog is the first in the series of blogs of demonstrating how the software tools deployed in this blog can be used along with Golang to create a real-time IoT device data dashboard and log management. Read the next [article](https://www.evergreeninnovations.co/blog-simulating-iot-devices-using-go/) to get started.

