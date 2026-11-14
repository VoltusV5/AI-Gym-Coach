import { spawnSync } from 'node:child_process'

const npm = process.platform === 'win32' ? 'npm.cmd' : 'npm'
const steps = ['lint', 'test', 'build']
let failed = false
for (const step of steps) {
  const result = spawnSync(npm, ['run', step], { encoding: 'utf8', shell: process.platform === 'win32' })
  const passed = result.status === 0
  console.log(`${passed ? 'PASS' : 'FAIL'} ${step}`)
  if (!passed) {
    failed = true
    console.error(result.error?.message ?? [result.stdout, result.stderr].filter(Boolean).join('\n'))
  }
}
process.exitCode = failed ? 1 : 0
