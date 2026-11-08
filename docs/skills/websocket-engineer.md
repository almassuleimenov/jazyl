# websocket-engineer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/websocket-engineer (MIT). Рабочая копия скилла: `.claude/skills/websocket-engineer/`.

---
name: websocket-engineer
description: Use when building real-time communication systems with WebSockets or Socket.IO. Invoke for bidirectional messaging, horizontal scaling with Redis, presence tracking, room management.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.1"
  domain: api-architecture
  triggers: WebSocket, Socket.IO, real-time communication, bidirectional messaging, pub/sub, server push, live updates, chat systems, presence tracking
  role: specialist
  scope: implementation
  output-format: code
  related-skills: fastapi-expert, nestjs-expert, devops-engineer, monitoring-expert, security-reviewer
---

# WebSocket Engineer

## Core Workflow

1. **Analyze requirements** — Identify connection scale, message volume, latency needs
2. **Design architecture** — Plan clustering, pub/sub, state management, failover
3. **Implement** — Build WebSocket server with authentication, rooms, events
4. **Validate locally** — Test connection handling, auth, and room behavior before scaling (e.g., `npx wscat -c ws://localhost:3000`); confirm auth rejection on missing/invalid tokens, room join/leave events, and message delivery
5. **Scale** — Verify Redis connection and pub/sub round-trip before enabling the adapter; configure sticky sessions and confirm with test connections across multiple instances; set up load balancing
6. **Monitor** — Track connections, latency, throughput, error rates; add alerts for connection-count spikes and error-rate thresholds

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Protocol | `references/protocol.md` | WebSocket handshake, frames, ping/pong, close codes |
| Scaling | `references/scaling.md` | Horizontal scaling, Redis pub/sub, sticky sessions |
| Patterns | `references/patterns.md` | Rooms, namespaces, broadcasting, acknowledgments |
| Security | `references/security.md` | Authentication, authorization, rate limiting, CORS |
| Alternatives | `references/alternatives.md` | SSE, long polling, when to choose WebSockets |

## Code Examples

### Server Setup (Socket.IO with Auth and Room Management)

```js
import { createServer } from "http";
import { Server } from "socket.io";
import { createAdapter } from "@socket.io/redis-adapter";
import { createClient } from "redis";
import jwt from "jsonwebtoken";

const httpServer = createServer();
const io = new Server(httpServer, {
  cors: { origin: process.env.ALLOWED_ORIGIN, credentials: true },
  pingTimeout: 20000,
  pingInterval: 25000,
});

// Authentication middleware — runs before connection is established
io.use((socket, next) => {
  const token = socket.handshake.auth.token;
  if (!token) return next(new Error("Authentication required"));
  try {
    socket.data.user = jwt.verify(token, process.env.JWT_SECRET);
    next();
  } catch {
    next(new Error("Invalid token"));
  }
});

// Redis adapter for horizontal scaling
const pubClient = createClient({ url: process.env.REDIS_URL });
const subClient = pubClient.duplicate();
await Promise.all([pubClient.connect(), subClient.connect()]);
io.adapter(createAdapter(pubClient, subClient));

io.on("connection", (socket) => {
  const { userId } = socket.data.user;
  console.log(`connected: ${userId} (${socket.id})`);

  // Presence: mark user online
  pubClient.hSet("presence", userId, socket.id);

  socket.on("join-room", (roomId) => {
    socket.join(roomId);
    socket.to(roomId).emit("user-joined", { userId });
  });

  socket.on("message", ({ roomId, text }) => {
    io.to(roomId).emit("message", { userId, text, ts: Date.now() });
  });

  socket.on("disconnect", () => {
    pubClient.hDel("presence", userId);
    console.log(`disconnected: ${userId}`);
  });
});

httpServer.listen(3000);
```

### Client-Side Reconnection with Exponential Backoff

```js
import { io } from "socket.io-client";

const socket = io("wss://api.example.com", {
  auth: { token: getAuthToken() },
  reconnection: true,
  reconnectionAttempts: 10,
  reconnectionDelay: 1000,       // initial delay (ms)
  reconnectionDelayMax: 30000,   // cap at 30 s
  randomizationFactor: 0.5,      // jitter to avoid thundering herd
});

// Queue messages while disconnected
let messageQueue = [];

socket.on("connect", () => {
  console.log("connected:", socket.id);
  // Flush queued messages
  messageQueue.forEach((msg) => socket.emit("message", msg));
  messageQueue = [];
});

socket.on("disconnect", (reason) => {
  console.warn("disconnected:", reason);
  if (reason === "io server disconnect") socket.connect(); // manual reconnect
});

socket.on("connect_error", (err) => {
  console.error("connection error:", err.message);
});

function sendMessage(roomId, text) {
  const msg = { roomId, text };
  if (socket.connected) {
    socket.emit("message", msg);
  } else {
    messageQueue.push(msg); // buffer until reconnected
  }
}
```

## Constraints

### MUST DO
- Use sticky sessions for load balancing (WebSocket connections are stateful — requests must route to the same server instance)
- Implement heartbeat/ping-pong to detect dead connections (TCP keepalive alone is insufficient)
- Use rooms/namespaces for message scoping rather than filtering in application logic
- Queue messages during disconnection windows to avoid silent data loss
- Plan connection limits per instance before scaling horizontally

### MUST NOT DO
- Store large state in memory without a clustering strategy (use Redis or an external store)
- Mix WebSocket and HTTP on the same port without explicit upgrade handling
- Forget to handle connection cleanup (presence records, room membership, in-flight timers)
- Skip load testing before production — connection-count spikes behave differently from HTTP traffic spikes

## Output Templates

