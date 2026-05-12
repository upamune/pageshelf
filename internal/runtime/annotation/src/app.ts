import shadowStyles from './shadow.css?inline';
import { annotationStorageKey, createAnnotationId, HOST_ID, ANNOTATION_SCHEMA } from './constants';
import { elementAnchor, selectedAnchor } from './anchors';
import { annotationMarkdown, copyText, downloadJson, exportData } from './copy';
import { highlightAll } from './highlights';
import { loadAnnotations, saveAnnotations } from './storage';
import { renderAnnotationList, renderShell } from './template';
import type { Anchor, Annotation, AnnotationElements } from './types';

export function mountAnnotationApp() {
  if (document.getElementById(HOST_ID)) {
    return;
  }

  const host = document.createElement('div');
  host.id = HOST_ID;
  document.documentElement.appendChild(host);

  const root = host.attachShadow({ mode: 'open' });
  renderShell(root, shadowStyles);

  const storageKey = annotationStorageKey();
  const elements = getElements(root);
  let annotations: Annotation[] = loadAnnotations(storageKey);
  let pendingAnchor: Anchor | null = null;
  let pickMode = false;
  let selectionTimer = 0;
  let viewportFrame = 0;
  let isComposing = false;

  const isCoarsePointer = () => window.matchMedia('(pointer: coarse), (hover: none)').matches;

  function persist() {
    saveAnnotations(storageKey, annotations);
    render();
    highlightAll(annotations, host);
  }

  function addAnnotation() {
    const anchor = pendingAnchor;
    const comment = elements.composerComment.value.trim();

    if (!anchor) {
      closeComposer();
      return;
    }

    if (!comment) {
      toast('Add a comment first.');
      elements.composerComment.focus({ preventScroll: true });
      return;
    }

    annotations.unshift({
      schema: ANNOTATION_SCHEMA,
      id: createAnnotationId(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: 'open',
      comment,
      url: location.href,
      ...anchor,
    });

    closeComposer();
    getSelection()?.removeAllRanges();
    persist();
    toast('Note added.');
  }

  function openComposer(anchor: Anchor, rect?: DOMRect | null, title = 'Selection note') {
    window.clearTimeout(selectionTimer);
    pendingAnchor = anchor;
    elements.composerTitle.textContent = title;
    elements.composerContext.textContent = anchor.quote ? `“${anchor.quote.slice(0, 220)}${anchor.quote.length > 220 ? '…' : ''}”` : (anchor.heading || 'Element note');
    elements.composerComment.value = '';
    elements.composer.classList.add('open');
    elements.composer.setAttribute('aria-hidden', 'false');
    positionComposer(rect);
    if (!isCoarsePointer()) {
      requestAnimationFrame(() => elements.composerComment.focus({ preventScroll: true }));
    }
  }

  function closeComposer() {
    window.clearTimeout(selectionTimer);
    pendingAnchor = null;
    elements.composer.classList.remove('open');
    elements.composer.setAttribute('aria-hidden', 'true');
    elements.composerComment.value = '';
  }

  function positionComposer(rect?: DOMRect | null) {
    if (window.matchMedia('(max-width: 759px), (pointer: coarse), (hover: none)').matches) {
      elements.composer.style.left = '';
      elements.composer.style.top = '';
      elements.composer.style.transform = '';
      return;
    }

    if (!rect) {
      elements.composer.style.left = '50%';
      elements.composer.style.top = '72px';
      elements.composer.style.transform = 'translateX(-50%)';
      return;
    }
    const width = Math.min(360, window.innerWidth - 24);
    const left = Math.max(12, Math.min(rect.left + rect.width / 2 - width / 2, window.innerWidth - width - 12));
    const top = Math.max(12, Math.min(rect.bottom + 10, window.innerHeight - 260));
    elements.composer.style.left = `${left}px`;
    elements.composer.style.top = `${top}px`;
    elements.composer.style.transform = 'none';
  }

  function maybeOpenSelectionComposer() {
    if (pickMode) return;
    const selection = getSelection();
    if (!selection || selection.isCollapsed || !selection.toString().trim()) return;
    const anchor = selectedAnchor(host);
    if (!anchor) return;
    const rect = selection.rangeCount ? selection.getRangeAt(0).getBoundingClientRect() : null;
    openComposer(anchor, rect, 'Selection note');
  }

  function scheduleSelectionComposer() {
    if (isComposerTextActive()) return;
    window.clearTimeout(selectionTimer);
    const selection = getSelection();
    if (!selection || selection.isCollapsed || !selection.toString().trim()) {
      return;
    }

    // iOS/Safari emits noisy selectionchange events while its native selection
    // tooltip is visible. Touch/coarse-pointer devices get a longer settle delay
    // and the composer is rendered as the bottom mini-sheet by CSS/positioning.
    selectionTimer = window.setTimeout(maybeOpenSelectionComposer, isCoarsePointer() ? 520 : 160);
  }

  function isComposerTextActive() {
    return isComposing || root.activeElement === elements.composerComment || document.activeElement === elements.composerComment;
  }

  function render() {
    const openCount = annotations.filter((annotation) => annotation.status !== 'resolved').length;
    const resolvedCount = annotations.length - openCount;
    elements.count.textContent = String(openCount);
    elements.titleCount.textContent = String(openCount);
    elements.clear.disabled = resolvedCount === 0;
    elements.clear.setAttribute('aria-disabled', String(resolvedCount === 0));
    renderAnnotationList(elements.list, annotations, {
      edit(annotation) {
        const value = prompt('Edit annotation', annotation.comment);
        if (value !== null) {
          annotation.comment = value.trim();
          annotation.updatedAt = new Date().toISOString();
          persist();
        }
      },
      toggleResolved(annotation) {
        annotation.status = annotation.status === 'resolved' ? 'open' : 'resolved';
        annotation.updatedAt = new Date().toISOString();
        persist();
      },
      delete(annotation) {
        annotations = annotations.filter((candidate) => candidate.id !== annotation.id);
        persist();
      },
    });
  }

  function openPanel() {
    elements.panel.classList.add('open');
    elements.panel.setAttribute('aria-hidden', 'false');
    elements.open.setAttribute('aria-expanded', 'true');
  }

  function closePanel() {
    elements.panel.classList.remove('open');
    elements.panel.setAttribute('aria-hidden', 'true');
    elements.open.setAttribute('aria-expanded', 'false');
    closeOverflowMenu();
  }

  function setOverflowMenu(open: boolean) {
    elements.moreToggle.setAttribute('aria-expanded', String(open));
    elements.moreMenu.hidden = !open;
    elements.moreMenu.classList.toggle('open', open);
  }

  function closeOverflowMenu() {
    setOverflowMenu(false);
  }

  function runMenuAction(action: () => void) {
    closeOverflowMenu();
    action();
  }

  function toast(message: string) {
    elements.toast.textContent = message;
    setTimeout(() => {
      if (elements.toast.textContent === message) {
        elements.toast.textContent = '';
      }
    }, 2500);
  }

  elements.open.onclick = openPanel;
  elements.close.onclick = closePanel;
  elements.composerAdd.onclick = addAnnotation;
  elements.composerCancel.onclick = closeComposer;
  elements.composerCancelSecondary.onclick = closeComposer;
  elements.moreToggle.onclick = (event) => {
    event.stopPropagation();
    setOverflowMenu(elements.moreMenu.hidden === true);
  };
  elements.copy.onclick = () => runMenuAction(() => copyText(annotations.map(annotationMarkdown).join('\n\n') || 'No annotations.', () => toast('Notes copied.')));
  elements.copyJson.onclick = () => runMenuAction(() => copyText(JSON.stringify(exportData(annotations), null, 2), () => toast('JSON copied.')));
  elements.exportJson.onclick = () => runMenuAction(() => {
    downloadJson('pageshelf-annotations.json', exportData(annotations));
    toast('JSON download started.');
  });
  elements.clear.onclick = () => {
    if (elements.clear.disabled) return;
    closeOverflowMenu();
    const resolvedCount = annotations.filter((annotation) => annotation.status === 'resolved').length;
    annotations = annotations.filter((annotation) => annotation.status !== 'resolved');
    persist();
    toast(`${resolvedCount} resolved ${resolvedCount === 1 ? 'note' : 'notes'} cleared.`);
  };
  elements.pick.onclick = () => {
    pickMode = !pickMode;
    elements.pick.classList.toggle('active', pickMode);
    elements.pick.setAttribute('aria-pressed', String(pickMode));
    toast(pickMode ? 'Tap an element to annotate.' : 'Pick cancelled.');
  };

  root.addEventListener('pointerdown', (event) => {
    const path = event.composedPath();
    if (!elements.moreMenu.hidden && !path.includes(elements.moreMenu) && !path.includes(elements.moreToggle)) {
      closeOverflowMenu();
    }
  });

  document.addEventListener(
    'pointerdown',
    (event) => {
      const target = event.target;
      if (!event.composedPath().includes(host) && !elements.moreMenu.hidden) {
        closeOverflowMenu();
      }
      if (!pickMode || !(target instanceof Element) || host.contains(target)) {
        return;
      }

      event.preventDefault();
      const anchor = elementAnchor(target);
      target.classList.add('ps-ann-pick');
      setTimeout(() => target.classList.remove('ps-ann-pick'), 700);
      pickMode = false;
      elements.pick.classList.remove('active');
      elements.pick.setAttribute('aria-pressed', 'false');
      openComposer(anchor, target.getBoundingClientRect(), 'Element note');
    },
    { capture: true },
  );

  const updateViewportVars = () => {
    viewportFrame = 0;
    const visualViewport = window.visualViewport;
    const viewportHeight = visualViewport?.height || innerHeight;
    const bottomInset = Math.max(0, Math.round(innerHeight - viewportHeight - (visualViewport?.offsetTop || 0)));
    const hostStyle = (root.host as HTMLElement).style;
    hostStyle.setProperty('--ps-vh', `${viewportHeight}px`);
    hostStyle.setProperty('--ps-vv-bottom', `${bottomInset}px`);
    hostStyle.setProperty('--ps-keyboard-inset', `${bottomInset}px`);
  };
  const scheduleViewportVars = () => {
    if (viewportFrame) return;
    viewportFrame = requestAnimationFrame(updateViewportVars);
  };
  const scheduleFocusViewportVars = () => {
    scheduleViewportVars();
    window.setTimeout(scheduleViewportVars, 80);
    window.setTimeout(scheduleViewportVars, 260);
  };
  addEventListener('resize', scheduleViewportVars);
  addEventListener('orientationchange', scheduleViewportVars);
  addEventListener('focusin', scheduleFocusViewportVars);
  addEventListener('focusout', scheduleFocusViewportVars);
  window.visualViewport?.addEventListener('resize', scheduleViewportVars);
  window.visualViewport?.addEventListener('scroll', scheduleViewportVars);

  document.addEventListener('selectionchange', scheduleSelectionComposer);
  elements.composerComment.addEventListener('compositionstart', () => {
    isComposing = true;
  });
  elements.composerComment.addEventListener('compositionend', () => {
    isComposing = false;
  });
  elements.composerComment.addEventListener('keydown', (event) => {
    if (event.isComposing || isComposing) return;
    if (event.key === 'Escape') {
      event.stopPropagation();
      closeComposer();
      return;
    }
    if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      addAnnotation();
    }
  });
  document.addEventListener('keydown', (event) => {
    if (event.isComposing || isComposing) return;
    if (event.key === 'Escape') {
      if (!elements.moreMenu.hidden) {
        closeOverflowMenu();
        elements.moreToggle.focus({ preventScroll: true });
      }
      closeComposer();
      if (pickMode) {
        pickMode = false;
        elements.pick.classList.remove('active');
        elements.pick.setAttribute('aria-pressed', 'false');
      }
    }
  });

  updateViewportVars();
  render();
  highlightAll(annotations, host);
}

