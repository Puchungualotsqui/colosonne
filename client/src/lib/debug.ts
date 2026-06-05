export const DEBUG_FRONTIERS = true;

export function debugLog(scope: string, data?: unknown) {
  if (!DEBUG_FRONTIERS) return;

  if (data === undefined) {
    console.log(`[Frontiers:${scope}]`);
    return;
  }

  console.log(`[Frontiers:${scope}]`, data);
}