When implementing WebSocket features, provide:
1. Server setup (Socket.IO/ws configuration)
2. Event handlers (connection, message, disconnect)
3. Client library (connection, events, reconnection)
4. Brief explanation of scaling strategy

## Knowledge Reference

Socket.IO, ws, uWebSockets.js, Redis adapter, sticky sessions, nginx WebSocket proxy, JWT over WebSocket, rooms/namespaces, acknowledgments, binary data, compression, heartbeat, backpressure, horizontal pod autoscaling

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/api-architecture/websocket-engineer/)

---

## Reference: alternatives

# Real-Time Communication Alternatives

## Technology Comparison

| Feature | WebSocket | SSE | Long Polling | HTTP/2 Push | WebRTC |
|---------|-----------|-----|--------------|-------------|--------|
| Bidirectional | Yes | No | Yes | No | Yes |
| Real-time | Yes | Yes | Near | Yes | Yes |
| Browser Support | Excellent | Good | Universal | Good | Good |
| Proxy Issues | Some | Rare | Rare | Some | Some |
| Overhead | Low | Low | High | Medium | Medium |
| Use Case | Chat, games | Feeds, updates | Legacy | Assets | Audio/video |

## Server-Sent Events (SSE)

### When to Use SSE

- One-way server-to-client communication
- Live feeds, notifications, stock tickers
- Automatic reconnection needed
- Simpler than WebSockets
- Better firewall/proxy compatibility

### SSE Server (Node.js)

```javascript
const express = require('express');
const app = express();

app.get('/events', (req, res) => {
  // Set SSE headers
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');
  res.setHeader('Access-Control-Allow-Origin', '*');

  // Send initial connection message
  res.write('data: {"message": "Connected"}\n\n');

  // Send updates every 5 seconds
  const intervalId = setInterval(() => {
    const data = {
      timestamp: Date.now(),
      value: Math.random()
    };

    res.write(`data: ${JSON.stringify(data)}\n\n`);
  }, 5000);

  // Cleanup on client disconnect
  req.on('close', () => {
    clearInterval(intervalId);
    res.end();
  });
});

app.listen(3000);
```

### SSE Client

```javascript
const eventSource = new EventSource('http://localhost:3000/events');

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  // Automatically reconnects
};

// Named events
eventSource.addEventListener('update', (event) => {
  console.log('Update:', event.data);
});

// Close connection
eventSource.close();
```

### SSE with Express

```javascript
const express = require('express');
const app = express();

class SSEManager {
  constructor() {
    this.clients = new Set();
  }

  addClient(res) {
    this.clients.add(res);
  }

  removeClient(res) {
    this.clients.delete(res);
  }

  broadcast(event, data) {
    const message = `event: ${event}\ndata: ${JSON.stringify(data)}\n\n`;

    this.clients.forEach(client => {
      client.write(message);
    });
  }
}

const sseManager = new SSEManager();

app.get('/events', (req, res) => {
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');

  sseManager.addClient(res);

  req.on('close', () => {
    sseManager.removeClient(res);
  });
});

// Broadcast to all clients
setInterval(() => {
  sseManager.broadcast('update', {
    timestamp: Date.now(),
    activeClients: sseManager.clients.size
  });
}, 10000);

app.listen(3000);
```

## Long Polling

### When to Use Long Polling

- Legacy browser support needed
- Firewall/proxy blocks WebSockets
- Very infrequent updates
- Fallback mechanism

### Long Polling Server

```javascript
const express = require('express');
const app = express();

const pendingRequests = new Map();
const messages = [];

app.get('/poll', (req, res) => {
  const clientId = req.query.clientId;

  // If messages available, send immediately
  if (messages.length > 0) {
    res.json({ messages });
    messages.length = 0; // Clear messages
    return;
  }

  // Hold request until timeout or new message
  const timeout = setTimeout(() => {
    pendingRequests.delete(clientId);
    res.json({ messages: [] });
  }, 30000); // 30 second timeout

  pendingRequests.set(clientId, { res, timeout });

  req.on('close', () => {
    clearTimeout(timeout);
    pendingRequests.delete(clientId);
  });
});

app.post('/send', express.json(), (req, res) => {
  messages.push(req.body.message);

  // Respond to all pending requests
  pendingRequests.forEach(({ res, timeout }, clientId) => {
    clearTimeout(timeout);
    res.json({ messages });
    pendingRequests.delete(clientId);
  });

  messages.length = 0; // Clear messages
  res.json({ success: true });
});

app.listen(3000);
```

### Long Polling Client

```javascript
const clientId = Math.random().toString(36);

async function poll() {
  try {
    const response = await fetch(
      `http://localhost:3000/poll?clientId=${clientId}`,
      { signal: AbortSignal.timeout(35000) }
    );

    const data = await response.json();

    if (data.messages.length > 0) {
      console.log('Received messages:', data.messages);
    }

    // Immediately poll again
    poll();
  } catch (error) {
    console.error('Polling error:', error);
    // Retry after delay
    setTimeout(poll, 5000);
  }
}

poll();
```

## HTTP/2 Server Push (Deprecated)

Note: HTTP/2 Server Push is deprecated and removed from Chrome. Use 103 Early Hints instead.

```javascript
// Example for historical context only
const http2 = require('http2');
const fs = require('fs');

const server = http2.createSecureServer({
  key: fs.readFileSync('server.key'),
  cert: fs.readFileSync('server.crt')
});

server.on('stream', (stream, headers) => {
  if (headers[':path'] === '/') {
    // Push assets before HTML response
    stream.pushStream({ ':path': '/style.css' }, (err, pushStream) => {
      if (!err) {
        pushStream.respondWithFile('style.css');
      }
    });

    stream.respondWithFile('index.html');
  }
});

