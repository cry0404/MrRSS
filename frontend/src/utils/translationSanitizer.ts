// stripThinkingContent removes model thinking/reasoning blocks that can
// occasionally leak into translation outputs.
export function stripThinkingContent(text: string): string {
  if (!text) return '';

  let cleaned = text;

  // Remove <thinking>...</thinking> and <think>...</think> blocks.
  cleaned = cleaned.replace(/<\s*(thinking|think)\b[^>]*>[\s\S]*?<\s*\/\s*\1\s*>/gi, '');

  // Normalize whitespace after block removal.
  cleaned = cleaned.trim();

  return cleaned;
}
