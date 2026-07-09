import { ReactNode } from 'react'
import Modal from './Modal'
import Button from './Button'

interface ConfirmDialogProps {
  title: string
  message: ReactNode
  confirmText?: string
  cancelText?: string
  type?: 'danger' | 'warning' | 'info'
  mode?: 'confirm' | 'alert'
  onConfirm?: () => void
  onClose: () => void
}

export default function ConfirmDialog({
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  type = 'info',
  mode = 'confirm',
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  const isDanger = type === 'danger'

  return (
    <Modal title={title} onClose={onClose} width={420}>
      <div style={{ padding: '24px 32px', display: 'flex', flexDirection: 'column', gap: 16 }}>
        <p style={{ 
          fontSize: 15, 
          lineHeight: '1.5', 
          color: 'var(--color-text-secondary)',
          margin: 0,
          fontWeight: 500,
        }}>
          {message}
        </p>
        
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 8 }}>
          {mode === 'confirm' && (
            <Button variant="secondary" onClick={onClose}>
              {cancelText}
            </Button>
          )}
          <Button 
            variant={isDanger ? 'danger' : 'primary'}
            onClick={() => {
              if (onConfirm) onConfirm()
              onClose()
            }}
            style={{ padding: '10px 24px' }}
          >
            {mode === 'alert' ? 'OK' : confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
