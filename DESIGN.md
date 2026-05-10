---
version: alpha
name: Pageshelf
description: A calm, kawaii, soft-3D developer-tool identity built around a sleepy shelf-keeper mascot, organized artifact shelves, private sharing, and local-first security.
colors:
  primary: "#1D2356"
  primary-soft: "#3B4F9C"
  secondary: "#7EA1E6"
  tertiary: "#B9A6F5"
  accent: "#F59E3D"
  mint: "#8FD3B6"
  amber: "#FFC766"
  surface: "#FFF7ED"
  surface-raised: "#FFFBF4"
  surface-tint: "#F6F0E7"
  on-surface: "#1D2356"
  on-surface-muted: "#66709A"
  border: "#E9DCCB"
  shelf-wood: "#E2C39A"
  shelf-wood-dark: "#B98245"
  document: "#FFF1DC"
  code-orange: "#F59E3D"
  lock-gold: "#D89437"
  success: "#55B98B"
  warning: "#FFC766"
  error: "#E75D54"
typography:
  headline-display:
    fontFamily: Inter
    fontSize: 56px
    fontWeight: 800
    lineHeight: 1.05
    letterSpacing: "-0.04em"
  headline-lg:
    fontFamily: Inter
    fontSize: 40px
    fontWeight: 800
    lineHeight: 1.1
    letterSpacing: "-0.035em"
  headline-md:
    fontFamily: Inter
    fontSize: 28px
    fontWeight: 750
    lineHeight: 1.15
    letterSpacing: "-0.025em"
  title-lg:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: 700
    lineHeight: 1.25
    letterSpacing: "-0.015em"
  title-md:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 700
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 400
    lineHeight: 1.65
    letterSpacing: "-0.005em"
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.6
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.55
  label-lg:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: 0.01em
  label-md:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: 0.04em
  code-md:
    fontFamily: JetBrains Mono
    fontSize: 14px
    fontWeight: 500
    lineHeight: 1.45
rounded:
  none: 0px
  xs: 6px
  sm: 10px
  md: 14px
  lg: 20px
  xl: 28px
  xxl: 36px
  full: 9999px
spacing:
  none: 0px
  xxs: 2px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  xxl: 48px
  xxxl: 64px
  page-max-width: 1120px
  content-max-width: 760px
  icon-grid: 8px
components:
  page:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    width: "{spacing.page-max-width}"
  card:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.lg}"
    padding: "{spacing.lg}"
  card-soft:
    backgroundColor: "#FFF4E4"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.xl}"
    padding: "{spacing.xl}"
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "#FFFFFF"
    rounded: "{rounded.full}"
    padding: 12px 18px
    typography: "{typography.label-lg}"
  button-secondary:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.primary}"
    rounded: "{rounded.full}"
    padding: 12px 18px
    typography: "{typography.label-lg}"
  input:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.md}"
    padding: 12px 14px
  chip:
    backgroundColor: "#EEF3FF"
    textColor: "{colors.primary-soft}"
    rounded: "{rounded.full}"
    padding: 6px 10px
  document-card:
    backgroundColor: "{colors.document}"
    textColor: "{colors.primary}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
  browser-card:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.primary}"
    rounded: "{rounded.lg}"
    padding: "{spacing.md}"
  cli-prompt-tile:
    backgroundColor: "#253052"
    textColor: "#BFD4FF"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
    typography: "{typography.code-md}"
  lock-badge:
    backgroundColor: "#EAF0FF"
    textColor: "{colors.primary-soft}"
    rounded: "{rounded.lg}"
    padding: "{spacing.sm}"
  shelf-tile:
    backgroundColor: "{colors.shelf-wood}"
    textColor: "{colors.primary}"
    rounded: "{rounded.xl}"
    padding: "{spacing.lg}"
  mascot-avatar:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.primary}"
    rounded: "{rounded.xl}"
    padding: "{spacing.md}"
---

# Pageshelf Design System

## Overview

Pageshelf should look like a secure little shelf for AI-generated HTML artifacts: calm, local-first, organized, and quietly competent. The identity is built around a sleepy but reliable kawaii shelf-keeper mascot, a rounded artifact shelf, glowing HTML pages, small browser cards, CLI prompt tiles, and subtle lock/key/security motifs.

