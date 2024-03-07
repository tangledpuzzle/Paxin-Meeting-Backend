const axios = require("axios");

const requestHelper = async ({ token, session, url, method, data }) => {
  try {
    // Prepare the headers
    const headers = {
      Authorization: `Bearer ${token}`,
    };

    // If a session is provided, include it in the headers
    if (session) {
      headers.Session = session;
    }

    // Set the content type to JSON for POST, PUT methods
    if (["POST", "PATCH"].includes(method.toUpperCase())) {
      headers["Content-Type"] = "application/json";
    }

    // Prepare request configuration
    const config = {
      method,
      url,
      headers,
    };

    // Add request body for methods that require it
    if (["POST", "PATCH"].includes(method.toUpperCase()) && data) {
      config.data = data;
    }

    // Send the request using axios and return the response data
    const response = await axios(config);

    // Return the necessary data from the response
    return response.data;
  } catch (error) {
    console.error(
      "Error in request:",
      error.response
        ? error.response.error || error.response.data
        : error.message,
    );
    if (error.response?.data) return error.response.data;
    throw error;
  }
};

module.exports = {
  requestHelper,
};