function getElements(root: ShadowRoot): AnnotationElements {
  const query = <T extends Element>(selector: string): T => {
    const element = root.querySelector<T>(selector);
    if (!element) {
      throw new Error(`Missing annotation runtime element: ${selector}`);
    }
    return element;
  };

  return {
    open: query<HTMLButtonElement>('#open'),
    close: query<HTMLButtonElement>('#close'),
    panel: query<HTMLElement>('#panel'),
    composer: query<HTMLElement>('#composer'),
    composerTitle: query<HTMLElement>('#composertitle'),
    composerContext: query<HTMLElement>('#composercontext'),
    composerComment: query<HTMLTextAreaElement>('#composercomment'),
    list: query<HTMLElement>('#list'),
    count: query<HTMLElement>('#count'),
    titleCount: query<HTMLElement>('#titlecount'),
    toast: query<HTMLElement>('#toast'),
    composerAdd: query<HTMLButtonElement>('#composeradd'),
    composerCancel: query<HTMLButtonElement>('#composercancel'),
    composerCancelSecondary: query<HTMLButtonElement>('#composercancel2'),
    pick: query<HTMLButtonElement>('#pick'),
    moreToggle: query<HTMLButtonElement>('#moretoggle'),
    moreMenu: query<HTMLElement>('#moremenu'),
    copy: query<HTMLButtonElement>('#copy'),
    copyJson: query<HTMLButtonElement>('#copyjson'),
    exportJson: query<HTMLButtonElement>('#exportjson'),
    clear: query<HTMLButtonElement>('#clear'),
  };
}
