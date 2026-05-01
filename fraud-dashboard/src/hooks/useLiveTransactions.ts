import { useEffect, useMemo, useState } from 'react'

export type LiveTransaction = {
  id: string
  userid: string
  userName: string
  userSurname: string
  amount: number
  location: string
  transactionStatus: string
  fraudstatus: boolean
  createdat: string
}

type ConnectionMode = 'idle' | 'sse' | 'disconnected'

type UseLiveTransactionsOptions = {
  sseUrl?: string
  maxItems?: number
}

type BackendTransaction = {
  id: string
  user_id: string
  user_name?: string
  user_surname?: string
  amount: number
  location: string
  transaction_status: string
  fraud_status: boolean
  created_at: string
}

type BackendEvaluatedTransaction = {
  transaction_id: string
  user_id: string
  user_name: string
  user_surname: string
  amount: number
  location: string
  transaction_status: string
  fraud_status: boolean
  created_at: string
}

function mapBackendTransaction(tx: BackendTransaction): LiveTransaction {
  return {
    id: tx.id,
    userid: tx.user_id,
    userName: tx.user_name ?? '',
    userSurname: tx.user_surname ?? '',
    amount: tx.amount,
    location: tx.location,
    transactionStatus: tx.transaction_status,
    fraudstatus: tx.fraud_status,
    createdat: tx.created_at,
  }
}

function mapBackendEvaluatedTransaction(tx: BackendEvaluatedTransaction): LiveTransaction {
  return {
    id: tx.transaction_id,
    userid: tx.user_id,
    userName: tx.user_name,
    userSurname: tx.user_surname,
    amount: tx.amount,
    location: tx.location,
    transactionStatus: tx.transaction_status,
    fraudstatus: tx.fraud_status,
    createdat: tx.created_at,
  }
}

function isBackendTransaction(value: unknown): value is BackendTransaction {
  if (!value || typeof value !== 'object') return false
  const v = value as Record<string, unknown>
  return (
    typeof v.id === 'string' &&
    typeof v.user_id === 'string' &&
    (typeof v.user_name === 'string' || typeof v.user_name === 'undefined') &&
    (typeof v.user_surname === 'string' || typeof v.user_surname === 'undefined') &&
    typeof v.amount === 'number' &&
    typeof v.location === 'string' &&
    typeof v.transaction_status === 'string' &&
    typeof v.fraud_status === 'boolean' &&
    typeof v.created_at === 'string'
  )
}

function isBackendEvaluatedTransaction(value: unknown): value is BackendEvaluatedTransaction {
  if (!value || typeof value !== 'object') return false
  const v = value as Record<string, unknown>
  return (
    typeof v.transaction_id === 'string' &&
    typeof v.user_id === 'string' &&
    typeof v.user_name === 'string' &&
    typeof v.user_surname === 'string' &&
    typeof v.amount === 'number' &&
    typeof v.location === 'string' &&
    typeof v.transaction_status === 'string' &&
    typeof v.fraud_status === 'boolean' &&
    typeof v.created_at === 'string'
  )
}

function isLiveTransaction(value: unknown): value is LiveTransaction {
  if (!value || typeof value !== 'object') return false
  const v = value as Record<string, unknown>
  return (
    typeof v.id === 'string' &&
    typeof v.userid === 'string' &&
    typeof v.userName === 'string' &&
    typeof v.userSurname === 'string' &&
    typeof v.amount === 'number' &&
    typeof v.location === 'string' &&
    typeof v.transactionStatus === 'string' &&
    typeof v.fraudstatus === 'boolean' &&
    typeof v.createdat === 'string'
  )
}

function normalizePayload(input: unknown): LiveTransaction[] {
  const list = Array.isArray(input) ? input : [input]

  const mapped: LiveTransaction[] = []
  for (const item of list) {
    if (isLiveTransaction(item)) {
      mapped.push(item)
      continue
    }
    if (isBackendTransaction(item)) {
      mapped.push(mapBackendTransaction(item))
      continue
    }
    if (isBackendEvaluatedTransaction(item)) {
      mapped.push(mapBackendEvaluatedTransaction(item))
    }
  }

  return mapped
}

function prependUnique(
  incoming: LiveTransaction[],
  previous: LiveTransaction[],
  maxItems: number,
): LiveTransaction[] {
  const next = [...incoming, ...previous]
  const seen = new Set<string>()
  const deduped: LiveTransaction[] = []
  for (const tx of next) {
    if (seen.has(tx.id)) continue
    seen.add(tx.id)
    deduped.push(tx)
    if (deduped.length >= maxItems) break
  }
  return deduped
}

export function useLiveTransactions({
  sseUrl,
  maxItems = 100,
}: UseLiveTransactionsOptions) {
  const [transactions, setTransactions] = useState<LiveTransaction[]>([])
  const [connectionMode, setConnectionMode] = useState<ConnectionMode>('idle')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let source: EventSource | null = null
    let unmounted = false

    if (sseUrl) {
      try {
        source = new EventSource(sseUrl)

        const handlePayload = (raw: string) => {
          const parsed = JSON.parse(raw) as unknown
          const payload =
            parsed &&
            typeof parsed === 'object' &&
            'data' in (parsed as Record<string, unknown>)
              ? (parsed as Record<string, unknown>).data
              : parsed
          const normalized = normalizePayload(payload)
          if (normalized.length > 0) {
            setTransactions((prev) => prependUnique(normalized, prev, maxItems))
          }
        }

        source.onopen = () => {
          if (unmounted) return
          setConnectionMode('sse')
          setIsLoading(false)
          setError(null)
        }
        source.onmessage = (event) => {
          if (unmounted) return
          try {
            handlePayload(event.data)
          } catch {
            setError('sse message parse failed')
          }
        }
        source.addEventListener('transaction_evaluated', (event) => {
          if (unmounted) return
          const message = event as MessageEvent
          try {
            handlePayload(message.data)
          } catch {
            setError('sse message parse failed')
          }
        })
        source.addEventListener('ready', () => {
          if (unmounted) return
          setConnectionMode('sse')
          setIsLoading(false)
          setError(null)
        })
        source.onerror = () => {
          if (unmounted) return
          source?.close()
          setConnectionMode('disconnected')
          setError('sse connection lost')
          setIsLoading(false)
        }
      } catch {
        setConnectionMode('disconnected')
        setError('sse initialization failed')
        setIsLoading(false)
      }
    } else {
      setConnectionMode('disconnected')
      setError('sse url is not configured')
      setIsLoading(false)
    }

    return () => {
      unmounted = true
      source?.close()
    }
  }, [sseUrl, maxItems])

  const sortedTransactions = useMemo(
    () =>
      [...transactions].sort((a, b) => {
        const at = new Date(a.createdat).getTime()
        const bt = new Date(b.createdat).getTime()
        return bt - at
      }),
    [transactions],
  )

  return {
    transactions: sortedTransactions,
    connectionMode,
    isLoading,
    error,
  }
}
