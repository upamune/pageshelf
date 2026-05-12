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
  let pickedAnchor: Anchor | null = null;
  let pickMode = false;

  function persist() {
    saveAnnotations(storageKey, annotations);
    render();
    highlightAll(annotations, host);
  }

  function addAnnotation() {
    const anchor = pickedAnchor || selectedAnchor(host) || { heading: document.title || 'Document', path: '', quote: '' };
    const comment = elements.comment.value.trim();

    if (!comment) {
      toast('Add a comment first.');
      elements.comment.focus();
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

    elements.comment.value = '';
    pickedAnchor = null;
    persist();
    openPanel();
    toast('Annotation added.');
  }

  function render() {
    elements.count.textContent = String(annotations.filter((annotation) => annotation.status !== 'resolved').length);
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
  elements.add.onclick = addAnnotation;
  elements.copy.onclick = () => copyText(annotations.map(annotationMarkdown).join('\n\n') || 'No annotations.', () => toast('Copied.'));
  elements.copyJson.onclick = () => copyText(JSON.stringify(exportData(annotations), null, 2), () => toast('Copied.'));
  elements.exportJson.onclick = () => downloadJson('pageshelf-annotations.json', exportData(annotations));
  elements.clear.onclick = () => {
    annotations = annotations.filter((annotation) => annotation.status !== 'resolved');
    persist();
  };
  elements.pick.onclick = () => {
    pickMode = !pickMode;
    toast(pickMode ? 'Tap an element to annotate.' : 'Pick cancelled.');
  };

  document.addEventListener(
    'pointerdown',
    (event) => {
      const target = event.target;
      if (!pickMode || !(target instanceof Element) || host.contains(target)) {
        return;
      }

      event.preventDefault();
      pickedAnchor = elementAnchor(target);
      target.classList.add('ps-ann-pick');
      setTimeout(() => target.classList.remove('ps-ann-pick'), 700);
      pickMode = false;
      openPanel();
      elements.comment.focus();
    },
    { capture: true },
  );

  const setViewportHeight = () => {
    (root.host as HTMLElement).style.setProperty('--ps-vh', `${(window.visualViewport && window.visualViewport.height) || innerHeight}px`);
  };
  addEventListener('resize', setViewportHeight);
  window.visualViewport?.addEventListener('resize', setViewportHeight);

  setViewportHeight();
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
    comment: query<HTMLTextAreaElement>('#comment'),
    list: query<HTMLElement>('#list'),
    count: query<HTMLElement>('#count'),
    toast: query<HTMLElement>('#toast'),
    add: query<HTMLButtonElement>('#add'),
    pick: query<HTMLButtonElement>('#pick'),
    copy: query<HTMLButtonElement>('#copy'),
    copyJson: query<HTMLButtonElement>('#copyjson'),
    exportJson: query<HTMLButtonElement>('#exportjson'),
    clear: query<HTMLButtonElement>('#clear'),
  };
}
