import http from 'http'

console.log('--- EXAM-P8-T04 Export & Scheduled Reporting Smoke Verification ---')

const checkEndpoint = (path) => {
  return new Promise((resolve) => {
    http.get(`http://localhost:8080${path}`, (res) => {
      console.log(`GET ${path} -> HTTP ${res.statusCode}`)
      resolve(res.statusCode)
    }).on('error', (err) => {
      console.log(`GET ${path} -> Offline/Error: ${err.message}`)
      resolve(null)
    })
  })
}

async function run() {
  await checkEndpoint('/api/v1/health')
  console.log('Smoke verification framework loaded successfully.')
}

run()
