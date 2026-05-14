import { useEffect } from 'react'
import ActionButtons from './components/ActionButtons.jsx'
import ContextBadge from './components/ContextBadge.jsx'
import DecisionCard from './components/DecisionCard.jsx'
import { useDecision } from './hooks/useDecision.js'
import { useLocation } from './hooks/useLocation.js'

export default function App() {
  const { location, status, error: locationError, requestLocation } = useLocation()
  const decisionState = useDecision(location)

  useEffect(() => {
    requestLocation()
  }, [requestLocation])

  const weather = decisionState.context?.weather
  const rejectedCount = decisionState.rejected.length

  return (
    <main className="app-shell">
      <section className="workspace">
        <header className="app-header">
          <div>
            <p className="eyebrow">NowHere</p>
            <h1>One good next place.</h1>
          </div>
          <button className="icon-button" type="button" aria-label="Refresh location" onClick={requestLocation}>
            +
          </button>
        </header>

        <div className="context-grid">
          <ContextBadge label="GPS" value={status === 'loading' ? 'Locating' : status} tone={status} />
          <ContextBadge label="Weather" value={weather ? `${Math.round(weather.temperature_c)}C ${weather.summary}` : 'Ready'} />
          <ContextBadge label="Rejected" value={rejectedCount} tone={rejectedCount ? 'warn' : 'default'} />
        </div>

        <section className="control-panel">
          <div className="field-group" aria-label="Budget">
            <span>Budget</span>
            {['low', 'medium', 'high'].map((item) => (
              <button
                key={item}
                type="button"
                className={decisionState.budget === item ? 'segmented active' : 'segmented'}
                onClick={() => decisionState.setBudget(item)}
              >
                {item}
              </button>
            ))}
          </div>
          <label className="select-field">
            <span>Mood</span>
            <select value={decisionState.mood} onChange={(event) => decisionState.setMood(event.target.value)}>
              <option value="open">Open</option>
              <option value="quiet">Quiet</option>
              <option value="work">Work</option>
              <option value="food">Food</option>
              <option value="walk">Walk</option>
            </select>
          </label>
        </section>

        <DecisionCard decision={decisionState.decision} source={decisionState.source} />

        <div className="command-bar">
          <button className="button button--primary" type="button" disabled={decisionState.loading} onClick={() => decisionState.decide()}>
            {decisionState.loading ? 'Deciding' : 'Decide'}
          </button>
          <ActionButtons
            decision={decisionState.decision}
            loading={decisionState.loading}
            onAccept={() => decisionState.sendFeedback('accept')}
            onReject={() => decisionState.sendFeedback('reject')}
          />
        </div>

        {(decisionState.error || locationError || decisionState.feedback) && (
          <aside className="status-strip">
            {decisionState.error || locationError || decisionState.feedback?.preference_hint}
          </aside>
        )}

        {decisionState.decision?.place?.map_url && (
          <a className="map-link" href={decisionState.decision.place.map_url} target="_blank" rel="noreferrer">
            Open in Google Maps
          </a>
        )}
      </section>
    </main>
  )
}