server.listen(3000);
```

## Decision Matrix

### Choose WebSocket When:

- Bidirectional communication needed
- Low latency critical (< 50ms)
- High message frequency (> 1 msg/sec)
- Gaming, chat, collaborative editing
- Binary data transfer
- Custom protocol needed

### Choose SSE When:

- One-way server-to-client only
- Stock tickers, live feeds
- News/notifications
- Simpler implementation preferred
- Better proxy compatibility needed
- Automatic reconnection important

### Choose Long Polling When:

- Legacy browser support required (IE8/9)
- WebSocket blocked by firewall
- Very infrequent updates
- Fallback mechanism only

### Choose HTTP Streaming When:

- Large data transfers
- File uploads with progress
- Video/audio streaming
- One-way data flow

### Choose WebRTC When:

- Peer-to-peer communication
- Audio/video calls
- Screen sharing
- File transfer between peers
- Low latency P2P needed

## Hybrid Approach

```javascript
// Socket.IO with automatic fallback
const io = require('socket.io')(3000, {
  transports: ['websocket', 'polling'], // Try WebSocket first
  upgrade: true,
  allowUpgrades: true
});

io.on('connection', (socket) => {
  console.log('Connected via:', socket.conn.transport.name);

  socket.conn.on('upgrade', () => {
    console.log('Upgraded to:', socket.conn.transport.name);
  });
});
```

## Performance Characteristics

### Latency (p99)

- WebSocket: 5-20ms
- SSE: 10-50ms
- Long Polling: 100-500ms
- HTTP/2: 20-100ms

### Throughput (messages/sec)

- WebSocket: 10,000+ per connection
- SSE: 1,000+ per connection
- Long Polling: 1-10 per connection

### Connection Limits (per server)

- WebSocket: 50,000-100,000
- SSE: 50,000-100,000
- Long Polling: 10,000-20,000

### Overhead (per message)

- WebSocket: 2-6 bytes
- SSE: ~20 bytes
- Long Polling: 500-2000 bytes (HTTP headers)

## Migration Path

### From Polling to WebSocket

```javascript
// Step 1: Support both
app.get('/api/messages', (req, res) => {
  // Legacy polling endpoint
  res.json({ messages: getRecentMessages() });
});

io.on('connection', (socket) => {
  // New WebSocket endpoint
  socket.on('subscribe', (channel) => {
    socket.join(channel);
  });
});

// Step 2: Gradually migrate clients
// Step 3: Deprecate polling endpoint
```

### From SSE to WebSocket

```javascript
// SSE provides read-only, add WebSocket for writes
app.get('/events', sseHandler);  // Keep for reads

io.on('connection', (socket) => {
  socket.on('action', (data) => {
    // Handle writes via WebSocket
    processAction(data);
  });
});

// Eventually migrate reads to WebSocket too
```

## Best Practices

1. Start with simplest solution (SSE for one-way)
2. Use Socket.IO for automatic fallbacks
3. Monitor actual requirements before over-engineering
4. Consider mobile/network constraints
5. Implement graceful degradation
6. Load test before production
7. Have fallback strategy
8. Monitor connection success rates

---

## Reference: patterns

# WebSocket Patterns Reference

## Rooms and Namespaces

### Rooms (Channel Grouping)

```javascript
const io = require('socket.io')(3000);

io.on('connection', (socket) => {
  // Join a room
  socket.on('join-room', (roomId) => {
    socket.join(roomId);
    socket.emit('joined', { room: roomId });

    // Notify others in room
    socket.to(roomId).emit('user-joined', {
      userId: socket.id,
      timestamp: Date.now()
    });
  });

  // Leave a room
  socket.on('leave-room', (roomId) => {
    socket.leave(roomId);
    socket.to(roomId).emit('user-left', { userId: socket.id });
  });

  // Send to specific room
  socket.on('message', ({ roomId, text }) => {
    io.to(roomId).emit('message', {
      userId: socket.id,
      text,
      timestamp: Date.now()
    });
  });

  // Get all rooms a socket is in
  console.log('Socket rooms:', socket.rooms); // Set { socketId, roomId1, roomId2 }

  // Disconnect from all rooms
  socket.on('disconnect', () => {
    // Automatically leaves all rooms
  });
});

// Broadcast to all sockets in a room
io.to('room123').emit('announcement', 'Hello room!');

// Broadcast to multiple rooms
io.to('room1').to('room2').emit('multi-room', 'data');

// Broadcast to room except specific socket
socket.to('room123').emit('message', 'Others see this');

// Get all sockets in a room
const sockets = await io.in('room123').fetchSockets();
console.log(`Room has ${sockets.length} connections`);
```

### Namespaces (Logical Separation)

```javascript
// Admin namespace
const adminNs = io.of('/admin');
adminNs.on('connection', (socket) => {
  console.log('Admin connected:', socket.id);

  socket.on('admin-action', (data) => {
    // Admin-only events
  });
});

// Chat namespace
const chatNs = io.of('/chat');
chatNs.on('connection', (socket) => {
  console.log('Chat user connected:', socket.id);

  socket.on('message', (msg) => {
    chatNs.emit('message', msg); // Broadcast to all in /chat
  });
});

// Dynamic namespaces
io.of(/^\/workspace-\d+$/).on('connection', (socket) => {
  const namespace = socket.nsp;
  console.log(`Connected to ${namespace.name}`);

  socket.on('message', (data) => {
    namespace.emit('message', data);
  });
});
```

## Broadcasting Patterns

```javascript
// Broadcast to everyone including sender
io.emit('event', data);

// Broadcast to everyone except sender
socket.broadcast.emit('event', data);

