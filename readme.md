# Chat Application

This is a simple chat application built with Golang for the backend and vanilla HTML, CSS, and JavaScript for the frontend. The app uses WebSockets for real-time communication and MongoDB as the database. Users can register, log in, and chat with other users in private channels.

## Features

- User registration and login with JWT authentication
- Real-time messaging via WebSockets
- Private channels, where only users with allowed IDs can join
- 24 hr Disappearing messages feature for added privacy

## Technologies

- **Backend**: Go (Golang)
  - WebSockets for real-time communication
  - JWT for user authentication
  - MongoDB for data storage
- **Frontend**: HTML, CSS, JavaScript
  - WebSocket integration on the client-side
