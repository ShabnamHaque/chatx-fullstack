const chatMessages = document.getElementById("chatMessages");
const messageInput = document.getElementById("messageInput");
const toggleButton = document.getElementById("toggleButton");
const sendButton = document.getElementById("sendButton");
const userID = localStorage.getItem("user_id"); // Retrieve logged-in user ID
const receiver_id = localStorage.getItem("receiver_id");
const token = localStorage.getItem("token");
const receiver_name = localStorage.getItem("receiver_name"); // Get receiver name
let isDisappearing = false; // Default OFF
let socket;

function goBackToContacts() {
  localStorage.removeItem("receiver_id");
  localStorage.removeItem("receiver_name");
  window.location.href = "contacts.html";
}

async function sendMessage() {
  const messageInput = document.getElementById("messageInput").value;
  const messageData = {
    sender_id: userID, // Replace with actual sender ID
    receiver_id: receiver_id, // Replace with actual receiver ID
    content: messageInput.trim(),
    disappearing: isDisappearing,
    unread: true,
  };
  if (socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(messageData)); // WebSocket communication
    console.log("📩 Sent message via WebSocket");
    console.log(messageData.unread);
    console.log(messageData.SenderID);
    appendMessage(messageData, messageData.SenderID !== userID);
  } else {
    console.warn("⚠️ WebSocket not connected!");
  }
  try {
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

function appendMessage(message, isSender) {
  if (!message || !message.content) {
    console.error("⚠️ Message content is missing!", message);
    return; // Stop execution if there's no content
  }
  const messageElement = document.createElement("div");
  messageElement.classList.add("message");

  if (!isSender) {
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
  messageElement.textContent = message.content;
  chatMessages.appendChild(messageElement);
  chatMessages.scrollTop = chatMessages.scrollHeight; // Auto-scroll to bottom
}

async function loadChatHistory(senderId, receiverId) {
  try {
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
        appendMessage(msg, msg.sender_id === userID);
      });
    }
  } catch (error) {
    console.error("⚠️ Error loading chat history:", error);
  }
}

const toggleDisappearing = document.getElementById("toggleButton");
let disappearingMessages = false;
toggleDisappearing.addEventListener("click", () => {
  disappearingMessages = !disappearingMessages;

  toggleButton.classList.toggle("active", disappearingMessages);
  localStorage.setItem("disappearing", disappearingMessages);

  console.log("Disappearing Messages:", disappearingMessages);
});

// document.addEventListener("DOMContentLoaded", function () {
//   const backButton = document.getElementById("back-button-chat");
//   if (backButton) {
//     backButton.addEventListener("click", goBackToContacts);
//   } else {
//     console.error("❌ Back button not found in DOM");
//   }
// });

// new
/*
let socket; // declare globally so everyone can use it

document.addEventListener("DOMContentLoaded", function () {
  socket = new WebSocket(
    `ws://localhost:8080/api/chat/ws?receiver_id=${encodeURIComponent(
      receiver_id
    )}&sender_id=${encodeURIComponent(userID)}&token=${encodeURIComponent(
      token
    )}`
  );

  socket.onopen = function () {
    console.log("✅ WebSocket connection established");
  };

  socket.onmessage = (event) => {
    console.log("📩 Message received:", event.data);
    const message = JSON.parse(event.data);
    appendMessage(message);
  };

  socket.onerror = function (error) {
    console.error("❌ WebSocket Error:", error);
  };

  sendButton.addEventListener("click", sendMessage);
});
*/
//new
document.addEventListener("DOMContentLoaded", function () {
  loadChatHistory(userID, receiver_id);

  document.getElementById("receiver_name").innerText =
    `Chat with ${receiver_name}` || "Chat";

  socket = new WebSocket(
    `ws://localhost:8080/api/chat/ws?receiver_id=${encodeURIComponent(
      receiver_id
    )}&sender_id=${encodeURIComponent(userID)}&token=${encodeURIComponent(
      token
    )}`
  );

  toggleButton.addEventListener("click", function () {
    isDisappearing = !isDisappearing;
    if (isDisappearing) {
      toggleButton.classList.add("active");
    } else {
      toggleButton.classList.remove("active");
    }
  });

  socket.onopen = function () {
    console.log("✅ WebSocket connection established"); //executing
  };
  socket.onmessage = (event) => {
    console.log("📩 Message received:", event.data);
    const message = JSON.parse(event.data);
    appendMessage(message, message.sender_id === userID); // Append received message
  };
  socket.onerror = function (error) {
    console.error("❌ WebSocket Error:", error);
  };

  sendButton.addEventListener("click", sendMessage);
});
window.onload = () => {
  const senderId = localStorage.getItem("user_id");
  const receiverId = localStorage.getItem("receiver_id");

  if (!senderId || !receiverId) {
    console.error("❌ Missing sender or receiver ID");
    return;
  }
  loadChatHistory(senderId, receiverId);

  let storedValue = localStorage.getItem("disappearing");
  if (storedValue === null) {
    localStorage.setItem("disappearing", "false"); // Set default to OFF
    disappearingMessages = false;
  } else {
    //disappearingMessages = storedValue === "true";
  }
  if (disappearingMessages) toggleDisappearing.classList.add("active");
};
