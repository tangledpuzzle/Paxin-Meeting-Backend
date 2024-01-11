module.exports = {
    apps: [
        {
            name: 'paxintrade-api',
            script: '/app/main',
            exec_mode: 'cluster',
            instances: 'max', // Or a number of instances
        }
    ]
}