The visual style is soft 3D kawaii developer tooling: more like a Japanese capsule toy or vinyl desk collectible than an anime character. It should feel friendly without becoming childish, technical without becoming cold, and secure without becoming enterprise-gray.

The core personality is:

- Calm: reduce cognitive load; avoid loud contrast and visual noise.
- Secure: imply privacy, local-first storage, tokens, and protected sharing through small lock/key/shield motifs.
- Organized: every artifact has a place; layouts should feel shelf-like, gridded, and tidy.
- Agent-friendly: patterns should be clear enough for AI agents to reproduce consistently.

Use the mascot for README hero images, onboarding, empty states, success states, and friendly product moments. Use simpler UI primitives for dense screens where the mascot would become distracting.

## Colors

The palette is warm, pastel, and low-noise. Use the deep indigo as the primary readable color, cream surfaces as the base, soft blues/lavenders for product identity, mint for safe/connected states, and amber/orange for HTML/code-page highlights.

- Primary / Ink Indigo (`#1D2356`): wordmark, headings, primary text, important controls.
- Primary Soft (`#3B4F9C`): secondary headings, icon strokes, active states.
- Sky Blue (`#7EA1E6`): mascot hood, primary brand surfaces, selected states.
- Lavender (`#B9A6F5`): secondary brand accent, dotted private-network arcs, soft UI blocks.
- Mint (`#8FD3B6`): safe connection, local/private status, key-tag details.
- Amber Glow (`#FFC766`): warm highlights, glowing document edges, success warmth.
- Code Orange (`#F59E3D`): `</>` marks, CLI affordance accents, small code details.
- Cloud Cream (`#FFF7ED`): page background and hero backdrop.
- Paper Cream (`#FFFBF4`): raised cards, document bodies, modal surfaces.
- Shelf Wood (`#E2C39A`): shelf surfaces and mascot-scene structure.
- Border Beige (`#E9DCCB`): card outlines, separators, low-emphasis boundaries.

Rules:

- Prefer cream and off-white surfaces over pure white.
- Use indigo for readability and brand weight.
- Use orange sparingly. It is for code symbols and artifact highlights, not broad fills.
- Avoid harsh neon, saturated cyberpunk colors, pure black, and aggressive red unless showing destructive/error states.
- Do not let pastel colors reduce text contrast. Text on pastel fills should usually be indigo.

## Typography

Use a rounded, modern sans-serif that feels readable in both OSS documentation and product UI. Default to Inter for most UI. Use JetBrains Mono only for command snippets, terminal cards, file paths, ports, tokens, and code-adjacent metadata.

- Wordmark / hero display: heavy rounded sans-serif, dark indigo, large, friendly, with tight letter spacing.
- Headlines: Inter, 700–800 weight, tight tracking, short lines.
- Body: Inter, 14–18px, regular weight, generous line height.
- Labels: Inter, 12–14px, bold or semi-bold, used for cards, chips, statuses, and section headers.
- Code / CLI: JetBrains Mono, 13–14px, medium weight, never overused.

Typography should feel polished but not precious. Avoid decorative fonts, thin weights, overly geometric techno fonts, and handwritten styles. Pageshelf is cute, but its type should remain product-grade.

## Layout

Pageshelf UI should feel like an organized shelf: clear compartments, predictable spacing, and visible hierarchy. Use an 8px spacing rhythm, with generous whitespace around important content.

- Use a max content width of 1120px for docs/product pages.
- Use 760px max width for prose-heavy content.
- Prefer 2–3 column layouts for cards on desktop, collapsing to a single column on mobile.
- Keep hero sections centered or left-symbol/right-wordmark for brand assets.
- In dense UI, use compact rows but preserve soft spacing and clear grouping.
- Align icons, labels, and controls to an 8px grid.

Do not create chaotic collage layouts. The mascot scene can be detailed, but application screens should be calmer and more systematic.

## Elevation & Depth

Depth should come from soft toy-like forms, tonal layering, subtle borders, and gentle shadows. Avoid heavy SaaS drop shadows and glassmorphism.

Use three depth levels:

1. Flat surface: page background, documentation body, low-emphasis areas.
2. Raised card: cards, panels, browser-preview blocks, shelf tiles.
3. Mascot / hero object: soft 3D mascot renders, app icons, README hero art.

