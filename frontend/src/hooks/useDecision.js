import { useCallback, useMemo, useState } from 'react'

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api'

export function useDecision(location) {
  const [budget, setBudget] = useState('medium')
  const [mood, setMood] = useState('open')
  const [decision, setDecision] = useState(null)
  const [context, setContext] = useState(null)
  const [source, setSource] = useState('')
  const [rejected, setRejected] = useState([])
  const [feedback, setFeedback] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const userId = useMemo(() => {
    const key = 'nowhere:user-id'
    const existing = localStorage.getItem(key)
    if (existing) return existing
    const created = crypto.randomUUID ? crypto.randomUUID() : `demo-${Date.now()}`
    localStorage.setItem(key, created)
    return created
  }, [])

  const decide = useCallback(
    async (nextRejected = rejected) => {
      setLoading(true)
      setError('')
      try {
        const response = await fetch(`${API_BASE}/decide`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            user_id: userId,
            location,
            budget,
            mood,
            time_of_day: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            rejected: nextRejected,
            last_context_id: context?.id || '',
          }),
        })
        if (!response.ok) throw new Error(await responseError(response, 'Decision failed'))
        const payload = await response.json()
        setContext(payload.context)
        setDecision(payload.decision)
        setSource(payload.source)
      } catch (err) {
        setError(err.message || 'Unable to get a recommendation.')
      } finally {
        setLoading(false)
      }
    },
    [budget, context?.id, location, mood, rejected, userId],
  )

  const sendFeedback = useCallback(
    async (action) => {
      if (!decision || !context) return
      const nextRejected = action === 'reject'
        ? [...rejected, decision.place.id || decision.place.name]
        : rejected

      setFeedback(null)
      try {
        const response = await fetch(`${API_BASE}/feedback`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            user_id: userId,
            context_id: context.id,
            decision_id: decision.id,
            action,
            place: decision.place,
            rejected: nextRejected,
          }),
        })
        if (!response.ok) throw new Error(await responseError(response, 'Feedback failed'))
        const payload = await response.json()
        setFeedback(payload)
        if (action === 'reject') {
          setRejected(nextRejected)
          await decide(nextRejected)
        }
      } catch (err) {
        setError(err.message || 'Unable to store feedback.')
      }
    },
    [context, decide, decision, rejected, userId],
  )

  const resetRejected = useCallback(() => {
    setRejected([])
    setFeedback(null)
  }, [])

  return {
    budget,
    setBudget,
    mood,
    setMood,
    context,
    decision,
    source,
    rejected,
    feedback,
    loading,
    error,
    decide,
    sendFeedback,
    resetRejected,
  }
}

async function responseError(response, fallback) {
  const text = await response.text()
  if (!text) return `${fallback} with ${response.status}`

  try {
    const payload = JSON.parse(text)
    return payload.error || payload.message || `${fallback} with ${response.status}`
  } catch {
    return text.slice(0, 240)
  }
}