// Broadcast to specific room
io.to('room1').emit('event', data);

// Broadcast to room except sender
socket.to('room1').emit('event', data);

// Broadcast to multiple rooms
io.to('room1').to('room2').emit('event', data);

// Broadcast to all connected clients in namespace
io.of('/namespace').emit('event', data);

// Volatile messages (ok to drop if client not ready)
socket.volatile.emit('high-frequency', data);

// Broadcast with acknowledgment (to all clients)
io.timeout(5000).emit('event', data, (err, responses) => {
  if (err) {
    console.log('Some clients did not acknowledge');
  } else {
    console.log('All clients acknowledged:', responses);
  }
});
```

## Acknowledgments

```javascript
// Server expects acknowledgment
socket.emit('question', 'Do you agree?', (answer) => {
  console.log('Client answered:', answer);
});

// Client sends acknowledgment
socket.on('question', (data, callback) => {
  console.log('Server asked:', data);
  callback('Yes, I agree');
});

// Timeout for acknowledgment
socket.timeout(5000).emit('request', data, (err, response) => {
  if (err) {
    console.log('Client did not acknowledge in time');
  } else {
    console.log('Got response:', response);
  }
});

// Error handling in acknowledgment
socket.on('save-data', (data, callback) => {
  try {
    saveToDatabase(data);
    callback({ success: true });
  } catch (error) {
    callback({ success: false, error: error.message });
  }
});
```

## Presence System

```javascript
const redis = require('ioredis');
const redisClient = new redis();

class PresenceManager {
  async userConnected(userId, socketId) {
    const key = `user:${userId}:sockets`;

    // Add socket to user's socket set
    await redisClient.sadd(key, socketId);

    // Set TTL to auto-cleanup stale connections
    await redisClient.expire(key, 3600);

    // Get total connections for user
    const socketCount = await redisClient.scard(key);

    // If first connection, mark user as online
    if (socketCount === 1) {
      await redisClient.hset('presence', userId, JSON.stringify({
        status: 'online',
        lastSeen: Date.now()
      }));

      // Notify friends
      const friends = await this.getUserFriends(userId);
      friends.forEach(friendId => {
        io.to(`user:${friendId}`).emit('presence', {
          userId,
          status: 'online'
        });
      });
    }

    return socketCount;
  }

  async userDisconnected(userId, socketId) {
    const key = `user:${userId}:sockets`;

    // Remove socket from set
    await redisClient.srem(key, socketId);

    const socketCount = await redisClient.scard(key);

    // If no more connections, mark offline
    if (socketCount === 0) {
      await redisClient.hset('presence', userId, JSON.stringify({
        status: 'offline',
        lastSeen: Date.now()
      }));

      const friends = await this.getUserFriends(userId);
      friends.forEach(friendId => {
        io.to(`user:${friendId}`).emit('presence', {
          userId,
          status: 'offline',
          lastSeen: Date.now()
        });
      });
    }

    return socketCount;
  }

  async getUserStatus(userId) {
    const data = await redisClient.hget('presence', userId);
    return data ? JSON.parse(data) : { status: 'offline' };
  }

  async getBulkPresence(userIds) {
    const pipeline = redisClient.pipeline();
    userIds.forEach(id => pipeline.hget('presence', id));

    const results = await pipeline.exec();
    return userIds.map((id, i) => ({
      userId: id,
      ...JSON.parse(results[i][1] || '{"status":"offline"}')
    }));
  }
}

// Usage
const presence = new PresenceManager();

io.on('connection', async (socket) => {
  const userId = socket.userId; // set by the JWT auth middleware in security.md

  await presence.userConnected(userId, socket.id);
  socket.join(`user:${userId}`);

  socket.on('disconnect', async () => {
    await presence.userDisconnected(userId, socket.id);
  });

  // Get presence for user's friends
  socket.on('get-presence', async (friendIds, callback) => {
    const presenceData = await presence.getBulkPresence(friendIds);
    callback(presenceData);
  });
});
```

## Message Queue Pattern

```javascript
// Queue messages when client disconnected
const messageQueue = new Map();

io.on('connection', (socket) => {
  const userId = socket.userId; // set by the JWT auth middleware in security.md

  // Deliver queued messages on connect
  const queuedMessages = messageQueue.get(userId) || [];
  if (queuedMessages.length > 0) {
    socket.emit('queued-messages', queuedMessages);
    messageQueue.delete(userId);
  }

  socket.on('disconnect', () => {
    // Mark user as disconnected
    setTimeout(() => {
      const userSockets = io.sockets.sockets;
      const hasOtherConnection = Array.from(userSockets.values())
        .some(s => s.userId === userId);

      if (!hasOtherConnection) {
        // User fully disconnected, queue new messages
        messageQueue.set(userId, []);
      }
    }, 1000);
  });
});

// Queue message if user offline
async function sendMessage(userId, message) {
  const userOnline = await isUserOnline(userId);

  if (userOnline) {
    io.to(`user:${userId}`).emit('message', message);
  } else {
    const queue = messageQueue.get(userId) || [];
    queue.push(message);
    messageQueue.set(userId, queue);

    // Persist to database for longer-term storage
    await saveMessageToDb(userId, message);
  }
}
```

## Pub/Sub Pattern

```javascript
const EventEmitter = require('events');

class MessageBus extends EventEmitter {
  constructor(io, redis) {
    super();
    this.io = io;
    this.redis = redis;
    this.setupSubscriptions();
  }

  setupSubscriptions() {
    // Subscribe to Redis channels
    this.redis.psubscribe('room:*', (err, count) => {
      console.log(`Subscribed to ${count} channels`);
    });

    this.redis.on('pmessage', (pattern, channel, message) => {
      const data = JSON.parse(message);
      const roomId = channel.split(':')[1];

      // Emit to Socket.IO room
      this.io.to(roomId).emit(data.event, data.payload);
    });
  }

