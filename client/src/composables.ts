'use strict'

// Output logging information to console.
export function log(level: 'INFO' | 'ERROR', text: string) {
  const time = new Date()
  const msg = '[' + time.toLocaleTimeString() + '] ' + text
  switch (level) {
    case 'INFO':
      console.log(msg)
      break
    case 'ERROR':
      console.trace(msg)
  }
}
