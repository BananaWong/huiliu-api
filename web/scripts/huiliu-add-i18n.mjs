// huiliu: append translation keys to every locale without re-serializing the
// files (keeps upstream escapes and ordering untouched). Usage:
//   node scripts/huiliu-add-i18n.mjs <translations.json>
// translations.json: { "English key": { "zh": "...", "zh-TW": "...", ... } }
// The en value is always the key itself; missing locales fall back to English.
import fs from 'node:fs'
import path from 'node:path'

const LOCALES_DIR = path.resolve('src/i18n/locales')
const table = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'))

for (const file of fs
  .readdirSync(LOCALES_DIR)
  .filter((f) => f.endsWith('.json'))) {
  const locale = file.replace(/\.json$/, '')
  const full = path.join(LOCALES_DIR, file)
  const raw = fs.readFileSync(full, 'utf8')
  const existing = JSON.parse(raw).translation
  const lines = []
  for (const [key, values] of Object.entries(table)) {
    if (Object.prototype.hasOwnProperty.call(existing, key)) continue
    const value = locale === 'en' ? key : (values[locale] ?? key)
    lines.push(`    ${JSON.stringify(key)}: ${JSON.stringify(value)}`)
  }
  if (!lines.length) continue
  const closing = raw.lastIndexOf('\n  }\n}')
  if (closing < 0) throw new Error(`unexpected layout in ${file}`)
  const next = `${raw.slice(0, closing)},\n${lines.join(',\n')}${raw.slice(closing)}`
  JSON.parse(next)
  fs.writeFileSync(full, next)
  console.log(`${locale}: +${lines.length}`)
}