  publish(roomId, event, payload) {
    // Publish to Redis (distributed across servers)
    this.redis.publish(
      `room:${roomId}`,
      JSON.stringify({ event, payload })
    );
  }
}

// Usage
const messageBus = new MessageBus(io, redisClient);

io.on('connection', (socket) => {
  socket.on('send-message', ({ roomId, text }) => {
    // Publish to all servers via Redis
    messageBus.publish(roomId, 'message', {
      userId: socket.id,
      text,
      timestamp: Date.now()
    });
  });
});
```

## Backpressure Handling

```javascript
io.on('connection', (socket) => {
  const MAX_BUFFER_SIZE = 10000;
  let bufferSize = 0;

  const originalEmit = socket.emit.bind(socket);

  socket.emit = function(event, ...args) {
    bufferSize++;

    if (bufferSize > MAX_BUFFER_SIZE) {
      console.warn('Buffer overflow, dropping message');
      return false;
    }

    const result = originalEmit(event, ...args);

    // Track buffer drain
    socket.once('drain', () => {
      bufferSize = 0;
    });

    return result;
  };

  // Monitor buffer size
  socket.on('drain', () => {
    console.log('Socket buffer drained');
  });
});
```

---

## Reference: protocol

# WebSocket Protocol Reference

## Protocol Basics

### Handshake Process

```
Client → Server: HTTP Upgrade Request
GET /socket.io/?EIO=4&transport=websocket HTTP/1.1
Host: example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
Sec-WebSocket-Version: 13

Server → Client: HTTP 101 Switching Protocols
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: HSmrc0sMlYUkAGmm5OPpG2HaGWk=
```

### Frame Structure

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - - +-------------------------------+
|                               |Masking-key, if MASK set to 1  |
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - - +
:                     Payload Data continued ...                :
+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
|                     Payload Data continued ...                |
+---------------------------------------------------------------+
```

## Opcodes

```javascript
const OPCODES = {
  CONTINUATION: 0x0, // Continuation frame
  TEXT: 0x1,         // Text frame
  BINARY: 0x2,       // Binary frame
  CLOSE: 0x8,        // Connection close
  PING: 0x9,         // Heartbeat ping
  PONG: 0xA          // Heartbeat pong
};
```

## Ping/Pong Mechanism

```javascript
// Server-side ping/pong with ws library
const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 8080 });

wss.on('connection', (ws) => {
  ws.isAlive = true;

  ws.on('pong', () => {
    ws.isAlive = true;
  });

  ws.on('message', (data) => {
    console.log('Received:', data);
  });
});

// Ping clients every 30 seconds
const interval = setInterval(() => {
  wss.clients.forEach((ws) => {
    if (ws.isAlive === false) {
      return ws.terminate();
    }

    ws.isAlive = false;
    ws.ping(); // Send ping frame
  });
}, 30000);

wss.on('close', () => {
  clearInterval(interval);
});
```

## Close Codes

```javascript
const CLOSE_CODES = {
  1000: 'Normal Closure',
  1001: 'Going Away',
  1002: 'Protocol Error',
  1003: 'Unsupported Data',
  1005: 'No Status Received',
  1006: 'Abnormal Closure',
  1007: 'Invalid Payload',
  1008: 'Policy Violation',
  1009: 'Message Too Big',
  1010: 'Mandatory Extension',
  1011: 'Internal Server Error',
  1015: 'TLS Handshake Fail'
};

// Proper close handling
ws.close(1000, 'Normal closure');
```

## Message Size Limits

```javascript
// Set max payload size (default 100MB)
const wss = new WebSocket.Server({
  port: 8080,
  maxPayload: 1024 * 1024 // 1MB
});

// Handle too large messages
ws.on('error', (error) => {
  if (error.message.includes('Max payload size exceeded')) {
    ws.close(1009, 'Message too big');
  }
});
```

## Compression (permessage-deflate)

```javascript
const wss = new WebSocket.Server({
  port: 8080,
  perMessageDeflate: {
    zlibDeflateOptions: {
      chunkSize: 1024,
      memLevel: 7,
      level: 3
    },
    zlibInflateOptions: {
      chunkSize: 10 * 1024
    },
    clientNoContextTakeover: true,
    serverNoContextTakeover: true,
    serverMaxWindowBits: 10,
    concurrencyLimit: 10,
    threshold: 1024 // Only compress messages > 1KB
  }
});
```

## Binary Data Handling

```javascript
// Send binary data
const buffer = Buffer.from([0x00, 0x01, 0x02, 0x03]);
ws.send(buffer, { binary: true });

// Receive binary data
ws.on('message', (data) => {
  if (data instanceof Buffer) {
    console.log('Received binary:', data);
  } else {
    console.log('Received text:', data.toString());
  }
});

// ArrayBuffer in browser
socket.binaryType = 'arraybuffer';
socket.onmessage = (event) => {
  if (event.data instanceof ArrayBuffer) {
    const view = new Uint8Array(event.data);
    console.log('Received:', view);
  }
};
```

## Protocol Comparison

| Feature | WebSocket | Socket.IO |
|---------|-----------|-----------|
| Protocol | Native WS | WS + fallbacks |
| Handshake | HTTP Upgrade | Engine.IO handshake |
| Reconnection | Manual | Automatic |
| Broadcasting | Manual | Built-in |
| Rooms | Manual | Built-in |
| Acknowledgments | Manual | Built-in |
| Binary | Native | Converted |
| Overhead | Minimal | Higher |
| Fallback | None | Long polling, SSE |

