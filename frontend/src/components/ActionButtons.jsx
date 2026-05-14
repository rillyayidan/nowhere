export default function ActionButtons({ decision, loading, onAccept, onReject }) {
  return (
    <div className="actions">
      <button className="button button--ghost" type="button" disabled={!decision || loading} onClick={onReject}>
        Another
      </button>
      <button className="button button--primary" type="button" disabled={!decision || loading} onClick={onAccept}>
        Go
      </button>
    </div>
  )
}
