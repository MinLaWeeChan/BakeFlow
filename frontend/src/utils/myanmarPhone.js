export function sanitizeMyanmarPhoneInput(raw) {
  const v = String(raw ?? '').trim();
  const hasPlus = v.startsWith('+');
  const digits = v.replace(/\D/g, '');
  if (!digits) return hasPlus ? '+' : '';
  return hasPlus ? `+${digits}` : digits;
}

export function normalizeMyanmarPhoneE164(raw) {
  const s = sanitizeMyanmarPhoneInput(raw);
  if (!s) return null;
  if (/^\+959\d{9}$/.test(s)) return s;
  if (/^09\d{9}$/.test(s)) return `+959${s.slice(2)}`;
  return null;
}

export function isValidMyanmarPhone(raw) {
  return normalizeMyanmarPhoneE164(raw) !== null;
}

