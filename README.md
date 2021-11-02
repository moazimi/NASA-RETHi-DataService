# Data Distribution Service for NASA RETHi Project

This repository contains the source code of Data Repository System of NASA-RETHi project. The whole RETHi project aims to develop the Resilient Extra-Terrestrial Habitats for future Moon/Mars expedition, which is divided into three related research thrusts:

1. **System Resilience** develop the techniques needed to establish a control-theoretic paradigm for resilience, and the computational capabilities needed to capture complex behaviors and perform trade studies to weigh different choices regarding habitat architecture and onboard decisions.
2. **Situational Awareness** develop and validate generic, robust, and scalable methods for detection and diagnosis of anticipated and unanticipated faults that incorporates an automated active learning framework with robots- and humans-in-the-loop.
3. **Robotic Maintenance** develop and demonstrate the technologies needed to realize teams of independent autonomous robots, incorporating the use of soft materials, that navigate through dynamic environments, use a variety of modular sensors and end-effectors for specific needs, and perform tasks such as collaboratively replacing damaged structural elements using deployable modular hardware.

Please visit https://www.purdue.edu/rethi for more information.

## 1. Project motivation 

~~Data distribution service is an important component in the extra-terrestrial habitat system and plays a key role in sensor monitoring, data remaining, communication, and decision making.~~

~~The primary concern of any outer space activity is ensuring safety, which usually involves tons of sensor data from several different subsystems, i.e. power system, interior environment system, and intervention agents to monitoring and controlling. How to ensure the real-time guarantee and~~

## 2. Current design

### 2.1 DDS - Data flow
<img src="./img/DDS_INTE.drawio.png">

### 2.2 DDS - Data flow in Server
<img src="./img/DDS_FLOW.drawio.png">

### 2.3 DDS - Database schema
<img src="./img/DDS_SCHEMA.drawio.png" style="zoom:33%;"  >


### 2.4 DDS - System design
<img src="./img/DDS_UML.drawio.png">

## 3. Service API
- **For Python, please reference [demo.py](./demo.py) and [api.py](./api.py).**
- **For GoLang, please reference demo.go and api.go.**
- **For other Language, please implement by following standards:**


### 3.1 Packet

Data packet is the basic form to send data and also to implement service API:

![dds_packet](./img/packet.png)


- Src: Source address
- Dst: Destination address
- Message_Type: Types of packet
  - 0x00: Packet defined by Communication network
  - 0x01: Packet defined Data Service
- Data_Type: Types of data from 0 to 255
  - 0x00: No data
  - 0x01: FDD data
  - 0x02: Sensor data
  - 0x03: Agent data
- Priority: Priority of frame
- Opt: Options from 0 to 65535
  - 0x000A: Send operation
  - 0x000B: Request operation
  - 0x000C: Publish operation
  - 0x000D: Subscribe operation
- Flag:
  - 0x0000: Last frame
  - 0x0001: More fragments
  - 0xFFFE: Warning
  - 0xFFFF: Error
- Time: Physical Unix time from 0 to 4294967295
- Row: Length of data
- Col: Width of data
- Length: Flatten length of data (Row * Col)
- Param: Depends on Opt
- SubParam: Depends on Opt
- Payload: Data in bytes



### 3.2 Send

Before use the API, please make sure:

- Understand IP and Port of server 
- Understand IP, Port and ID of client: ID should be unique from 0 to 255, ID 0 is saved for habitat db, ID 1 is saved for ground db.
- Client information must be registered in server configuration files.

To send asynchronous data, first set up headers:

- Src = ID of client
- Des = 0
- Message_Type = 1
- Data_Type = 0
- Time = Physical time of sending send command
- Priority = Priority
- [Type, Row, Col, Length] depend on the data
- Opt = 10
- Param = ID of data will be sent
- Simulink_Time = Simulink time of data generated (Primary key)
- Data = No bytes here

Then set payload as the bytes array of the data, each element of data for 8 bits.

