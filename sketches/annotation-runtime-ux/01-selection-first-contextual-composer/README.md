# Variant 01 — Selection-first contextual composer

## Stance
When text is selected, Pageshelf promotes a compact composer at the top of the runtime panel. When there is no selection, the same surface becomes a review drawer with notes and global actions.

## Interaction model
- `Selected text` card appears only during selection state.
- `Add selection note` and `Pick element instead` are target-specific actions.
- Existing note management and copy/export actions stay below the contextual card.
- Mobile collapses into a bottom sheet; desktop sits as a right-side embedded panel/FAB-style surface.

## Tradeoffs
- Pros: very clear state switch; low implementation risk in Shadow DOM; uses one panel surface.
- Pros: text quote is visible before writing, reducing anchor ambiguity.
- Cons: composer can push existing notes down when selection exists.
- Cons: element picking is still near selection composition, so labeling must make its target-specific nature explicit.

## Runtime fit
Works with localStorage-only annotations, selection anchors, element anchors, edit/delete/resolve, Copy annotations, Copy JSON, and Export JSON. No network assumptions.
