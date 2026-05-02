import { useOutletContext } from 'react-router-dom'
import LiveDataPanel from '../components/LiveDataPanel'
import type { LiveTransactionsOutletContext } from '../hooks/useLiveTransactions'

export default function LiveDataPage() {
  const ctx = useOutletContext<LiveTransactionsOutletContext>()
  return <LiveDataPanel {...ctx} />
}