---

## Reference: scaling

# Horizontal Scaling Reference

## Architecture Overview

```
┌─────────────┐
│Load Balancer│ (nginx/HAProxy with sticky sessions)
└──────┬──────┘
       │
   ┌───┴───┐
   │       │
┌──▼──┐ ┌──▼──┐
│WS #1│ │WS #2│ ... (Socket.IO servers)
└──┬──┘ └──┬──┘
   │       │
   └───┬───┘
       │
   ┌───▼───┐
   │ Redis │ (Pub/Sub adapter)
   └───────┘
```

## Redis Adapter Configuration

### Socket.IO with Redis

```javascript
const { createServer } = require('http');
const { Server } = require('socket.io');
const { createAdapter } = require('@socket.io/redis-adapter');
const { createClient } = require('redis');

const httpServer = createServer();
const io = new Server(httpServer, {
  cors: { origin: '*' }
});

// Redis pub/sub client setup
const pubClient = createClient({
  host: 'localhost',
  port: 6379
});
const subClient = pubClient.duplicate();

Promise.all([pubClient.connect(), subClient.connect()]).then(() => {
  io.adapter(createAdapter(pubClient, subClient));
  console.log('Redis adapter connected');
});

// Now broadcasts work across all servers
io.emit('news', { hello: 'world' });

httpServer.listen(3000);
```

### Redis Streams for Reliable Delivery

```javascript
const { createAdapter } = require('@socket.io/redis-streams-adapter');

const redisClient = createClient({ url: 'redis://localhost:6379' });

redisClient.connect().then(() => {
  io.adapter(createAdapter(redisClient, {
    streamName: 'socket.io-stream',
    maxLen: 10000, // Keep last 10k messages
    readCount: 100 // Process 100 messages at a time
  }));
});
```

## Sticky Sessions

### Nginx Configuration

```nginx
upstream websocket_backend {
    ip_hash; # Sticky sessions based on IP
    server ws1.example.com:3000;
    server ws2.example.com:3000;
    server ws3.example.com:3000;
}

server {
    listen 80;
    server_name example.com;

    location /socket.io/ {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}
```

### HAProxy Configuration

```haproxyconf
frontend websocket_frontend
    bind *:80
    mode http
    option httplog
    use_backend websocket_backend

backend websocket_backend
    mode http
    balance source # Sticky sessions by source IP
    hash-type consistent # Consistent hashing

    # Health checks
    option httpchk GET /health
    http-check expect status 200

    server ws1 10.0.1.1:3000 check
    server ws2 10.0.1.2:3000 check
    server ws3 10.0.1.3:3000 check
```

### Cookie-based Sticky Sessions

```javascript
// Server-side: Set affinity cookie
io.engine.on('connection', (rawSocket) => {
  const serverID = process.env.SERVER_ID || 'server1';
  rawSocket.request.res.setHeader(
    'Set-Cookie',
    `io=${serverID}; Path=/; HttpOnly; SameSite=Lax`
  );
});
```

```nginx
# Nginx: Use cookie for routing
upstream websocket_backend {
    server ws1.example.com:3000;
    server ws2.example.com:3000;
}

map $cookie_io $backend_server {
    "server1" ws1.example.com:3000;
    "server2" ws2.example.com:3000;
    default websocket_backend;
}

location /socket.io/ {
    proxy_pass http://$backend_server;
    # ... other proxy settings
}
```

## State Management

### Shared State in Redis

```javascript
const Redis = require('ioredis');
const redis = new Redis();

// Store user connection info
io.on('connection', async (socket) => {
  const userId = socket.userId; // set by the JWT auth middleware in security.md

  // Track which server has this user
  await redis.hset('user:connections', userId, process.env.SERVER_ID);

  // Store user presence
  await redis.hset(`user:${userId}`, {
    socketId: socket.id,
    serverId: process.env.SERVER_ID,
    connectedAt: Date.now(),
    status: 'online'
  });

  socket.on('disconnect', async () => {
    await redis.hdel('user:connections', userId);
    await redis.del(`user:${userId}`);
  });
});

// Send message to specific user across cluster
async function sendToUser(userId, event, data) {
  const serverId = await redis.hget('user:connections', userId);

  if (serverId === process.env.SERVER_ID) {
    // User is on this server
    const sockets = await io.in(`user:${userId}`).fetchSockets();
    sockets.forEach(socket => socket.emit(event, data));
  } else {
    // User is on another server - use Redis to route
    io.to(`user:${userId}`).emit(event, data);
  }
}
```

## Connection Limits

### Per-Server Limits

```javascript
const MAX_CONNECTIONS = 50000;

io.engine.on('connection', (socket) => {
  const currentConnections = io.engine.clientsCount;

  if (currentConnections > MAX_CONNECTIONS) {
    socket.close(1008, 'Server at capacity');
    return;
  }
});
```

### Kubernetes Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: websocket-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: websocket-server
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: websocket_connections
      target:
        type: AverageValue
        averageValue: "40000" # Scale when avg > 40k connections/pod
```

## Graceful Shutdown

```javascript
const gracefulShutdown = () => {
  console.log('Shutting down gracefully...');

  // Stop accepting new connections
  io.close(() => {
    console.log('All connections closed');
    process.exit(0);
  });

  // Force close after 30 seconds
  setTimeout(() => {
    console.error('Forcing shutdown after timeout');
    process.exit(1);
  }, 30000);
};

process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);
```

## Performance Optimization

### Node.js Clustering

```javascript
const cluster = require('cluster');
const os = require('os');

