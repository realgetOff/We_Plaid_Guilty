# Backend - Server - ft_transcendence · We Plaid Guilty

---

## Description


### Gin

This app is built using `Gin`, a high perfomance HTTP web framework written in golang.
It routes incoming requests, handles certain middleware logic and allows us to serve JSON responses or upgrade connections to websockets.
We chose `Gin` for two major reasons: it's simple to implement and incredibly fast.

### Gorilla

In this project, we utilize the `Gorilla WebSocket` library to handle the real-time, bidirectional communication required for a live game.
While Gin handles our standard HTTP traffic, `Gorilla` takes over the moment a "handshake" occurs to upgrade that connection.

