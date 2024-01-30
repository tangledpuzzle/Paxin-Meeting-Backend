module.exports = {
  apps: [
    {
      name: "paxintrade-api",
      exec_mode: "cluster",
      instances: 2, // 'max' or a number of instances
      autorestart: true,
      script: "/app/paxintrade-api",
    },
  ],
};
