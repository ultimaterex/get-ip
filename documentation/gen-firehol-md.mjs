import fs from "node:fs";

const names = fs
  .readFileSync(new URL("./firehol-filenames.txt", import.meta.url), "utf8")
  .trim()
  .split(/\r?\n/)
  .filter(Boolean)
  .sort();
const base = "https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/";

const out = `# FireHOL blocklist-ipsets — HTTP feed URLs (\`BLOCKLIST_URLS\`)

This inventory lists **${names.length}** feed files in the [firehol/blocklist-ipsets](https://github.com/firehol/blocklist-ipsets) repo \`master\` branch (snapshot used for this doc). Files change upstream — confirm names on GitHub before relying on them.

**Pattern:**

\`\`\`text
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/<filename>
\`\`\`

**Official Spamhaus text feeds** (not in the FireHOL tree — often used for DROP):

- \`https://www.spamhaus.org/drop/drop.txt\`
- \`https://www.spamhaus.org/drop/edrop.txt\`

**Caution:** Many feeds overlap; some are ISP/game/ad oriented or extremely large. See [blocklist-examples.md](./blocklist-examples.md). get-ip caps each URL download at **64 MiB**.

---

## Full URL list (one per line)

\`\`\`text
${names.map((f) => base + f).join("\n")}
\`\`\`

---

## Regenerating

1. Refresh \`documentation/firehol-filenames.txt\` from the [GitHub API directory listing](https://api.github.com/repos/firehol/blocklist-ipsets/contents?ref=master) (keep only \`*.netset\` / \`*.ipset\` names).
2. Run: \`node documentation/gen-firehol-md.mjs\`
`;

fs.writeFileSync(new URL("./blocklist-firehol-urls.md", import.meta.url), out);
