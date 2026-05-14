export default function ContextBadge({ label, value, tone = 'default' }) {
  return (
    <div className={`context-badge context-badge--${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}
