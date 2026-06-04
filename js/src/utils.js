import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

export async function readResponseText(response, debugSave = false, extension = 'json') {
  const text = await response.text();
  if (debugSave) {
    const filePath = path.join(os.tmpdir(), `teams-${Date.now()}.${extension}`);
    await fs.writeFile(filePath, text, 'utf8');
    console.log(`saved temporary json to ${filePath}`);
  }
  return text;
}

export function sortTeamsByName(teams) {
  return [...teams].sort((a, b) => a.displayName.toLowerCase().localeCompare(b.displayName.toLowerCase()));
}

export function sortChannelsByName(channels) {
  return [...channels].sort((a, b) => a.displayName.toLowerCase().localeCompare(b.displayName.toLowerCase()));
}

export function sortMessagesByTime(messages) {
  return [...messages].sort((a, b) => {
    const aTime = a.composeTime ?? a.composetime;
    const bTime = b.composeTime ?? b.composetime;
    return new Date(aTime).getTime() - new Date(bTime).getTime();
  });
}

export function toByteArray(value) {
  return Uint8Array.from(value);
}
