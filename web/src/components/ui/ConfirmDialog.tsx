import { ReactNode, useEffect } from 'react'
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

  // Enter confirms non-danger dialogs; danger actions still require a click.
  useEffect(() => {
    if (isDanger) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Enter') return
      e.preventDefault()
      if (onConfirm) onConfirm()
      onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isDanger, onConfirm, onClose])

  return (
    <Modal
      title={title}
      onClose={onClose}
      width={420}
      footer={
        <>
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
        </>
      }
    >
      <div style={{
        fontSize: 15,
        lineHeight: 1.5,
        color: 'var(--color-text-secondary)',
        fontWeight: 500,
      }}>
        {message}
      </div>
    </Modal>
  )
}
