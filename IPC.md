# IPC Protocol

This article defines the IPC protocol between the xivpn app (java) and libxivpn.

## Definition
Server: android app  
Client: libxivpn

## IPC Socket

Server must pass the path where the unix domain socket is listening on via `IPC_PATH` environment variable. The path is usually `/data/data/io.github.exclude0122.xivpn/cache/ipcsock`.

## Packets


### Ping (Clientbound & Serverbound)

```
ping\n
```

Receiver must reply with Pong. Server should send this packet along with the fd to the TUN device.

### Pong (Clientbound & Serverbound)

```
pong\n
```

### Stop (Clientbound)

```
stop\n
```

Client must exit after receiving this message.


### Latency Test (Clientbound)

|  | Description | Example |
| --- | --- | --- |
| URL | http or https url | https://www.gstatic.com/generate_204 |
| Request ID | any random number | 879346587 |

```
test https://www.gstatic.com/generate_204 879346587\n
```

Client must perform a latency test to the given URL following all routing rules.

### Latency Test Result (Serverbound)

|  | Description | Example |
| --- | --- | --- |
| Request ID | same as request | 879346587 |
| Result | latency (include tls handshake if applicable) in ms | 200 |

```
test_result 879346587 200\n
```


### Protect (Serverbound)

```
protect\n
```

Client must send this packet followed by sending a fd. Client must not send any new protect requests until the previous protect request has been acknowledged.

### Protect Ack (Clientbound)

```
protect_ack\n
```

Server must send this packet once the fd sent from the client has been protected.