Guidelines:

- Use soft shadows with low opacity and wide blur.
- Combine a light border with a subtle shadow for cards.
- Raised cards should look touchable, not floating aggressively.
- 3D assets should have smooth studio lighting, rounded forms, and clean edges.
- Avoid hard black shadows, dramatic perspective, noisy textures, and photorealistic materials.

Suggested CSS shadow styles:

```css
--shadow-sm: 0 1px 2px rgb(29 35 86 / 0.06), 0 1px 1px rgb(29 35 86 / 0.04);
--shadow-md: 0 8px 24px rgb(29 35 86 / 0.10), 0 2px 6px rgb(29 35 86 / 0.06);
--shadow-soft-3d: 0 18px 48px rgb(137 102 61 / 0.16), 0 4px 12px rgb(29 35 86 / 0.08);
```

## Shapes

The shape language is rounded, soft, and shelf-like. Corners should feel intentionally friendly. Sharp rectangles break the identity.

Core shapes:

- Soft rounded rectangles: default panels, cards, inputs, browser windows.
- Pills and capsules: chips, tags, status indicators, small controls.
- Document cards: cream pages with folded corners and orange `</>` symbols.
- Browser window cards: mini metadata previews with top-dot controls and simple content blocks.
- Rounded shelf corners: big, toy-like container shapes for storage metaphors.
- Dotted connection arcs: subtle private-network motif; use sparingly.
- Lock/key/shield badges: small security affordances, never large threatening security theater.

Avoid spikes, angular sci-fi shapes, sharp-edged terminal aesthetics, aggressive security iconography, and inconsistent corner radii in the same view.

## Components

### Buttons

Buttons should be rounded pills with clear hierarchy.

- Primary: indigo fill, white text, used for one main action per screen.
- Secondary: cream/white fill, indigo text, beige border.
- Tertiary / ghost: text-only or soft-tinted fill for low-emphasis actions.
- Use 12px vertical padding and 18px horizontal padding as a default.
- Icons inside buttons should be rounded/simple, not sharp.

### Cards

Cards are the default grouping primitive.

- Use paper cream or soft tinted surfaces.
- Use a 1px beige border and low-opacity soft shadow.
- Keep card content organized: icon/title/body/action.
- Prefer 20px radius for standard cards.
- Avoid dense borders, table-like boxes, and overloaded cards.

### Document Card

Represents a saved HTML artifact.

- Cream document body.
- Optional folded corner.
- Orange `</>` mark or small code glyph.
- 2–3 muted metadata lines.
- Can glow subtly with amber when active, generated, or newly saved.

### Browser Card

Represents a preview or metadata snapshot for a web artifact.

- Rounded card with a muted blue/lavender header bar.
- Three tiny circular controls can appear in the header.
- Body contains abstract blocks/lines, not detailed screenshots unless needed.
- Keep it readable at small sizes.

### CLI Prompt Tile

Represents local CLI commands or agent actions.

- Dark indigo surface.
- Light blue or mint monospace glyphs.
- Use `>_`, `$`, or command-like marks sparingly.
- Should feel friendly and contained, not hacker/noir.

### Shelf / Storage Tile

Represents organized local storage.

- Warm wood or cream container.
- Rounded shelf corners.
- Can contain document cards, browser cards, and small status dots.
- Use for onboarding, empty states, storage views, and docs illustrations.

### Lock Badge / Key Tag / Shield

Represents privacy, tokenized sharing, or protected local access.

- Prefer small badges attached to UI or mascot objects.
- Use mint, blue, or gold accents.
- Avoid large lock walls, warning shields, or fear-based security visuals.

### Mascot

The mascot is a sleepy but competent shelf-keeper.

- Rounded blue hood, cream face/body, tiny dot/oval eyes, minimal mouth.
- Expression: calm, slightly sleepy, reliable.
- Props: document card, CLI prompt card, key tag, lock charm.
- Use as a friendly guide, not as a loud mascot shouting at the user.
- Never make the mascot angry, edgy, overly anime, overly human, or hyperactive.

### App Icon / GitHub Avatar

- Use the mascot head or mascot-with-shelf symbol.
- Keep details minimal at small sizes.
- Prefer a rounded square blue/cream background for app-icon contexts.
- Avoid including the full wordmark in avatar-size icons.

