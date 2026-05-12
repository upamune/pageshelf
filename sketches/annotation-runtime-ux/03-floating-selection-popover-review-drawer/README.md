# Variant 03 — Floating selection popover + separate Review drawer

## Stance
Selection annotation lives near the selected text as a transient popover. The review drawer is reserved for existing notes and global export/copy actions.

## Interaction model
- Selection opens a floating composer near the quote.
- Adding a note dismisses the popover and can reveal the review drawer.
- Drawer contains only durable review tasks: edit, resolve/reopen, delete, copy/export, clear resolved.
- Mobile adapts the popover into a compact bottom composer above the drawer/FAB.

## Tradeoffs
- Pros: cleanest separation between ephemeral selection and persistent review.
- Pros: keeps annotation action spatially tied to the selected text anchor.
- Cons: positioning a Shadow DOM popover near selected ranges is more complex than rendering inside one panel.
- Cons: needs careful collision handling with page edges, iframes, zoom, and mobile keyboards.

## Runtime fit
Still compatible with localStorage-only storage and text/element anchors. It would require robust range rectangle measurement for the popover, while the drawer remains a standard Shadow DOM panel/bottom sheet.
