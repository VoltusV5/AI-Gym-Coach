const fs = require('node:fs')
const path = require('node:path')
const [modules, fixturePath, reportPath] = process.argv.slice(2)
const newman = require(path.resolve(modules, 'newman'))
const fixture = JSON.parse(fs.readFileSync(fixturePath, 'utf8'))
const endpoint = new URL(fixture.base_url)
if (endpoint.hostname !== '127.0.0.1' || !fixture.admin_token) throw new Error('An isolated local fixture is required')
const environment = JSON.parse(fs.readFileSync(path.join(__dirname, 'Local.postman_environment.json'), 'utf8'))
const overrides = { base_url: fixture.base_url, admin_token: fixture.admin_token, test_environment: 'true', max_stage: '15', verification_document_ref: 'https://example.test/synthetic-certificate' }
for (const [key, value] of Object.entries(overrides)) {
  const item = environment.values.find((entry) => entry.key === key)
  if (item) item.value = value
  else environment.values.push({ key, value, enabled: true })
}
newman.run({ collection: path.join(__dirname, 'Sportik.postman_collection.json'), environment, reporters: [], timeoutRequest: 15000 }, (error, summary) => {
  if (error) { process.stderr.write('Collection execution failed\n'); process.exitCode = 1; return }
  const failures = summary.run.failures.map((failure) => ({ name: failure.source?.name || failure.at, message: String(failure.error?.message || '').replace(/[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+/g, '[REDACTED]') }))
  const report = { isolated: true, requests: summary.run.stats.requests, assertions: summary.run.stats.assertions, failures }
  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2) + '\n')
  process.stdout.write(JSON.stringify({ requests: report.requests, assertions: report.assertions, failures: failures.length }) + '\n')
  if (failures.length) process.exitCode = 1
})
