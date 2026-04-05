/** Returns the uppercased first non-whitespace character of a group name, or '?' if empty. */
export function groupInitial(name: string): string {
  const trimmed = name.trimStart()
  return trimmed.length > 0 ? trimmed[0].toUpperCase() : '?'
}
