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
      Review <span class="count" id="count">0</span>
    </button>
    <section class="sheet" id="panel" aria-label="Pageshelf annotations" aria-hidden="true">
      <div class="head">
        <div>
          <div class="title">Review annotations</div>
          <div class="hint">Select text or tap Pick element, then add a note.</div>
        </div>
        <button class="btn" id="close" type="button" aria-label="Close">Close</button>
      </div>
      <textarea id="comment" placeholder="Requested change or approval note…"></textarea>
      <div class="row">
        <button class="btn primary" id="add" type="button">Add annotation</button>
        <button class="btn" id="pick" type="button">Pick element</button>
      </div>
      <div class="row">
        <button class="btn" id="copy" type="button">Copy annotations</button>
        <button class="btn" id="copyjson" type="button">Copy JSON</button>
      </div>
      <div class="row">
        <button class="btn" id="exportjson" type="button">Export JSON</button>
        <button class="btn" id="clear" type="button">Clear resolved</button>
      </div>
      <div class="toast" id="toast" role="status" aria-live="polite"></div>
      <div id="list" aria-live="polite"></div>
    </section>
  `;
}

export function renderAnnotationList(list: HTMLElement, annotations: Annotation[], actions: AnnotationActions): void {
  list.textContent = '';

  if (!annotations.length) {
    const empty = document.createElement('div');
    empty.className = 'empty';
    empty.textContent = 'No annotations yet.';
    list.appendChild(empty);
    return;
  }

  annotations.forEach((annotation) => {
    const item = document.createElement('article');
    item.className = `item ${annotation.status === 'resolved' ? 'resolved' : ''}`;
    item.innerHTML = `
      <div class="itemHead">
        <div class="meta"></div>
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

    requireElement<HTMLElement>(item, '.meta').textContent = annotation.heading || annotation.path || 'Document';
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
