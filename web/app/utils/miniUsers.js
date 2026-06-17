export function normalizeUsername(value) {
  return String(value || '')
    .trim()
    .replace(/^@/, '')
    .toLowerCase();
}

export function displayUserName(user) {
  return user?.name || user?.username || userInitials(user);
}

export function userInitials(userOrValue) {
  if (typeof userOrValue === 'object' && userOrValue !== null) {
    return (
      userOrValue.initials ||
      initialsFromText(userOrValue.name || userOrValue.username)
    );
  }
  return initialsFromText(userOrValue);
}

function initialsFromText(value) {
  const text = String(value || '').trim();
  if (!text) return '?';

  const parts = text.split(/\s+/).filter(Boolean);
  if (parts.length === 1) return parts[0].slice(0, 1).toUpperCase();
  return `${parts[0].slice(0, 1)}${parts[1].slice(0, 1)}`.toUpperCase();
}