if (cluster.isMaster) {
  const numWorkers = os.cpus().length;

  console.log(`Master ${process.pid} starting ${numWorkers} workers`);

  for (let i = 0; i < numWorkers; i++) {
    cluster.fork();
  }

  cluster.on('exit', (worker) => {
    console.log(`Worker ${worker.process.pid} died, spawning new`);
    cluster.fork();
  });
} else {
  // Worker process runs Socket.IO server
  const io = require('./socket-server');
  io.listen(3000);
  console.log(`Worker ${process.pid} started`);
}
```

### uWebSockets.js for Maximum Performance

```javascript
const uWS = require('uWebSockets.js');

const app = uWS.App()
  .ws('/*', {
    compression: uWS.SHARED_COMPRESSOR,
    maxPayloadLength: 16 * 1024,
    idleTimeout: 60,

    open: (ws) => {
      console.log('Client connected');
    },

    message: (ws, message, isBinary) => {
      // Echo message
      ws.send(message, isBinary);
    },

    close: (ws, code, message) => {
      console.log('Client disconnected');
    }
  })
  .listen(9001, (token) => {
    if (token) {
      console.log('Listening on port 9001');
    }
  });
```

---

## Reference: security

# WebSocket Security Reference

## Authentication

### JWT Authentication

```javascript
const io = require('socket.io')(3000);
const jwt = require('jsonwebtoken');

// Middleware for authentication
io.use((socket, next) => {
  const token = socket.handshake.auth.token;

  if (!token) {
    return next(new Error('Authentication error: No token provided'));
  }

  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    socket.userId = decoded.userId;
    socket.username = decoded.username;
    next();
  } catch (err) {
    next(new Error('Authentication error: Invalid token'));
  }
});

io.on('connection', (socket) => {
  console.log(`User ${socket.username} connected`);

  socket.on('message', (data) => {
    // socket.userId is already verified
    saveMessage(socket.userId, data);
  });
});
```

### Query Parameter Authentication (Less Secure)

```javascript
// Use only for initial handshake, then upgrade to token
io.use((socket, next) => {
  const token = socket.handshake.query.token;

  if (!token) {
    return next(new Error('Authentication required'));
  }

  jwt.verify(token, process.env.JWT_SECRET, (err, decoded) => {
    if (err) return next(new Error('Invalid token'));
    socket.userId = decoded.userId;
    next();
  });
});
```

### Cookie Authentication

```javascript
const cookieParser = require('cookie-parser');

io.use((socket, next) => {
  const cookies = socket.handshake.headers.cookie;

  if (!cookies) {
    return next(new Error('No cookies'));
  }

  // Parse cookies
  cookieParser(process.env.COOKIE_SECRET)(
    socket.request,
    {},
    () => {
      const sessionId = socket.request.signedCookies.sessionId;

      if (!sessionId) {
        return next(new Error('No session'));
      }

      // Verify session in Redis/DB
      verifySession(sessionId).then(user => {
        socket.userId = user.id;
        next();
      }).catch(err => {
        next(new Error('Invalid session'));
      });
    }
  );
});
```

## Authorization

### Room-Based Authorization

```javascript
io.on('connection', (socket) => {
  socket.on('join-room', async (roomId) => {
    // Check if user has permission
    const hasAccess = await checkRoomAccess(socket.userId, roomId);

    if (!hasAccess) {
      socket.emit('error', { message: 'Access denied to room' });
      return;
    }

    socket.join(roomId);
    socket.emit('joined', { room: roomId });
  });

  socket.on('send-message', async ({ roomId, text }) => {
    // Verify user is in room
    if (!socket.rooms.has(roomId)) {
      socket.emit('error', { message: 'Not in room' });
      return;
    }

    // Check write permissions
    const canWrite = await checkWritePermission(socket.userId, roomId);

    if (!canWrite) {
      socket.emit('error', { message: 'No write permission' });
      return;
    }

    io.to(roomId).emit('message', {
      userId: socket.userId,
      text,
      timestamp: Date.now()
    });
  });
});
```

### Admin-Only Events

```javascript
const ADMIN_EVENTS = ['kick-user', 'ban-user', 'delete-message'];

io.use((socket, next) => {
  // Attach role to socket after auth
  getUserRole(socket.userId).then(role => {
    socket.role = role;
    next();
  });
});

io.on('connection', (socket) => {
  ADMIN_EVENTS.forEach(event => {
    socket.on(event, async (data) => {
      if (socket.role !== 'admin') {
        socket.emit('error', { message: 'Admin access required' });
        return;
      }

      // Execute admin action
      await handleAdminAction(event, data);
    });
  });
});
```

## Rate Limiting

### Per-Socket Rate Limiting

```javascript
const rateLimit = require('express-rate-limit');

class SocketRateLimiter {
  constructor(maxRequests = 100, windowMs = 60000) {
    this.maxRequests = maxRequests;
    this.windowMs = windowMs;
    this.requests = new Map();
  }

  check(socketId) {
    const now = Date.now();
    const userRequests = this.requests.get(socketId) || [];

    // Remove expired requests
    const validRequests = userRequests.filter(
      time => now - time < this.windowMs
    );

    if (validRequests.length >= this.maxRequests) {
      return false; // Rate limit exceeded
    }

    validRequests.push(now);
    this.requests.set(socketId, validRequests);
    return true;
  }

  reset(socketId) {
    this.requests.delete(socketId);
  }
}

const limiter = new SocketRateLimiter(100, 60000); // 100 req/min

io.on('connection', (socket) => {
  socket.on('message', (data) => {
    if (!limiter.check(socket.id)) {
      socket.emit('error', { message: 'Rate limit exceeded' });
      return;
    }

    // Process message
    io.to(data.roomId).emit('message', data);
  });

  socket.on('disconnect', () => {
    limiter.reset(socket.id);
  });
});
```

### Redis-Based Distributed Rate Limiting

```javascript
const Redis = require('ioredis');
const redis = new Redis();

