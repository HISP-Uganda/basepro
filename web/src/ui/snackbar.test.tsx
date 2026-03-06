import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { SnackbarProvider, useNotify } from './snackbar'

function NotifyFixture() {
  const notify = useNotify()
  return (
    <button
      type="button"
      onClick={() =>
        notify.success('Saved successfully.', {
          requestId: 'req-web-1',
          title: 'Success',
        })
      }
    >
      Trigger
    </button>
  )
}

describe('notification facade', () => {
  it('renders app notification payload via snackbar provider', async () => {
    render(
      <SnackbarProvider>
        <NotifyFixture />
      </SnackbarProvider>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Trigger' }))
    expect(await screen.findByText('Success')).toBeInTheDocument()
    expect(await screen.findByText('Saved successfully. Request ID: req-web-1')).toBeInTheDocument()
  })
})
