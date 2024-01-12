module.exports = {
    apps: [
        {
            name: 'paxintrade-api',
            exec_mode: 'cluster',
            instances: 'max', // Or a number of instances
            autorestart: true,
            script: '/app/main',
        }
    ]
}