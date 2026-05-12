import type { Annotation, AnnotationActions } from './types';

function requireElement<T extends Element>(parent: ParentNode, selector: string): T {
  const element = parent.querySelector<T>(selector);
  if (!element) {
    throw new Error(`Missing annotation template element: ${selector}`);
  }
  return element;
}

export function renderShell(root: ShadowRoot, styles: string): void {
  root.innerHTML = `
    <style>${styles}</style>
    <button class="btn fab" id="open" type="button" aria-expanded="false">
      <span class="fabLabel">Review notes</span><span class="count" id="count">0</span>
    </button>
    <section class="sheet" id="panel" aria-label="Pageshelf annotations" aria-hidden="true">
      <div class="handle" aria-hidden="true"></div>
      <div class="head">
        <div>
          <div class="eyebrow">Pageshelf</div>
          <div class="title">Review notes <span class="titleCount">(<span id="titlecount">0</span>)</span></div>
          <div class="hint">Select text in the page to add a contextual note. Use Pick element for layout or image feedback.</div>
        </div>
        <button class="iconBtn" id="close" type="button" aria-label="Close review notes">×</button>
      </div>
      <div class="drawerActions">
        <button class="btn" id="pick" type="button" aria-pressed="false">Pick element</button>
        <div class="overflowWrap">
          <button class="iconBtn overflowBtn" id="moretoggle" type="button" aria-label="More annotation actions" aria-haspopup="menu" aria-expanded="false" aria-controls="moremenu">⋯</button>
          <div class="overflowMenu" id="moremenu" role="menu" aria-label="Annotation utility actions" hidden>
            <button class="menuItem" id="copy" type="button" role="menuitem">Copy notes</button>
            <button class="menuItem" id="copyjson" type="button" role="menuitem">Copy as JSON</button>
            <button class="menuItem" id="exportjson" type="button" role="menuitem">Download JSON</button>
            <div class="menuDivider" role="separator"></div>
            <button class="menuItem danger" id="clear" type="button" role="menuitem">Clear resolved</button>
          </div>
        </div>
      </div>
      <div class="toast" id="toast" role="status" aria-live="polite"></div>
      <div id="list" aria-live="polite"></div>
    </section>
    <section class="composer" id="composer" aria-label="Add contextual review note" aria-hidden="true">
      <div class="composerTop">
        <div>
          <div class="eyebrow" id="composertitle">Selection note</div>
          <div class="composerContext" id="composercontext"></div>
        </div>
        <button class="iconBtn small" id="composercancel" type="button" aria-label="Cancel note">×</button>
      </div>
      <textarea id="composercomment" placeholder="What should change, or what is approved?"></textarea>
      <div class="row primaryRow">
        <button class="btn primary" id="composeradd" type="button">Add note</button>
        <button class="btn" id="composercancel2" type="button">Cancel</button>
      </div>
    </section>
  `;
}

export function renderAnnotationList(list: HTMLElement, annotations: Annotation[], actions: AnnotationActions): void {
  list.textContent = '';

  if (!annotations.length) {
    const empty = document.createElement('div');
    empty.className = 'empty';
    empty.innerHTML = '<strong>No notes yet.</strong><span>Select text in the page to open a note composer, or use Pick element for non-text feedback.</span>';
    list.appendChild(empty);
    return;
  }

  annotations.forEach((annotation) => {
    const item = document.createElement('article');
    item.className = `item ${annotation.status === 'resolved' ? 'resolved' : ''}`;
    item.innerHTML = `
      <div class="itemHead">
        <div class="meta"><span class="where"></span><span class="when"></span></div>
        <span class="status"></span>
      </div>
      <div class="quote"></div>
      <div class="comment"></div>
      <div class="itemActions">
        <button class="btn" data-act="edit">Edit</button>
        <button class="btn" data-act="resolve"></button>
        <button class="btn" data-act="delete">Delete</button>
      </div>
    `;

    requireElement<HTMLElement>(item, '.where').textContent = annotation.heading || annotation.path || 'Document';
    requireElement<HTMLElement>(item, '.when').textContent = new Date(annotation.updatedAt || annotation.createdAt).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
    requireElement<HTMLElement>(item, '.status').textContent = annotation.status || 'open';
    requireElement<HTMLElement>(item, '.quote').textContent = annotation.quote ? `“${annotation.quote}”` : 'Element note';
    requireElement<HTMLElement>(item, '.comment').textContent = annotation.comment;
    requireElement<HTMLButtonElement>(item, '[data-act=resolve]').textContent =
      annotation.status === 'resolved' ? 'Reopen' : 'Resolve';

    requireElement<HTMLButtonElement>(item, '[data-act=edit]').onclick = () => actions.edit(annotation);
    requireElement<HTMLButtonElement>(item, '[data-act=resolve]').onclick = () => actions.toggleResolved(annotation);
    requireElement<HTMLButtonElement>(item, '[data-act=delete]').onclick = () => actions.delete(annotation);

    list.appendChild(item);
  });
}
