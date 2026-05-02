function envOr(key: string, fallback: string): string {
  const raw = (import.meta.env as Record<string, string | undefined>)[key]
  return typeof raw === 'string' && raw.trim() !== '' ? raw.trim() : fallback
}

export const apiSseLiveTransactionsUrl = envOr('VITE_SSE_URL', '/stream/transactions')

export const apiUsersCollectionUrl = envOr('VITE_USERS_URL', '/users')

export function apiUserMemberUrl(userId: string): string {
  return `${apiUsersCollectionUrl}/${encodeURIComponent(userId)}`
}

export const apiUserTransactionsBaseUrl = envOr('VITE_USER_TRANSACTIONS_URL', '/transactions/user')

export function apiUserTransactionListUrl(
  userId: string,
  query: string,
  baseUrl: string = apiUserTransactionsBaseUrl,
): string {
  const q = query.startsWith('?') ? query.slice(1) : query
  const base = baseUrl.replace(/\/$/, '')
  return `${base}/${encodeURIComponent(userId)}${q ? `?${q}` : ''}`
}

export function apiUserTransactionSummaryUrl(
  userId: string,
  query: string,
  baseUrl: string = apiUserTransactionsBaseUrl,
): string {
  const q = query.startsWith('?') ? query.slice(1) : query
  const base = baseUrl.replace(/\/$/, '')
  return `${base}/${encodeURIComponent(userId)}/summary${q ? `?${q}` : ''}`
}
