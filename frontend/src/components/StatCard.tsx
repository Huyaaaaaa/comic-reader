import { ReactNode } from 'react'

type Props = {
  label: string
  value: string
  hint?: string
  accent?: string
  icon?: ReactNode
}

export default function StatCard({ label, value, hint, accent = '#1c7c54', icon }: Props) {
  return (
    <article className="stat-card" style={{ ['--accent' as string]: accent }}>
      <div className="stat-head">
        <span>{label}</span>
        <span>{icon}</span>
      </div>
      <strong>{value}</strong>
      {hint ? <p>{hint}</p> : null}
    </article>
  )
}
