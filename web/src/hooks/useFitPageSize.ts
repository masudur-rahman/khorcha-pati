import { RefObject, useLayoutEffect, useState } from 'react'

interface Params {
  topRef: RefObject<HTMLElement | null>        // mobile: sticky header (filter tabs)
  cardRef: RefObject<HTMLElement | null>       // mobile: a real rendered card
  rowRef: RefObject<HTMLElement | null>        // desktop: a real rendered table row
  wrapRef: RefObject<HTMLElement | null>       // desktop: the scroll frame (thead + tbody)
  paginationRef: RefObject<HTMLElement | null> // bottom boundary (both)
  gap?: number
  extra?: number
  navId?: string
  min?: number
  max?: number
}

// Rows-per-page that fit the inside space, measured entirely from the DOM once at
// mount (before paint, so no flicker), then kept stable for the session so mobile
// URL-bar show/hide can't flip it.
//
// The inside space is framed by real elements, not guessed constants:
//   mobile:  viewport − pinnedFilter − pagination − nav   (chips scroll away above
//            the sticky filter, so they are correctly excluded)
//   desktop: viewport − firstRowTop − pagination          (firstRowTop already
//            includes the topbar, chips, filter and sticky table header)
// A real rendered card/row supplies the item height. rows = floor(inside / item).
//
// On desktop it also bounds the table frame (wrapRef) to the measured height so the
// sticky header stays put and the page itself doesn't scroll — the body scrolls
// (extra rows) or shows empty space (last page) inside the frame instead.
export function useFitPageSize({
  topRef, cardRef, rowRef, wrapRef, paginationRef,
  gap = 8, extra = 8, navId = 'khp-mobile-nav', min = 4, max = 20,
}: Params): number {
  const [size, setSize] = useState(8)

  useLayoutEffect(() => {
    const clamp = (n: number) => Math.max(min, Math.min(max, n))
    const isMobile = window.innerWidth < 768
    const pagination = paginationRef.current?.offsetHeight ?? 0

    if (isMobile && topRef.current && cardRef.current) {
      const topBar = topRef.current.offsetHeight
      const cardH = cardRef.current.offsetHeight || 80
      const nav = document.getElementById(navId)?.offsetHeight ?? 0
      const inside = window.innerHeight - topBar - pagination - nav - extra
      setSize(clamp(Math.floor((inside + gap) / (cardH + gap))))
    } else if (rowRef.current && wrapRef.current) {
      const wrapTop = wrapRef.current.getBoundingClientRect().top
      const rowTop = rowRef.current.getBoundingClientRect().top
      const rowH = rowRef.current.offsetHeight || 52
      // Bound the frame so header + pagination stay fixed and the page won't scroll.
      const frameHeight = window.innerHeight - wrapTop - pagination - extra
      wrapRef.current.style.height = `${frameHeight}px`
      wrapRef.current.style.overflowY = 'auto'
      const inside = window.innerHeight - rowTop - pagination - extra
      setSize(clamp(Math.floor(inside / rowH)))
    }
    // Intentionally run once — the count must stay stable for the session.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return size
}
