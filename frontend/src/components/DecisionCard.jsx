export default function DecisionCard({ decision, source }) {
  if (!decision) {
    return (
      <section className="decision-card decision-card--empty">
        <p className="eyebrow">Waiting for context</p>
        <h2>Tap decide to get one clear place.</h2>
      </section>
    )
  }

  const { place } = decision
  return (
    <section className="decision-card">
      <div className="decision-card__topline">
        <span>{place.category}</span>
        <span>{source}</span>
      </div>
      <h2>{place.name}</h2>
      <p>{decision.reason}</p>
      <dl className="decision-meta">
        <div>
          <dt>Duration</dt>
          <dd>{decision.duration_minutes} min</dd>
        </div>
        <div>
          <dt>Rating</dt>
          <dd>{place.rating ? place.rating.toFixed(1) : 'N/A'}</dd>
        </div>
        <div>
          <dt>Status</dt>
          <dd>{place.open_now ? 'Open' : 'Check hours'}</dd>
        </div>
      </dl>
      <p className="address">{place.address}</p>
    </section>
  )
}
