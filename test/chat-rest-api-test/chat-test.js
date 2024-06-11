const { login } = require("./utils/pax_login");
const { requestHelper } = require("./utils/utils");

const authenticateAndCreateRoom = async (email, acceptorId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: "https://go.myru.online/api/chat/createRoom",
      method: "POST",
      data: {
        acceptorId: acceptorId,
        initialMessage: "Hi",
      },
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const authenticateAndGetSubscribedRooms = async (email) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: "https://go.myru.online/api/chat/rooms",
      method: "GET",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const authenticateAndGetUnsubscribedNewRooms = async (email) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: "https://go.myru.online/api/chat/newRooms",
      method: "GET",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const authenticateAndSubscribe = async (email, roomId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/subscribe/${roomId}`,
      method: "PATCH",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const authenticateAndUnSubscribe = async (email, roomId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/unsubscribe/${roomId}`,
      method: "PATCH",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const authenticateAndGetRoomDetails = async (email, roomId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/room/${roomId}`,
      method: "GET",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const sendMessage = async (email, roomId, messageContent) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/message/${roomId}`,
      method: "POST",
      data: {
        content: messageContent,
      },
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const editMessage = async (email, messageId, newContent) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/message/${messageId}`,
      method: "PATCH",
      data: {
        content: newContent,
      },
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const deleteMessage = async (email, messageId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/message/${messageId}`,
      method: "DELETE",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    closeWebSocket();
  }
};

const getAllMessages = async (email, roomId) => {
  const { token, session, closeWebSocket } = await login({
    email: email,
    password: "123123",
  });

  try {
    const res = await requestHelper({
      url: `https://go.myru.online/api/chat/message/${roomId}`,
      method: "GET",
      token: token,
      session: session,
    });
    console.log(JSON.stringify(res, null, 2));
    return res;
  } catch (error) {
    console.error("Error in get all messages:", error.message);
  } finally {
    closeWebSocket();
  }
};

const users = [
  {
    email: "demir.hoogerwerf@kpnmail.nl",
    userId: "19128500-44dd-4f43-af2e-d029e818570e",
  },
  {
    email: "eddie.rose@gmail.com",
    userId: "db3734b0-5547-4386-aa3f-d7255bdd3152",
  },
  {
    email: "elio.arnaud@outlook.com",
    userId: "da979770-a98b-4868-8bd9-a7a3e4c5520a",
  },
  {
    userId: "f69f6282-484b-4aca-b2d9-1ebd6c34415b",
    email: "bronislava.gordiienko@email.ua",
  },
  {
    userId: "f73d1253-c9af-4d69-a9c9-6cbff214814d",
    email: "ege.adivar@yahoo.com.tr",
  },
  {
    userId: "083b0e11-c3a6-4054-b01e-1253ff797fc8",
    email: "vickie.harris@gmail.com",
  },
];

async function runTestCases() {
  try {
    // Case: Initial Room Creation
    console.log("Attempting to create a room with Demir and Eddie...");
    const roomCreateResp = await authenticateAndCreateRoom(
      users[0].email,
      users[1].userId,
    );
    if (roomCreateResp.status !== "success" || !roomCreateResp.data.room.ID)
      throw new Error("Room creation failed");
    const roomId = roomCreateResp.data.room.ID;
    console.log(`Room created successfully with ID: ${roomId}`);

    // Case: Attempt to Recreate the Same Room
    console.log("Attempting to recreate the same room...");
    try {
      await authenticateAndCreateRoom(users[0].email, users[1].userId);
      console.error(
        "Room was recreated; the system failed to prevent duplicate rooms.",
      );
    } catch (error) {
      console.log(
        "Expected error preventing room recreation received:",
        error.message,
      );
    }

    // Case: Attempt to Create a Room in Reverse Order
    console.log(
      "Attempting to create a room with Eddie and Demir (reverse order)...",
    );
    try {
      await authenticateAndCreateRoom(users[1].email, users[0].userId);
      console.error(
        "Room was recreated in reverse order; the system failed to prevent duplicate rooms.",
      );
    } catch (error) {
      console.log(
        "Expected error preventing room recreation in reverse order received:",
        error.message,
      );
    }

    // Case: Attempt to send a message before other members subscribe to check for errors
    try {
      console.log("Attempting to send a message without full subscription...");
      await sendMessage(users[0].email, roomId, "Hello, world!");
    } catch (error) {
      console.log(
        "Expected error on sending message without full subscription:",
        error.message,
      );
    }

    // Case: Subscribe all members
    console.log("Subscribing Eddie to the room...");
    await authenticateAndSubscribe(users[1].email, roomId);
    console.log("All members subscribed.");

    // Case: Send new message
    console.log("Attempting to send a message after full subscription...");
    const sendMessageResp = await sendMessage(
      users[0].email,
      roomId,
      "Hello from Demir!",
    );
    if (!sendMessageResp.data.message.ID)
      throw new Error("Sending message failed");
    const messageId = sendMessageResp.data.message.ID;
    console.log(`Message sent successfully with ID: ${messageId}`);

    // Case: Multiple Messages and Ordering
    console.log("Sending multiple messages to check ordering...");
    const messageIds = [];
    for (let i = 0; i < 3; i++) {
      const resp = await sendMessage(users[0].email, roomId, `Message ${i}`);
      if (!resp.data.message.ID) throw new Error("Sending message failed");
      messageIds.push(resp.data.message.ID);
    }

    // Case: Attempt to send a message by a non-member
    console.log("Attempting to send a message by a non-member...");
    try {
      await sendMessage(users[2].email, roomId, "This should not be possible!");
      console.error("Test failed: Non-member was able to send a message.");
    } catch (error) {
      console.log(
        "Test passed: Non-member could not send a message. Error:",
        error.message,
      );
    }

    // Case: Edit the message and verify changes
    console.log("Editing the sent message...");
    const editMessageResp = await editMessage(
      users[0].email,
      messageId,
      "Updated message content",
    );
    // Assuming editMessage function returns the updated message or status. Adjust according to your implementation.
    if (editMessageResp.status !== "success")
      throw new Error("Editing message failed");

    // Case: Delete the message and verify deletion
    console.log("Deleting the message...");
    const deleteMessageResp = await deleteMessage(users[0].email, messageId);
    console.log(deleteMessageResp);
    // Assuming deleteMessage function returns a status. Adjust according to your implementation.
    if (deleteMessageResp.status !== "success")
      throw new Error("Deleting message failed");

    // Case: Attempt to edit a message by a non-owner
    console.log("Attempting to edit a message by a non-owner...");
    try {
      await editMessage(users[1].email, messageId, "Illegal edit attempt!");
      console.error("Test failed: Non-owner was able to edit the message.");
    } catch (error) {
      console.log(
        "Test passed: Non-owner could not edit the message. Error:",
        error.message,
      );
    }

    // Case: Get all messages in the room and verify the count
    console.log("Retrieving all messages in the room...");
    const allMessagesResp = await getAllMessages(users[0].email, roomId);
    if (!allMessagesResp || !Array.isArray(allMessagesResp.data.messages))
      throw new Error("Fetching messages failed");
    console.log(
      `Total messages fetched: ${allMessagesResp.data.messages.length}`,
    );
  } catch (error) {
    console.error("Test case failed with error:", error);
  }
}

runTestCases()
  .then(() => {
    console.log("Test suite execution completed.");
  })
  .catch((error) => {
    console.error("An error occurred during the test suite execution:", error);
  });