## Do's and Don'ts

### Do

- Do keep the UI calm, pastel, and organized.
- Do use rounded soft forms consistently.
- Do use cream/off-white surfaces instead of sterile white when possible.
- Do use indigo for clear hierarchy and readable text.
- Do use orange for HTML/code symbols and artifact highlights.
- Do use lock, key, shield, and dotted arc motifs subtly to communicate secure private sharing.
- Do make artifact-related UI feel like cards placed on a shelf.
- Do keep mascot details simple enough to survive at GitHub-avatar size.
- Do preserve generous whitespace in README hero images and onboarding screens.
- Do maintain accessibility contrast, especially for text on pastel surfaces.

### Don't

- Don't use harsh neon colors, cyberpunk gradients, pure black backgrounds, or loud saturated palettes.
- Don't turn the mascot into an anime character, animal mascot, realistic robot, or aggressive security character.
- Don't overcrowd screens with mascot props, dotted arcs, locks, and shelves all at once.
- Don't use sharp corners, spikes, jagged shapes, or hard technical brutalism.
- Don't use orange as the primary UI color.
- Don't place long text inside tiny 3D illustration assets.
- Don't rely on 3D illustrations for core product comprehension.
- Don't stretch, crop awkwardly, recolor, or distort the mascot.
- Don't mix multiple unrelated illustration styles in the same screen.
- Don't let cuteness reduce credibility. Pageshelf is an OSS developer tool first.

## Iconography

Icons should be simple, rounded, and object-like. Use filled or softly shaded icons for brand/marketing surfaces, and simple line/filled hybrid icons for product UI.

Core motifs:

- HTML Page: cream paper card, folded corner, orange `</>`.
- CLI Prompt: dark indigo rounded tile with light prompt mark.
- Browser Window: rounded mini window with header dots and abstract content blocks.
- Lock: small gold/cream lock for protected items.
- Key Tag: mint tag with lock mark for token/private access.
- Shield: soft blue shield for security status.
- Private Network Arc: dotted lavender/blue arc, never a copied third-party network logo.
- Home Cloud Badge: cream cloud with small blue home/cubby mark for local-first/private shelf.

## Illustration & Mascot Art Direction

Pageshelf illustrations should look like soft vinyl collectibles photographed in a clean studio, but simplified enough for product use.

Art direction:

- Smooth 3D forms with soft ambient shadows.
- Low-detail faces and props.
- Warm cream background or transparent export.
- Centered compositions with clean silhouette.
- No complex environments.
- No text except symbolic code marks such as `</>` or `>_`.
- Security and network elements are secondary details, not the main subject.

Preferred scene types:

- Mascot sitting beside the artifact shelf.
- Mascot holding an HTML page.
- Shelf filled with glowing document cards.
- Tiny CLI prompt tile near browser cards.
- Lock/key charm attached to mascot or shelf.
- Dotted arc around shelf for private sharing.

For README hero images, use a horizontal lockup: mascot-and-shelf symbol on the left, Pageshelf wordmark on the right, ultra-light background, high readability, and no clutter.

## Voice in UI Copy

The visual identity is calm and competent; copy should match.

- Prefer plain, direct wording.
- Use security language without fear-mongering.
- Explain local-first behavior concretely.
- Avoid mascot catchphrases, jokes, or overly cute copy in core product flows.
- Use playful warmth only in empty states, success moments, and README visuals.

Good examples:

- “Saved to your local shelf.”
- “Share with a private token.”
- “Artifact is only available from this server.”
- “Open local preview.”

Bad examples:

- “OMG your magical files are safe!!!”
- “Military-grade secure artifact vault.”
- “Hack the planet.”

## Implementation Notes for Agents

When generating UI for Pageshelf:

1. Start with a cream/off-white page surface.
2. Use indigo text, pastel cards, rounded corners, and 8px spacing.
3. Use document cards, browser cards, CLI tiles, and lock badges as recurring primitives.
4. Use the mascot only when it clarifies or warms the experience; do not put it everywhere.
5. Favor calm order over decorative density.
6. Preserve the “local-first private shelf” metaphor across screens.
7. If a token is missing, choose the closest existing token before inventing a new style.
8. When in doubt, make it simpler, softer, and more organized.