Finally send this packet by UDP channel to server.

*⚠️ Note - Send data can be lost, and no response from server.*



### 3.3 Request

To require asynchronous data, first set up headers:

- Opt = 11
- Src = ID of client
- Dst = 0
- Param 1 = ID of data will be sent
- Param 2 * $2^{16}$ + Param 3 = The start Simulink time of required data
- Param 4 = The length of required data
- Time = Physical time of sending request command
- Type = 0
- Priority = Priority
- Col = 0
- Row = 0
- Length = 0

Then send this packet by UDP channel to server.

Next keep listening from server, a packet will be send back with following headers:

- Opt = 11
- Src = 0
- Dst = ID of client
- Param 1 = ID of data sent from server
- Param 2 * $2^{16}$ + Param 3 = The start simulink time of required data
- Param 4 = The length of required data
- Time = Physical time of data sending from sever
- Type = Type of data send back
- Priority = Priority
- Row = Length of data in payload (Should be equal to Param4)
- Col = Width of data in payload
- Length = Row * Col
- Payload = The data you requested

Finally decode payload by its shape [Row * Col]

*⚠️ Note - Both request operation and response data can be lost*



### 3.4 Publish

To publish data synchronously, set up headers for registering publish first:

- Opt = 12
- Src = ID of client
- Dst = 0
- Type = 0
- Param 1 =  ID of data being published
- Time = Physical time of sending publish command
- Priority = Priority
- Length = 0
- Payload = No data for publish request

Then send this packet by UDP channel to server.

Keep listening from server, a packet will be send back with following headers:

- Opt = 12
- Src = 0
- Dst = ID of client
- Type = 0
- Param 1 = ID of data being published
- Param 2 = Simulink time intervals of published data
- Priority = Priority
- Row = 0
- Length = 0
- Payload = No data for publish request

When receive the above packet, start continuously pushing streaming to server with following headers setting. Decide the shape[Row and Col] of data based on the estimated latency of network and data frequency:

- Opt = 12
- Src = ID of client
- Dst = 0
- Type = Type of data
- Param 1 = ID of data being published
- Param 2 * $2^{16}$ + Param 3 = The simulink time of publishing data
- Priority = Priority
- Time = Physical time of publishing data
- Payload = Data published to server
- [Type, Row, Col, Length] are depended on the data

~~Once server finds data missing or latency it will send warning or error packet back.~~



### 3.5 Subscribe

To subscribe data synchronously, set up headers for registering subscribe first:

- Opt = 13
- Src = ID of client
- Dst = 0
- Param 1 = ID of data being published
- Param 2 * $2^{16}$ + Param 3 = The start simulink time of subscribing
- Priority = Priority
- Time =  Physical time of publishing data
- Row = 0
- Length = 0
- Length = 0
- Payload = No data for subscribe request

Then keep listening from server, a stream will be continuously send back with following headers:

- Opt = 13
- Src = ID of client
- Dst = 0
- Param 1 = ID of data being subscribed
- Param 2 * $2^{16}$ + Param 3 = The start simulink time of subscribing
- Param 4 = Time intervals of published data
- Priority = Priority
- Time = Physical time of data sending from server
- Row = Length of data in payload (Should be equal to Param4)
- Col = Width of data in payload
- Length = Row * Col
- Payload = Subscribed data

~~Once client finds data missing it need to send a subscribe from the missing data again.~~

 



## 4. Current Plan

This is the current [plan](https://docs.google.com/document/d/1GJCyouMTSlMumpTqZ8Hr3953wPf2M3Aw3xg-r41WJaQ/edit#heading=h.ppyfpgqg4oc5) for the DMG group.

1. Fill in the info table and link table with Ryan.
2. Change the code of MCVT simulink and send to server.
3. Can integrate MVCT simulink and server.


<img src="./img/nasa_logo.jpg" width="50" height="50"> *This project is supported by the National Aeronautics and Space Administration*



