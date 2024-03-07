const WebSocket = require("ws");
const axios = require("axios");

function login({
  email,
  password,
  loginUrl = "https://go.paxintrade.com/api/auth/login",
  wsUrl = "wss://go.paxintrade.com/socket.io/?EIO=4&transport=websocket",
}) {
  return new Promise((resolve, reject) => {
    const websocket = new WebSocket(wsUrl);

    websocket.on("open", () => {
      console.log("Connected to the WebSocket server.");
    });

    websocket.on("message", (data) => {
      try {
        const parsedData = JSON.parse(data);

        if (parsedData?.session) {
          console.log(parsedData);
          loginUser(loginUrl, email, password, parsedData.session, () =>
            websocket.close(),
          )
            .then(resolve)
            .catch(reject);
        }
      } catch (error) {
        reject(error);
      }
    });

    websocket.on("error", (error) => {
      reject(error);
    });
  });
}

async function loginUser(loginUrl, email, password, session, closeWebSocket) {
  const payload = {
    email: email,
    password: password,
  };

  const response = await axios.post(loginUrl, payload, {
    headers: {
      Session: session,
    },
  });

  const cookie = response.headers["set-cookie"];
  const cookieString = Array.isArray(cookie) ? cookie.join("; ") : cookie;

  return {
    authInfo: response.data,
    token: response.data.access_token,
    cookie: cookieString,
    session: session,
    closeWebSocket,
  };
}

module.exports = {
  login,
};