async function checkRateLimit(userId, maxRequests = 100, windowSec = 60) {
  const key = `rate_limit:${userId}`;
  const now = Date.now();
  const windowStart = now - (windowSec * 1000);

  const pipeline = redis.pipeline();

  // Remove old entries
  pipeline.zremrangebyscore(key, 0, windowStart);

  // Count requests in window
  pipeline.zcard(key);

  // Add current request
  pipeline.zadd(key, now, `${now}-${Math.random()}`);

  // Set expiry
  pipeline.expire(key, windowSec);

  const results = await pipeline.exec();
  const count = results[1][1];

  return count < maxRequests;
}

io.on('connection', (socket) => {
  socket.on('message', async (data) => {
    const allowed = await checkRateLimit(socket.userId, 50, 60);

    if (!allowed) {
      socket.emit('error', { message: 'Too many requests' });
      return;
    }

    io.to(data.roomId).emit('message', data);
  });
});
```

## CORS Configuration

```javascript
const io = require('socket.io')(3000, {
  cors: {
    origin: ['https://example.com', 'https://app.example.com'],
    methods: ['GET', 'POST'],
    credentials: true,
    allowedHeaders: ['Authorization']
  }
});

// Dynamic CORS
io.engine.on('initial_headers', (headers, req) => {
  headers['Access-Control-Allow-Origin'] = req.headers.origin;
});
```

## Input Validation

```javascript
const Joi = require('joi');

const messageSchema = Joi.object({
  roomId: Joi.string().uuid().required(),
  text: Joi.string().min(1).max(1000).required(),
  attachments: Joi.array().items(Joi.string().uri()).max(5).optional()
});

io.on('connection', (socket) => {
  socket.on('message', (data) => {
    // Validate input
    const { error, value } = messageSchema.validate(data);

    if (error) {
      socket.emit('error', {
        message: 'Invalid message format',
        details: error.details
      });
      return;
    }

    // Process validated data
    io.to(value.roomId).emit('message', {
      userId: socket.userId,
      ...value,
      timestamp: Date.now()
    });
  });
});
```

## XSS Protection

```javascript
const sanitizeHtml = require('sanitize-html');

function sanitizeMessage(text) {
  return sanitizeHtml(text, {
    allowedTags: [], // Strip all HTML
    allowedAttributes: {},
    disallowedTagsMode: 'escape'
  });
}

io.on('connection', (socket) => {
  socket.on('message', (data) => {
    const sanitized = {
      ...data,
      text: sanitizeMessage(data.text)
    };

    io.to(data.roomId).emit('message', sanitized);
  });
});
```

## DDoS Protection

### Connection Limiting

```javascript
const connectionLimits = new Map();
const MAX_CONNECTIONS_PER_IP = 10;

io.engine.on('connection', (rawSocket) => {
  const ip = rawSocket.request.headers['x-forwarded-for'] ||
              rawSocket.request.connection.remoteAddress;

  const currentConnections = connectionLimits.get(ip) || 0;

  if (currentConnections >= MAX_CONNECTIONS_PER_IP) {
    rawSocket.close(1008, 'Too many connections from IP');
    return;
  }

  connectionLimits.set(ip, currentConnections + 1);

  rawSocket.on('close', () => {
    const count = connectionLimits.get(ip) - 1;
    if (count <= 0) {
      connectionLimits.delete(ip);
    } else {
      connectionLimits.set(ip, count);
    }
  });
});
```

### Message Size Limits

```javascript
const io = require('socket.io')(3000, {
  maxHttpBufferSize: 1e6, // 1MB max message size
  pingTimeout: 60000,
  pingInterval: 25000
});

io.on('connection', (socket) => {
  socket.on('message', (data) => {
    if (JSON.stringify(data).length > 10000) {
      socket.emit('error', { message: 'Message too large' });
      return;
    }

    // Process message
  });
});
```

## Secure Session Management

```javascript
const sessions = new Map();

io.on('connection', (socket) => {
  const sessionId = generateSecureSessionId();

  sessions.set(socket.id, {
    sessionId,
    userId: socket.userId,
    createdAt: Date.now(),
    lastActivity: Date.now()
  });

  // Timeout inactive sessions; each message restarts the 30-minute window
  let timeout;
  const resetTimeout = () => {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      socket.disconnect(true);
    }, 30 * 60 * 1000); // 30 minutes
  };
  resetTimeout();

  socket.on('message', () => {
    const session = sessions.get(socket.id);
    if (session) {
      session.lastActivity = Date.now();
      resetTimeout();
    }
  });

  socket.on('disconnect', () => {
    sessions.delete(socket.id);
    clearTimeout(timeout);
  });
});

function generateSecureSessionId() {
  return require('crypto').randomBytes(32).toString('hex');
}
```

## Audit Logging

```javascript
const winston = require('winston');

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.File({ filename: 'websocket-audit.log' })
  ]
});

io.on('connection', (socket) => {
  logger.info('Connection', {
    socketId: socket.id,
    userId: socket.userId,
    ip: socket.handshake.address,
    timestamp: Date.now()
  });

  socket.on('message', (data) => {
    logger.info('Message', {
      socketId: socket.id,
      userId: socket.userId,
      roomId: data.roomId,
      messageLength: data.text.length,
      timestamp: Date.now()
    });
  });

  socket.on('disconnect', (reason) => {
    logger.info('Disconnect', {
      socketId: socket.id,
      userId: socket.userId,
      reason,
      timestamp: Date.now()
    });
  });
});
```
