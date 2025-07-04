document.addEventListener("DOMContentLoaded", async () => {
  const token = localStorage.getItem("token");
  if (!token) {
    alert("Unauthorized! Please login.");
    window.location.href = "login.html";
    return;
  }
  try {
    const response = await fetch(
      "http://localhost:8080/api/chat/unread-users",
      {
        method: "GET",
        headers: {
          Authorization: token,
        },
      }
    );
    const unreadUsersList = document.getElementById("contactsList");
    const data = await response.json();
    console.log(data);
    if (response.ok) {
      unreadUsersList.innerHTML = "";

      data.users.forEach((contact) => {
        const li = document.createElement("li");
        li.id = contact.id;
        li.className = "contact-item";
        li.innerHTML = `
          <button class="chat-btn" onclick="openChat('${
            contact.id
          }')">💬</button>
          <span class="contact-name">${contact.username || "Unknown"}</span>
        `;
        unreadUsersList.appendChild(li);
      });
    } else {
      console.error("Error fetching list of unread users:", data.error);
    }
  } catch (error) {
    console.error("Failed to fetch list of unread users:", error);
  }
});
async function openChat(contactId) {
  if (!contactId) {
    console.error("Error: contactId is undefined");
    alert("Error opening chat. Contact ID is missing.");
    return;
  }
  localStorage.setItem("receiver_id", contactId);

  const token = localStorage.getItem("token");

  try {
    const response = await fetch(
      `http://localhost:8080/api/chat/user?id=${encodeURIComponent(contactId)}`,
      {
        method: "GET",
        headers: {
          // "Content-Type": "application/json",
          Authorization: token,
        },
      }
    );
    const data = await response.json();
    if (response.ok) {
      localStorage.setItem("receiver_name", data.receiver); // Store receiver's name
    } else {
      console.log("Error fetching receiver name:", data.error);
      alert("Error fetching receiver's name. Please try again.");
      localStorage.setItem("receiver_name", "Unknown"); // Fallback if not found]
    }
  } catch (error) {
    alert(error.message || "Error fetching receiver's name. Please try again!");
    localStorage.setItem("receiver_name", "Unknown");
    window.location.reload();
  }
  window.location.href = "chat.html"; // Redirect to chat window
  // Call loadChatHistory after a small delay to ensure the new page is loaded

  setTimeout(() => {
    const senderId = localStorage.getItem("user_id"); // Current user (logged-in)
    const receiverId = localStorage.getItem("receiver_id"); // Contact ID

    if (senderId && receiverId) {
      loadChatHistory(senderId, receiverId);
    } else {
      console.error("Sender or Receiver ID missing.");
    }
  }, 500);
}
