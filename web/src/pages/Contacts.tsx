import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'

export default function Contacts() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Contacts" subtitle="People you transact with" />
      <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
        <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
          Contacts page coming soon. Currently managed under Wallets.
        </p>
      </Card>
    </div>
  )
}
