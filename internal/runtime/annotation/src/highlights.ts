import { ID_PREFIX } from './constants';
import { cleanText } from './anchors';
import type { Annotation } from './types';

export function highlightAll(annotations: Annotation[], host: HTMLElement): void {
  document
    .querySelectorAll(`.${ID_PREFIX}highlight`)
    .forEach((span) => span.replaceWith(document.createTextNode(span.textContent || '')));

  annotations.forEach((annotation) => wrapQuote(annotation, host));
}

function wrapQuote(annotation: Annotation, host: HTMLElement): void {
  if (!annotation.quote || annotation.status === 'resolved') {
    return;
  }

  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      const parent = node.parentElement;

      if (!parent || host.contains(parent) || parent.closest('script,style,textarea,input')) {
        return NodeFilter.FILTER_REJECT;
      }

      return cleanText(node.nodeValue).includes(annotation.quote)
        ? NodeFilter.FILTER_ACCEPT
        : NodeFilter.FILTER_SKIP;
    },
  });

  const node = walker.nextNode();
  if (!node || !node.nodeValue) {
    return;
  }

  const index = node.nodeValue.indexOf(annotation.quote);
  if (index < 0) {
    return;
  }

  const range = document.createRange();
  range.setStart(node, index);
  range.setEnd(node, index + annotation.quote.length);

  const span = document.createElement('span');
  span.className = `${ID_PREFIX}highlight`;
  span.title = annotation.comment;

  try {
    range.surroundContents(span);
  } catch {
    // The prior runtime also ignored ranges that cannot be surrounded.
  }
}
