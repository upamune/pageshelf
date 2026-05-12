# Variant 02 — Mode segmented panel

## Stance
A segmented panel separates intent before showing controls: `Annotate`, `Review`, and `Export`. Selection state is only represented inside `Annotate`.

## Interaction model
- `Annotate`: target acquisition and note composition only.
- `Review`: existing notes with edit/delete/resolve only.
- `Export`: Copy annotations, Copy JSON, Export JSON, and Clear resolved only.
- Toggling selection returns the user to `Annotate`, making selection state explicit.

## Tradeoffs
- Pros: strongest conceptual separation; export actions cannot be confused with a selected quote.
- Pros: scales well if annotation review gains filters later.
- Cons: extra click to move between composing and reviewing.
- Cons: tabs can feel heavier for a tiny embedded runtime.

## Runtime fit
Maps cleanly to Shadow DOM panel and mobile bottom sheet. All data remains local; no network behavior is implied. Supports selection anchors, element picking, note lifecycle actions, and JSON/markdown copy/export.
