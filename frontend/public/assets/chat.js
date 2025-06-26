document.addEventListener("DOMContentLoaded", function () {
  const chatMessages = document.getElementById("chatMessages");
  const messageInput = document.getElementById("messageInput");
  const toggleButton = document.getElementById("toggleButton");
  const sendButton = document.getElementById("sendButton");
  const userID = localStorage.getItem("user_id"); // Retrieve logged-in user ID
  const receiver_id = localStorage.getItem("receiver_id");
  const receiver_name = localStorage.getItem("receiver_name"); // Get receiver name
  let isDisappearing = false; // Default OFF
  loadChatHistory(userID, receiver_id);

  toggleButton.addEventListener("click", function () {
    isDisappearing = !isDisappearing;
    if (isDisappearing) {
      toggleButton.classList.add("active");
    } else {
      toggleButton.classList.remove("active");
    }
  });
  document.getElementById("receiverName").innerText =
    `Chat with ${receiver_name}` || "Chat";
  const socket = new WebSocket(
    "ws://localhost:8080/api/chat/ws?receiver_id=" +
      receiver_id +
      "&sender_id=" +
      userID
  );
  socket.onopen = function () {
    console.log("✅ WebSocket connection established");
  };
  socket.onerror = function (error) {
    console.error("❌ WebSocket Error:", error);
  };
  socket.onmessage = (event) => {
    console.log("📩 Message received:", event.data);
    const message = JSON.parse(event.data);
    appendMessage(message); // Append received message
  };
  socket.onerror = (error) => {
    console.error("❌ WebSocket Error:", error);
  };
  sendButton.addEventListener("click", sendMessage);
  async function sendMessage() {
    const token = localStorage.getItem("token");
    const messageInput = document.getElementById("messageInput").value;
    const messageData = {
      sender_id: localStorage.getItem("user_id"), // Replace with actual sender ID
      receiver_id: localStorage.getItem("receiver_id"), // Replace with actual receiver ID
      content: messageInput,
      disappearing: isDisappearing,
      unread: true,
    };
    if (socket.readyState === WebSocket.OPEN) {
      //to send via ws conn
      socket.send(JSON.stringify(messageData)); // WebSocket communication
      console.log("📩 Sent message via WebSocket");
      console.log(messageData.unread);
      //  console.log(messageData.SenderID);
    } else {
      console.warn("⚠️ WebSocket not connected!");
    }
    try {
      //function to store in DB
      const response = await fetch(`http://localhost:8080/api/chat/send`, {
        method: "POST", //to initMessageHandler - ws server in backend
        headers: {
          "Content-Type": "application/json",
          Authorization: token, // Send JWT Token
        },
        body: JSON.stringify(messageData),
      });
      const data = await response.json();
      if (response.ok) {
        console.log("✅ Message stored in database:", data);
      } else {
        console.error("❌ Error storing message:", data.error);
      }
    } catch (error) {
      console.error("❌ HTTP request failed:", error);
    }
    document.getElementById("messageInput").value = "";
  }

  function appendMessage(message) {
    console.log("Appending message:", message);

    if (!message || !message.Content) {
      console.error("⚠️ Message content is missing!", message);
      return; // Stop execution if there's no content
    }
    const messageElement = document.createElement("div");
    messageElement.classList.add("message");

    const isOwnMessage = message.sender_id === localStorage.getItem("user_id"); // Check if message is from the logged-in user
    if (!isOwnMessage) {
      messageElement.classList.add("read"); // You can omit this if you don't want to display a tick for outgoing messages.
      messageElement.style.alignSelf = "flex-start";
      messageElement.style.backgroundColor = "#ffff"; //white for texts - incoming
    } else {
      if (!message.Unread) {
        // fetch the read status from backend and update
        messageElement.classList.add("read"); // Mark as read if Unread is false
      } else {
        messageElement.classList.add("unread"); // Otherwise, mark as unread
      }
      messageElement.style.alignSelf = "flex-end";
      messageElement.style.backgroundColor = "#4d6f94"; //blue for outgoing texts
      messageElement.style.color = "#ffff";
    }
    messageElement.style.padding = "8px 12px";
    messageElement.style.borderRadius = "10px";
    messageElement.style.margin = "5px 0";
    messageElement.textContent = message.Content;
    chatMessages.appendChild(messageElement);
    chatMessages.scrollTop = chatMessages.scrollHeight; // Auto-scroll to bottom
  }
  async function loadChatHistory(senderId, receiverId) {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        console.error("⚠️ No token found in localStorage!");
        return; // Stop execution if there's no token
      }
      const response = await fetch(
        `http://localhost:8080/api/chat/history?sender_id=${encodeURIComponent(
          senderId
        )}&receiver_id=${encodeURIComponent(receiverId)}`,
        {
          method: "GET", //get an array of messages between the two users
          headers: {
            //  "Content-Type": "application/json",
            Authorization: token,
          },
        }
      );
      if (!response.ok) {
        const errorText = await response.text(); // Read response error message
        console.error(
          `⚠️ HTTP Error! Status: ${response.status}, Response: ${errorText}`
        );
        throw new Error(
          `HTTP error! Status: ${response.status}, Response: ${errorText}`
        );
      }
      const data = await response.json();
      if (data.messages) {
        //data.messages comes from backend gin.H{"messages":messages}
        document.getElementById("chatMessages").innerHTML = ""; // Clear old messages
        data.messages.forEach((msg) => {
          // appendMessage(msg, msg.sender_id === senderId);
          appendMessage(msg);
        });
      }
    } catch (error) {
      console.error("⚠️ Error loading chat history:", error);
    }
  }
});
