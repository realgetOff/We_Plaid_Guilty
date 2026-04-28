# Backend - Server - ft_transcendence · We Plaid Guilty

---

## Description

### Gin

This app is built using `Gin`, a high perfomance HTTP web framework written in golang.
It routes incoming requests, handles certain middleware logic and allows us to serve JSON responses or upgrade connections to websockets.
We chose `Gin` for two major reasons: it's simple to implement and incredibly fast.


### Gorilla

#### What?

In this project, we utilize the `Gorilla WebSocket` library to handle the real-time, bidirectional communication required for a live multiplayer game. While Gin efficiently manages our standard HTTP traffic, `Gorilla` takes over the moment a "handshake" occurs to upgrade a standard HTTP connection into a persistent WebSocket connection.

#### Why?

While standard REST patterns (POST/GET) are effective for static data, they are fundamentally "pull-based," meaning the server cannot talk to the client unless the client asks first. This "stateless" approach introduces too much latency and overhead for real-time applications. WebSockets solve this by keeping a persistent, full-duplex connection open between the server and the client.

This continuous connection is absolutely critical for our project, allowing the server to instantly push events to clients without waiting for them to poll for updates. We rely on Gorilla to handle:

- Live Chat
	Instantly routing messages between players in global or private chat rooms.
    
- Live Notifications
	Pushing real time room invites and online/offline status updates seamlessly.

