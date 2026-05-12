import type { Anchor } from './types';

export function cleanText(value: string | null | undefined): string {
  return (value || '').replace(/\s+/g, ' ').trim();
}

function clipContext(value: string): string {
  return value.length > 180 ? value.slice(0, 180) : value;
}

function elementPath(element: Element): string {
  const parts: string[] = [];

  for (
    let node: Element | null = element;
    node && node.nodeType === Node.ELEMENT_NODE && node !== document.body;
    node = node.parentElement
  ) {
    let index = 1;
    let previous = node.previousElementSibling;

    while (previous) {
      if (previous.tagName === node.tagName) {
        index += 1;
      }
      previous = previous.previousElementSibling;
    }

    parts.unshift(`${node.tagName.toLowerCase()}:nth-of-type(${index})`);
  }

  return parts.join('>');
}

function headingFor(element: Element): string {
  const precedingHeadings = [...document.querySelectorAll('h1,h2,h3,h4,h5,h6')].filter(
    (heading) => heading.compareDocumentPosition(element) & Node.DOCUMENT_POSITION_FOLLOWING,
  );

  return precedingHeadings.length ? cleanText(precedingHeadings[precedingHeadings.length - 1]?.textContent) : '';
}

function rangeContext(range: Range): Pick<Anchor, 'prefix' | 'suffix'> {
  const containerText = cleanText(range.startContainer.textContent || '');
  const quote = cleanText(range.toString());
  const index = containerText.indexOf(quote);

  return {
    prefix: clipContext(index > 0 ? containerText.slice(Math.max(0, index - 80), index) : ''),
    suffix: clipContext(index >= 0 ? containerText.slice(index + quote.length, index + quote.length + 80) : ''),
  };
}

export function selectedAnchor(host: HTMLElement): Anchor | null {
  const selection = getSelection();

  if (!selection || selection.isCollapsed) {
    return null;
  }

  const range = selection.getRangeAt(0);
  const element =
    range.commonAncestorContainer.nodeType === Node.ELEMENT_NODE
      ? (range.commonAncestorContainer as Element)
      : range.commonAncestorContainer.parentElement;
  const quote = cleanText(selection.toString()).slice(0, 1000);

  if (!quote || !element || host.contains(element)) {
    return null;
  }

  return {
    quote,
    ...rangeContext(range),
    heading: headingFor(element),
    path: elementPath(element),
  };
}

export function elementAnchor(element: Element): Anchor {
  return {
    quote: cleanText(element.textContent).slice(0, 500),
    heading: headingFor(element),
    path: elementPath(element),
  };
}
